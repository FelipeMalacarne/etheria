package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	appauth "github.com/felipemalacarne/etheria/internal/app/auth"
	"github.com/felipemalacarne/etheria/internal/app/auth/password"
	"github.com/felipemalacarne/etheria/internal/domain/account"
	"github.com/felipemalacarne/etheria/internal/game/engine"
	"github.com/felipemalacarne/etheria/internal/infrastructure/id"
	filerepo "github.com/felipemalacarne/etheria/internal/infrastructure/repositories/file"
	"github.com/felipemalacarne/etheria/internal/infrastructure/session"
	"github.com/felipemalacarne/etheria/internal/network/websocket"
)

const (
	defaultPort       = "8080"
	defaultTickMs     = 50
	defaultMapPath    = "shared/maps/basic.json"
	defaultUserDBPath = "shared/data/users.json"
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

	userRepo, err := filerepo.NewUserRepository(getenv("USER_DB_PATH", defaultUserDBPath))
	if err != nil {
		log.Fatalf("failed to load user store: %v", err)
	}
	authService := appauth.NewService(
		userRepo,
		password.NewBcryptHasher(),
		session.NewMemoryStore(),
		id.NewUUIDGenerator(),
	)

	world := engine.NewWorld(mapData)
	server := websocket.NewServer(world, authService)
	loop := engine.NewLoop(tickRate, func(tick int64, delta time.Duration) {
		world.Step(delta.Seconds())
		if world.DrainDirty() {
			server.BroadcastState(tick)
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.HandleWS)
	mux.HandleFunc("/players", withCORS(handlePlayers(world, authService)))
	mux.HandleFunc("/map", withCORS(handleMap(mapData)))
	mux.HandleFunc("/auth/login", withCORS(handleAuthLogin(authService)))
	mux.HandleFunc("/auth/register", withCORS(handleAuthRegister(authService)))

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

func handlePlayers(world *engine.World, authService *appauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		token := r.URL.Query().Get("token")
		user, ok, err := authService.AuthenticateToken(r.Context(), token)
		if err != nil {
			log.Printf("auth error: %v", err)
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		players, ok := world.SnapshotPlayersInChunkRadius(user.ID, websocket.ChunkRadius, websocket.ChunkSizeTiles)
		if !ok {
			http.Error(w, "player not found", http.StatusNotFound)
			return
		}

		response := playerList{Players: make([]playerInfo, 0, len(players))}
		for _, player := range players {
			response.Players = append(response.Players, playerInfo{
				ID: player.ID,
				X:  float64(player.X) / engine.PositionScale,
				Y:  float64(player.Y) / engine.PositionScale,
			})
		}

		writeJSON(w, response)
	}
}

func handleMap(mapData engine.MapData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writeJSON(w, mapData)
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string             `json:"token"`
	User  account.PublicUser `json:"user"`
}

func handleAuthLogin(manager *appauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		user, token, err := manager.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			status := http.StatusInternalServerError
			message := "server error"
			if errors.Is(err, account.ErrInvalidCredentials) {
				status = http.StatusUnauthorized
				message = "invalid credentials"
			}
			http.Error(w, message, status)
			return
		}

		writeJSON(w, authResponse{Token: token, User: user})
	}
}

func handleAuthRegister(manager *appauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		user, token, err := manager.Register(r.Context(), req.Email, req.Username, req.Password)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, account.ErrEmailExists) {
				status = http.StatusConflict
			} else if errors.Is(err, account.ErrInvalidInput) {
				status = http.StatusBadRequest
			} else {
				status = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), status)
			return
		}

		writeJSON(w, authResponse{Token: token, User: user})
	}
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Cache-Control", "no-store")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
