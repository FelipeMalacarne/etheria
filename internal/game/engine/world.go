package engine

import "sync"

type Player struct {
	ID string
	X  int
	Y  int
}

type World struct {
	mu      sync.RWMutex
	players map[string]*Player
}

func NewWorld() *World {
	return &World{
		players: make(map[string]*Player),
	}
}

func (w *World) AddPlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.players[id] = &Player{
		ID: id,
		X:  0,
		Y:  0,
	}
}

func (w *World) RemovePlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.players, id)
}

func (w *World) SetPlayerPosition(id string, x, y int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	player, ok := w.players[id]
	if !ok {
		return
	}

	player.X = x
	player.Y = y
}

func (w *World) SnapshotPlayers() []Player {
	w.mu.RLock()
	defer w.mu.RUnlock()

	players := make([]Player, 0, len(w.players))
	for _, player := range w.players {
		players = append(players, *player)
	}

	return players
}
