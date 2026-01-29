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
	mapData [][]int
}

func NewWorld() *World {
	return &World{
		players: make(map[string]*Player),
		mapData: buildMapData(mapWidth, mapHeight),
	}
}

func (w *World) AddPlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	spawnX := w.tileCenter(mapWidth / 2)
	spawnY := w.tileCenter(mapHeight / 2)

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

func (w *World) SetPlayerTarget(id string, x, y int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	player, ok := w.players[id]
	if !ok {
		return false
	}

	targetTileX, targetTileY := w.toTileCoords(x, y)
	if !w.isWalkable(targetTileX, targetTileY) {
		return false
	}

	currentTileX, currentTileY := w.toTileCoords(player.X, player.Y)
	if absInt(targetTileX-currentTileX)+absInt(targetTileY-currentTileY) > 1 {
		return false
	}

	player.TargetX = w.tileCenter(targetTileX)
	player.TargetY = w.tileCenter(targetTileY)
	player.HasTarget = !(player.TargetX == player.X && player.TargetY == player.Y)

	return true
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
const tileSize = 32
const mapWidth = 50
const mapHeight = 50
const tileWorldSize = tileSize * positionScale

func (w *World) toTileCoords(x, y int) (int, int) {
	return x / tileWorldSize, y / tileWorldSize
}

func (w *World) tileCenter(tile int) int {
	return tile*tileWorldSize + tileWorldSize/2
}

func (w *World) isWalkable(x, y int) bool {
	if x < 0 || y < 0 || x >= mapWidth || y >= mapHeight {
		return false
	}

	return w.mapData[y][x] != 2
}

func buildMapData(width, height int) [][]int {
	data := make([][]int, 0, height)

	for y := 0; y < height; y += 1 {
		row := make([]int, 0, width)
		for x := 0; x < width; x += 1 {
			isBorder := x == 0 || y == 0 || x == width-1 || y == height-1
			tileIndex := 0

			if isBorder {
				tileIndex = 2
			} else if (x+y)%7 == 0 {
				tileIndex = 1
			}

			row = append(row, tileIndex)
		}
		data = append(data, row)
	}

	return data
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
