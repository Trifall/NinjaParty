package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

var validRoomCode = regexp.MustCompile(`^[0-9a-fA-F]{3,20}$`)
var validRunnerName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,24}$`)

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
		r.Post("/rooms/{code}/force-close", api.HandleForceCloseRoom)
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
	if !validRunnerName.MatchString(name) {
		return fmt.Errorf("runner name must be 1–24 alphanumeric characters, hyphens, or underscores")
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

	code, writeToken, createdAt, err := api.rooms.CreateRoom(req.RunnerName, req.Password)
	if err != nil {
		api.logger.Error("failed to create room", zap.Error(err))
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"roomCode":   code,
		"writeToken": writeToken,
		"createdAt":  createdAt,
		"isCreator":  true,
	}); err != nil {
		api.logger.Error("failed to encode create room response", zap.Error(err))
	}
}

type JoinRoomRequest struct {
	RunnerName string `json:"runnerName"`
	Password   string `json:"password,omitempty"`
	PrevToken  string `json:"prevToken,omitempty"`
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

	writeToken, isCreator, createdAt, err := api.rooms.JoinRoom(code, req.RunnerName, req.Password, req.PrevToken)
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
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"writeToken": writeToken,
		"createdAt":  createdAt,
		"isCreator":  isCreator,
	}); err != nil {
		api.logger.Error("failed to encode join room response", zap.Error(err))
	}
}

func (api *API) HandleRoomExists(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if api.rooms.RoomExists(code) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "room not found", http.StatusNotFound)
	}
}

type ForceCloseRequest struct {
	WriteToken string `json:"writeToken"`
}

func (api *API) HandleForceCloseRoom(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var req ForceCloseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.WriteToken == "" {
		http.Error(w, "writeToken is required", http.StatusBadRequest)
		return
	}

	err := api.rooms.ForceCloseRoom(code, req.WriteToken)
	if err != nil {
		if strings.Contains(err.Error(), "room not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (api *API) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	if roomCode == "" {
		http.Error(w, "missing room query param", http.StatusBadRequest)
		return
	}

	if !validRoomCode.MatchString(roomCode) {
		http.Error(w, "invalid room code format", http.StatusBadRequest)
		return
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
	defer func() {
		if err := c.CloseNow(); err != nil {
			api.logger.Debug("websocket CloseNow error", zap.Error(err))
		}
	}()

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

	isCreator := room.IsWriterTokenCreator(token)

	// mark as connected
	room.mu.Lock()
	if w, ok := room.writers[runnerName]; ok {
		w.connected = true
	}
	room.mu.Unlock()

	writerCtx, writerCancel := context.WithCancel(ctx)
	defer writerCancel()

	go func() {
		select {
		case <-writerCtx.Done():
		case <-room.stop:
			if err := c.Close(websocket.StatusCode(4000), "room closed"); err != nil {
				api.logger.Debug("writer close error", zap.Error(err))
			}
		}
	}()

	defer func() {
		// mark as disconnected
		room.mu.Lock()
		if w, ok := room.writers[runnerName]; ok {
			w.connected = false
		}
		room.mu.Unlock()

		presencePayload := PresencePayload{
			Status:    "offline",
			IsCreator: isCreator,
		}
		payloadBytes, _ := json.Marshal(presencePayload)
		offlineMsg := RoomMessage{
			Type:       "presence",
			Ts:         time.Now().UnixMilli(),
			Source:     "server",
			RunnerName: runnerName,
			Payload:    payloadBytes,
		}
		data, _ := json.Marshal(offlineMsg)
		room.UpdateSnapshot(runnerName, "presence", data)

		room.WriterDisconnected(runnerName)
	}()

	presencePayload := PresencePayload{
		Status:    "online",
		IsCreator: isCreator,
	}
	payloadBytes, _ := json.Marshal(presencePayload)
	onlineMsg := RoomMessage{
		Type:       "presence",
		Ts:         time.Now().UnixMilli(),
		Source:     "server",
		RunnerName: runnerName,
		Payload:    payloadBytes,
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

	// send room info first
	room.mu.RLock()
	createdAt := room.createdAt
	room.mu.RUnlock()

	infoMsg := RoomMessage{
		Type:       "room_info",
		Ts:         time.Now().UnixMilli(),
		Source:     "server",
		RunnerName: "",
		Payload:    json.RawMessage(fmt.Sprintf(`{"createdAt":%d}`, createdAt.UnixMilli())),
	}
	infoData, _ := json.Marshal(infoMsg)
	if err := c.Write(ctx, websocket.MessageText, infoData); err != nil {
		return
	}

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
			if err := c.Close(websocket.StatusCode(4000), "room closed"); err != nil {
				api.logger.Debug("observer close error", zap.Error(err))
			}
			return
		case msg := <-spec.ch:
			if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		}
	}
}
