package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/felipemalacarne/etheria/internal/game/engine"
	"github.com/felipemalacarne/etheria/internal/network/websocket"
)

const (
	defaultPort       = "8080"
	defaultTickMs     = 50
	defaultMapPath    = "shared/maps/basic.json"
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	addr := ":" + getenv("PORT", defaultPort)
	tickRate := time.Duration(getenvInt("TICK_MS", defaultTickMs)) * time.Millisecond

	mapPath := getenv("MAP_PATH", defaultMapPath)
	mapData, err := engine.LoadMapData(mapPath)
	if err != nil {
		log.Printf("map load failed (%s): %v", mapPath, err)
		mapData = engine.DefaultMapData(engine.DefaultMapWidth, engine.DefaultMapHeight)
	}

	world := engine.NewWorld(mapData)
	server := websocket.NewServer(world)
	loop := engine.NewLoop(tickRate, func(tick int64, delta time.Duration) {
		world.Step(delta.Seconds())
		if world.DrainDirty() {
			server.BroadcastState(tick)
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.HandleWS)
	mux.HandleFunc("/players", handlePlayers(world))
	mux.HandleFunc("/map", handleMap(mapData))

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go loop.Start(ctx)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("game server listening on %s (tick %s)", addr, tickRate)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}

	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	server.Close()
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getenvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

type playerInfo struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type playerList struct {
	Players []playerInfo `json:"players"`
}

func handlePlayers(world *engine.World) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-store")

		players := world.SnapshotPlayers()
		response := playerList{Players: make([]playerInfo, 0, len(players))}
		for _, player := range players {
			response.Players = append(response.Players, playerInfo{
				ID: player.ID,
				X:  float64(player.X) / engine.PositionScale,
				Y:  float64(player.Y) / engine.PositionScale,
			})
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

func handleMap(mapData engine.MapData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-store")

		if err := json.NewEncoder(w).Encode(mapData); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
