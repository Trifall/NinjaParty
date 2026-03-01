package main

import (
	"context"
	"encoding/json"
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
	})

	r.Get("/ws", api.HandleWebSocket)

	return r
}

func (api *API) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	api.logger.Debug("HandleCreateRoom called")
	code, token, err := api.rooms.CreateRoom()
	if err != nil {
		api.logger.Error("failed to create room", zap.Error(err))
		http.Error(w, "failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"roomCode":   code,
		"writeToken": token,
	})
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
	isWriter := token != ""

	if isWriter {
		if !room.ValidateToken(token) {
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
		InsecureSkipVerify: true, // own origin validation above
	})
	if err != nil {
		api.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if isWriter {
		api.handleWriter(ctx, c, room)
	} else {
		api.handleObserver(ctx, c, room)
	}
}

func (api *API) handleWriter(ctx context.Context, c *websocket.Conn, room *RoomState) {
	api.logger.Info("runner connected", zap.String("room", room.code))

	// ensure presence offline is sent on disconnect
	defer func() {
		api.logger.Debug("defer: starting cleanup", zap.String("room", room.code))

		offlineMsg := RoomMessage{
			Type:    "presence",
			Ts:      time.Now().UnixMilli(),
			Source:  "server",
			Payload: json.RawMessage(`{"status":"offline"}`),
		}
		data, _ := json.Marshal(offlineMsg)
		api.logger.Debug("defer: calling UpdateSnapshot", zap.String("room", room.code))
		room.UpdateSnapshot("presence", data)
		api.logger.Debug("defer: runner disconnected", zap.String("room", room.code))

		// notify observers that room is closing, then close the room
		roomClosedMsg := RoomMessage{
			Type:    "room_closed",
			Ts:      time.Now().UnixMilli(),
			Source:  "server",
			Payload: json.RawMessage(`{"message":"The runner has closed the relay. This room is no longer active."}`),
		}
		roomClosedData, _ := json.Marshal(roomClosedMsg)
		api.logger.Debug("defer: before broadcast", zap.String("room", room.code))
		room.BroadcastToObservers(roomClosedData)
		api.logger.Debug("defer: after broadcast", zap.String("room", room.code))

		// close the room and kick all observers - runner has left
		api.rooms.CloseRoom(room.code)
		api.logger.Debug("defer: room closed successfully", zap.String("room", room.code))
	}()

	// send online presence immediately
	onlineMsg := RoomMessage{
		Type:    "presence",
		Ts:      time.Now().UnixMilli(),
		Source:  "server",
		Payload: json.RawMessage(`{"status":"online"}`),
	}
	data, _ := json.Marshal(onlineMsg)
	room.UpdateSnapshot("presence", data)

	for {
		var v json.RawMessage
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			break
		}

		var msgType struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(v, &msgType); err != nil {
			api.logger.Warn("invalid message format", zap.Error(err))
			continue
		}

		// when wsjson.Read completes it parses it to []byte, no need to wrap again
		room.UpdateSnapshot(msgType.Type, v)
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

	// send all last snapshots immediately
	for _, snap := range lastSnaps {
		if err := c.Write(ctx, websocket.MessageText, snap); err != nil {
			return
		}
	}

	// read from WebSocket to detect disconnect
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
