package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func normalizeOrigin(o string) string {
	o = strings.TrimSpace(o)
	if !strings.HasPrefix(o, "http://") && !strings.HasPrefix(o, "https://") {
		o = "http://" + o
	}
	return strings.ToLower(o)
}

type API struct {
	logger *zap.Logger
	config *Config
	rooms  *RoomManager
}

func NewAPI(config *Config, logger *zap.Logger, rm *RoomManager) *API {
	RoomManagerInstance = rm
	return &API{
		config: config,
		logger: logger,
		rooms:  rm,
	}
}

func (api *API) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	c := cors.New(cors.Options{
		AllowedOrigins: strings.Split(api.config.ObserverURL, ","),
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	})
	r.Use(c.Handler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/rooms", api.HandleCreateRoom)
		r.Post("/rooms/{code}/join", api.HandleJoinRoom)
		r.Get("/rooms/{code}/exists", api.HandleRoomExists)
	})

	r.Get("/ws", api.HandleWebSocket)

	return r
}

type CreateRoomRequest struct {
	RunnerName string `json:"runnerName"`
	Password   string `json:"password,omitempty"`
}

func ValidateRunnerName(name string) error {
	if len(name) > 24 {
		return fmt.Errorf("runner name must be 24 characters or less")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("runner name must contain only alphanumeric characters, hyphens, and underscores")
		}
	}
	return nil
}

func (api *API) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug("HandleCreateRoom called")

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RunnerName == "" {
		http.Error(w, "runnerName is required", http.StatusBadRequest)
		return
	}

	// validate runner name: alphanumeric + -_ only, max 24 chars
	if err := ValidateRunnerName(req.RunnerName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, writeToken, creatorToken, err := api.rooms.CreateRoom(req.RunnerName, req.Password)
	if err != nil {
		api.logger.Error("failed to create room", zap.Error(err))
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"roomCode":     code,
		"writeToken":   writeToken,
		"creatorToken": creatorToken,
	})
}

type JoinRoomRequest struct {
	RunnerName string `json:"runnerName"`
	Password   string `json:"password,omitempty"`
}

func (api *API) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RunnerName == "" {
		http.Error(w, "runnerName is required", http.StatusBadRequest)
		return
	}

	// validate runner name: alphanumeric + -_ only, max 24 chars
	if err := ValidateRunnerName(req.RunnerName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeToken, err := api.rooms.JoinRoom(code, req.RunnerName, req.Password)
	if err != nil {
		api.logger.Warn("failed to join room", zap.String("code", code), zap.Error(err))
		if strings.Contains(err.Error(), "room not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "invalid password") {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if strings.Contains(err.Error(), "already taken") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"writeToken": writeToken,
	})
}

func (api *API) HandleRoomExists(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if api.rooms.RoomExists(code) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "room not found", http.StatusNotFound)
	}
}

func (api *API) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	if roomCode == "" {
		http.Error(w, "missing room query param", http.StatusBadRequest)
		return
	}

	if len(roomCode) > 20 || len(roomCode) < 3 {
		http.Error(w, "invalid room code length", http.StatusBadRequest)
		return
	}

	for _, c := range roomCode {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			http.Error(w, "invalid room code format", http.StatusBadRequest)
			return
		}
	}

	room := api.rooms.GetRoom(roomCode)
	if room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	token := r.URL.Query().Get("token")
	runnerName := r.URL.Query().Get("runnerName")
	isWriter := token != "" && runnerName != ""

	if isWriter {
		if !room.ValidateToken(runnerName, token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	origin := r.Header.Get("Origin")
	if origin != "" {
		origin = normalizeOrigin(origin)
		allowedOrigins := strings.Split(api.config.ObserverURL, ",")
		allowed := false
		for _, ao := range allowedOrigins {
			allowed = allowed || (strings.TrimSpace(normalizeOrigin(ao)) == origin)
		}
		if !allowed {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode:    websocket.CompressionNoContextTakeover,
		InsecureSkipVerify: true,
	})
	if err != nil {
		api.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if isWriter {
		api.handleWriter(ctx, c, room, runnerName, token)
	} else {
		api.handleObserver(ctx, c, room)
	}
}

func (api *API) handleWriter(ctx context.Context, c *websocket.Conn, room *RoomState, runnerName, token string) {
	api.logger.Info("runner connected", zap.String("room", room.code), zap.String("runner", runnerName))

	defer func() {
		offlineMsg := RoomMessage{
			Type:       "presence",
			Ts:         time.Now().UnixMilli(),
			Source:     "server",
			RunnerName: runnerName,
			Payload:    json.RawMessage(`{"status":"offline"}`),
		}
		data, _ := json.Marshal(offlineMsg)
		room.UpdateSnapshot(runnerName, "presence", data)

		if room.IsCreator(runnerName, token) {
			// cancel grace period before closing room
			room.cancelGracePeriod()
			roomClosedMsg := RoomMessage{
				Type:       "room_closed",
				Ts:         time.Now().UnixMilli(),
				Source:     "server",
				RunnerName: runnerName,
				Payload:    json.RawMessage(`{"message":"The room creator has closed the relay. This room is no longer active."}`),
			}
			roomClosedData, _ := json.Marshal(roomClosedMsg)
			room.BroadcastToObservers(roomClosedData)
			api.rooms.CloseRoom(room.code)
		} else {
			room.WriterDisconnected(runnerName)
		}
	}()

	onlineMsg := RoomMessage{
		Type:       "presence",
		Ts:         time.Now().UnixMilli(),
		Source:     "server",
		RunnerName: runnerName,
		Payload:    json.RawMessage(`{"status":"online"}`),
	}
	data, _ := json.Marshal(onlineMsg)
	room.UpdateSnapshot(runnerName, "presence", data)

	for {
		var v json.RawMessage
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			break
		}

		var msg RoomMessage
		if err := json.Unmarshal(v, &msg); err != nil {
			api.logger.Warn("invalid message format", zap.Error(err))
			continue
		}

		msg.RunnerName = runnerName
		finalData, _ := json.Marshal(msg)

		room.UpdateSnapshot(runnerName, msg.Type, finalData)
	}
}

func (api *API) handleObserver(ctx context.Context, c *websocket.Conn, room *RoomState) {
	api.logger.Info("observer connected", zap.String("room", room.code))

	spec := &ObserverClient{
		ch:   make(chan []byte, 16),
		stop: make(chan struct{}),
	}

	lastSnaps := api.rooms.AddObserver(room.code, spec)
	defer api.rooms.RemoveObserver(room.code, spec)

	for _, snap := range lastSnaps {
		if err := c.Write(ctx, websocket.MessageText, snap); err != nil {
			return
		}
	}

	disconnected := make(chan struct{})
	go func() {
		defer close(disconnected)
		for {
			_, _, err := c.Read(ctx)
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-disconnected:
			return
		case <-spec.stop:
			c.Close(4000, "room closed")
			return
		case msg := <-spec.ch:
			if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		}
	}
}
