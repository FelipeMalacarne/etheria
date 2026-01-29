package engine

import (
	"fmt"
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
	Path      []tilePoint
	PathIndex int
}

type World struct {
	mu      sync.RWMutex
	players map[string]*Player
	mapData [][]int
	dirty   bool
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
		Path:      nil,
		PathIndex: 0,
	}

	w.dirty = true
}

func (w *World) RemovePlayer(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.players, id)

	w.dirty = true
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

	startTileX, startTileY := w.toTileCoords(player.X, player.Y)
	path := w.findPath(tilePoint{X: startTileX, Y: startTileY}, tilePoint{X: targetTileX, Y: targetTileY})
	if len(path) == 0 {
		return false
	}

	player.Path = path
	player.PathIndex = 1
	if len(path) <= 1 {
		player.TargetX = player.X
		player.TargetY = player.Y
		player.HasTarget = false
		player.Path = nil
		player.PathIndex = 0
		return true
	}

	next := path[player.PathIndex]
	player.TargetX = w.tileCenter(next.X)
	player.TargetY = w.tileCenter(next.Y)
	player.HasTarget = true

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
		if player.PathIndex >= len(player.Path) {
			player.Path = nil
			player.PathIndex = 0
			player.HasTarget = false
			continue
		}

		if !player.HasTarget {
			next := player.Path[player.PathIndex]
			player.TargetX = w.tileCenter(next.X)
			player.TargetY = w.tileCenter(next.Y)
			player.HasTarget = true
		}

		dx := float64(player.TargetX - player.X)
		dy := float64(player.TargetY - player.Y)
		distance := math.Hypot(dx, dy)
		if distance == 0 {
			player.PathIndex += 1
			player.HasTarget = false
			if player.PathIndex >= len(player.Path) {
				player.Path = nil
				player.PathIndex = 0
			}
			continue
		}

		step := speedScaled * deltaSeconds
		if distance <= step {
			player.X = player.TargetX
			player.Y = player.TargetY
			player.PathIndex += 1
			player.HasTarget = false
			if player.PathIndex >= len(player.Path) {
				player.Path = nil
				player.PathIndex = 0
			}
			w.dirty = true
			continue
		}

		ratio := step / distance
		player.X += int(math.Round(dx * ratio))
		player.Y += int(math.Round(dy * ratio))
		w.dirty = true
	}
}

func (w *World) DrainDirty() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.dirty {
		return false
	}

	w.dirty = false
	return true
}

func (w *World) SnapshotPlayersInChunkRadius(id string, chunkRadius int, chunkSizeTiles int) ([]Player, bool) {
	if chunkSizeTiles <= 0 {
		chunkSizeTiles = 1
	}
	if chunkRadius < 0 {
		chunkRadius = 0
	}

	w.mu.RLock()
	player, ok := w.players[id]
	if !ok {
		w.mu.RUnlock()
		return nil, false
	}

	centerTileX, centerTileY := w.toTileCoords(player.X, player.Y)
	centerChunkX := centerTileX / chunkSizeTiles
	centerChunkY := centerTileY / chunkSizeTiles

	players := make([]Player, 0, len(w.players))
	for _, other := range w.players {
		tileX, tileY := w.toTileCoords(other.X, other.Y)
		chunkX := tileX / chunkSizeTiles
		chunkY := tileY / chunkSizeTiles
		if absInt(chunkX-centerChunkX) <= chunkRadius && absInt(chunkY-centerChunkY) <= chunkRadius {
			players = append(players, *other)
		}
	}
	w.mu.RUnlock()

	return players, true
}

const positionScale = 100
const tileSize = 32
const mapWidth = 50
const mapHeight = 50
const tileWorldSize = tileSize * positionScale

type tilePoint struct {
	X int
	Y int
}

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

func (w *World) findPath(start, goal tilePoint) []tilePoint {
	if !w.isWalkable(goal.X, goal.Y) {
		return nil
	}

	startKey := keyFor(start.X, start.Y)
	goalKey := keyFor(goal.X, goal.Y)

	open := []pathNode{{
		point: start,
		f:     heuristic(start, goal),
	}}
	openSet := map[string]struct{}{startKey: {}}
	cameFrom := map[string]string{}
	gScore := map[string]int{startKey: 0}

	for len(open) > 0 {
		currentIndex := 0
		for i := 1; i < len(open); i += 1 {
			if open[i].f < open[currentIndex].f {
				currentIndex = i
			}
		}

		current := open[currentIndex]
		open = append(open[:currentIndex], open[currentIndex+1:]...)
		currentKey := keyFor(current.point.X, current.point.Y)
		delete(openSet, currentKey)

		if currentKey == goalKey {
			return reconstructPath(cameFrom, goalKey)
		}

		neighbors := []tilePoint{
			{X: current.point.X + 1, Y: current.point.Y},
			{X: current.point.X - 1, Y: current.point.Y},
			{X: current.point.X, Y: current.point.Y + 1},
			{X: current.point.X, Y: current.point.Y - 1},
		}

		for _, neighbor := range neighbors {
			if !w.isWalkable(neighbor.X, neighbor.Y) {
				continue
			}

			neighborKey := keyFor(neighbor.X, neighbor.Y)
			tentativeG := gScore[currentKey] + 1
			if existingG, ok := gScore[neighborKey]; ok && tentativeG >= existingG {
				continue
			}

			cameFrom[neighborKey] = currentKey
			gScore[neighborKey] = tentativeG
			fScore := tentativeG + heuristic(neighbor, goal)

			if _, ok := openSet[neighborKey]; !ok {
				open = append(open, pathNode{point: neighbor, f: fScore})
				openSet[neighborKey] = struct{}{}
			}
		}
	}

	return nil
}

type pathNode struct {
	point tilePoint
	f     int
}

func keyFor(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func heuristic(a, b tilePoint) int {
	return absInt(a.X-b.X) + absInt(a.Y-b.Y)
}

func reconstructPath(cameFrom map[string]string, goalKey string) []tilePoint {
	var path []tilePoint
	currentKey := goalKey

	for {
		x, y := parseKey(currentKey)
		path = append(path, tilePoint{X: x, Y: y})
		prev, ok := cameFrom[currentKey]
		if !ok {
			break
		}
		currentKey = prev
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func parseKey(key string) (int, int) {
	var x, y int
	fmt.Sscanf(key, "%d,%d", &x, &y)
	return x, y
}
