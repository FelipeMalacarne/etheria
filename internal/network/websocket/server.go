package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/felipemalacarne/etheria/internal/game/engine"
	"github.com/felipemalacarne/etheria/internal/network/packets"
)

const (
	writeTimeout = 5 * time.Second
	sendBuffer   = 16
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	world   *engine.World
	clients map[*client]struct{}
	nextID  uint64
	mu      sync.RWMutex
}

type client struct {
	id        string
	conn      *websocket.Conn
	send      chan packets.Packet
	closeOnce sync.Once
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

func NewServer(world *engine.World) *Server {
	return &Server{
		world:   world,
		clients: make(map[*client]struct{}),
	}
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := s.newClient(conn)
	s.addClient(client)

	go s.writeLoop(client)
	go s.readLoop(client)
}

func (s *Server) BroadcastState(tick int64) {
	players := s.world.SnapshotPlayers()
	statePlayers := make([]packets.PlayerState, 0, len(players))
	for _, player := range players {
		statePlayers = append(statePlayers, packets.PlayerState{
			ID: player.ID,
			X:  player.X,
			Y:  player.Y,
		})
	}

	packet, err := packets.NewPacket(packets.PacketStateUpdate, packets.StateUpdate{
		Tick:    tick,
		Players: statePlayers,
	})
	if err != nil {
		log.Printf("state update encode failed: %v", err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for client := range s.clients {
		select {
		case client.send <- packet:
		default:
		}
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

func (s *Server) newClient(conn *websocket.Conn) *client {
	id := atomic.AddUint64(&s.nextID, 1)

	return &client{
		id:   strconv.FormatUint(id, 10),
		conn: conn,
		send: make(chan packets.Packet, sendBuffer),
	}
}

func (s *Server) addClient(client *client) {
	s.mu.Lock()
	s.clients[client] = struct{}{}
	s.mu.Unlock()

	s.world.AddPlayer(client.id)
	s.sendPacket(client, packets.PacketWelcome, packets.Welcome{ID: client.id})
}

func (s *Server) removeClient(client *client) {
	s.mu.Lock()
	if _, ok := s.clients[client]; !ok {
		s.mu.Unlock()
		return
	}
	delete(s.clients, client)
	s.mu.Unlock()

	s.world.RemovePlayer(client.id)
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
			log.Printf("read error (%s): %v", client.id, err)
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
			log.Printf("write error (%s): %v", client.id, err)
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
			log.Printf("invalid move intent (%s): %v", client.id, err)
			return
		}
		s.world.SetPlayerPosition(client.id, intent.X, intent.Y)
	default:
	}
}

func (s *Server) sendPacket(client *client, packetType string, payload any) {
	packet, err := packets.NewPacket(packetType, payload)
	if err != nil {
		log.Printf("packet encode failed (%s): %v", client.id, err)
		return
	}

	select {
	case client.send <- packet:
	default:
	}
}
