package engine

import (
	"math"
	"sync"
)

type Player struct {
	ID        string
	X         int
	Y         int
	TargetX   int
	TargetY   int
	HasTarget bool
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

	spawnX := 800 * positionScale
	spawnY := 800 * positionScale

	w.players[id] = &Player{
		ID:        id,
		X:         spawnX,
		Y:         spawnY,
		TargetX:   spawnX,
		TargetY:   spawnY,
		HasTarget: false,
	}
}

func (w *World) RemovePlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.players, id)
}

func (w *World) SetPlayerTarget(id string, x, y int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	player, ok := w.players[id]
	if !ok {
		return
	}

	player.TargetX = x
	player.TargetY = y
	player.HasTarget = true
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

func (w *World) Step(deltaSeconds float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if deltaSeconds <= 0 {
		return
	}

	const speed = 140.0
	const speedScaled = speed * positionScale

	for _, player := range w.players {
		if !player.HasTarget {
			continue
		}

		dx := float64(player.TargetX - player.X)
		dy := float64(player.TargetY - player.Y)
		distance := math.Hypot(dx, dy)
		if distance == 0 {
			player.HasTarget = false
			continue
		}

		step := speedScaled * deltaSeconds
		if distance <= step {
			player.X = player.TargetX
			player.Y = player.TargetY
			player.HasTarget = false
			continue
		}

		ratio := step / distance
		player.X += int(math.Round(dx * ratio))
		player.Y += int(math.Round(dy * ratio))
	}
}

const positionScale = 100
