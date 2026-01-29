package main

import (
	"context"
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
	defaultTickMs     = 200
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	addr := ":" + getenv("PORT", defaultPort)
	tickRate := time.Duration(getenvInt("TICK_MS", defaultTickMs)) * time.Millisecond

	world := engine.NewWorld()
	server := websocket.NewServer(world)
	loop := engine.NewLoop(tickRate, func(tick int64, delta time.Duration) {
		world.Step(delta.Seconds())
		server.BroadcastState(tick)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.HandleWS)

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
