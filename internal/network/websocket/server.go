package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	appauth "github.com/felipemalacarne/etheria/internal/app/auth"
	"github.com/felipemalacarne/etheria/internal/game/engine"
	"github.com/felipemalacarne/etheria/internal/network/packets"
)

const (
	writeTimeout   = 5 * time.Second
	sendBuffer     = 16
	ChunkSizeTiles = 8
	ChunkRadius    = 1
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	world         *engine.World
	auth          *appauth.Service
	clients       map[*client]struct{}
	clientsByUser map[string]*client
	mu            sync.RWMutex
	lastTick      int64
}

type client struct {
	userID    string
	conn      *websocket.Conn
	send      chan packets.Packet
	closeOnce sync.Once
	mu        sync.Mutex
	lastSent  map[string]packets.PlayerState
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

func NewServer(world *engine.World, authManager *appauth.Service) *Server {
	return &Server{
		world:         world,
		auth:          authManager,
		clients:       make(map[*client]struct{}),
		clientsByUser: make(map[string]*client),
	}
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	user, ok, err := s.auth.AuthenticateToken(r.Context(), token)
	if err != nil {
		log.Printf("auth token error: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := s.newClient(conn, user.ID)
	s.addClient(client)

	go s.writeLoop(client)
	go s.readLoop(client)
}

func (s *Server) BroadcastState(tick int64) {
	atomic.StoreInt64(&s.lastTick, tick)

	s.mu.RLock()
	clients := make([]*client, 0, len(s.clients))
	for client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.RUnlock()

	for _, client := range clients {
		s.sendDelta(client, tick)
	}
}

func (s *Server) Close() {
	s.mu.RLock()
	clients := make([]*client, 0, len(s.clients))
	for client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.RUnlock()

	for _, client := range clients {
		s.removeClient(client)
	}
}

func (s *Server) newClient(conn *websocket.Conn, userID string) *client {
	return &client{
		userID:   userID,
		conn:     conn,
		send:     make(chan packets.Packet, sendBuffer),
		lastSent: make(map[string]packets.PlayerState),
	}
}

func (s *Server) addClient(client *client) {
	s.mu.Lock()
	if existing, ok := s.clientsByUser[client.userID]; ok {
		delete(s.clients, existing)
		delete(s.clientsByUser, client.userID)
		s.mu.Unlock()
		s.world.RemovePlayer(existing.userID)
		existing.close()
		s.mu.Lock()
	}
	s.clients[client] = struct{}{}
	s.clientsByUser[client.userID] = client
	s.mu.Unlock()

	s.world.AddPlayer(client.userID)
	s.sendPacket(client, packets.PacketWelcome, packets.Welcome{ID: client.userID})
	s.sendSnapshot(client)
}

func (s *Server) removeClient(client *client) {
	s.mu.Lock()
	if _, ok := s.clients[client]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.clients, client)
	if current, ok := s.clientsByUser[client.userID]; ok && current == client {
		delete(s.clientsByUser, client.userID)
	}
	s.mu.Unlock()

	s.world.RemovePlayer(client.userID)
	client.close()
}

func (s *Server) readLoop(client *client) {
	defer s.removeClient(client)

	for {
		var packet packets.Packet
		if err := client.conn.ReadJSON(&packet); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			log.Printf("read error (%s): %v", client.userID, err)
			return
		}

		s.handlePacket(client, packet)
	}
}

func (s *Server) writeLoop(client *client) {
	for packet := range client.send {
		if err := client.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			s.removeClient(client)
			return
		}

		if err := client.conn.WriteJSON(packet); err != nil {
			log.Printf("write error (%s): %v", client.userID, err)
			s.removeClient(client)
			return
		}
	}
}

func (s *Server) handlePacket(client *client, packet packets.Packet) {
	switch packet.Type {
	case packets.PacketMoveIntent:
		var intent packets.MoveIntent
		if err := json.Unmarshal(packet.Payload, &intent); err != nil {
			log.Printf("invalid move intent (%s): %v", client.userID, err)
			return
		}
		if ok := s.world.SetPlayerTarget(client.userID, intent.X, intent.Y); !ok {
			return
		}
	default:
	}
}

func (s *Server) sendPacket(client *client, packetType string, payload any) {
	packet, err := packets.NewPacket(packetType, payload)
	if err != nil {
		log.Printf("packet encode failed (%s): %v", client.userID, err)
		return
	}

	select {
	case client.send <- packet:
	default:
	}
}

func (s *Server) sendSnapshot(client *client) {
	tick := atomic.LoadInt64(&s.lastTick)
	players, ok := s.world.SnapshotPlayersInChunkRadius(client.userID, ChunkRadius, ChunkSizeTiles)
	if !ok {
		return
	}

	statePlayers := make([]packets.PlayerState, 0, len(players))
	nextSent := make(map[string]packets.PlayerState, len(players))

	for _, player := range players {
		state := packets.PlayerState{
			ID: player.ID,
			X:  player.X,
			Y:  player.Y,
		}
		statePlayers = append(statePlayers, state)
		nextSent[player.ID] = state
	}

	client.mu.Lock()
	client.lastSent = nextSent
	client.mu.Unlock()

	s.sendPacket(client, packets.PacketStateSnapshot, packets.StateSnapshot{
		Tick:    tick,
		Players: statePlayers,
	})
}

func (s *Server) sendDelta(client *client, tick int64) {
	players, ok := s.world.SnapshotPlayersInChunkRadius(client.userID, ChunkRadius, ChunkSizeTiles)
	if !ok {
		return
	}

	statePlayers := make([]packets.PlayerState, 0, len(players))
	nextSent := make(map[string]packets.PlayerState, len(players))

	client.mu.Lock()
	prevSent := client.lastSent
	for _, player := range players {
		state := packets.PlayerState{
			ID: player.ID,
			X:  player.X,
			Y:  player.Y,
		}
		nextSent[player.ID] = state

		prev, ok := prevSent[player.ID]
		if !ok || prev.X != state.X || prev.Y != state.Y {
			statePlayers = append(statePlayers, state)
		}
	}

	removed := make([]string, 0)
	for id := range prevSent {
		if _, ok := nextSent[id]; !ok {
			removed = append(removed, id)
		}
	}

	if len(statePlayers) == 0 && len(removed) == 0 {
		client.mu.Unlock()
		return
	}

	client.lastSent = nextSent
	client.mu.Unlock()

	s.sendPacket(client, packets.PacketStateDelta, packets.StateDelta{
		Tick:    tick,
		Players: statePlayers,
		Removed: removed,
	})
}
