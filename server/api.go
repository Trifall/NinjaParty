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
		AllowedOrigins: strings.Split(api.config.UIDomain, ","),
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

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // we have CORS middleware before this covering it usually, but typically we allow all for websocket anyway
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
		api.handleSpectator(ctx, c, room)
	}
}

func (api *API) handleWriter(ctx context.Context, c *websocket.Conn, room *RoomState) {
	api.logger.Info("writer connected", zap.String("room", room.code))

	// ensure presence offline is sent on disconnect
	defer func() {
		offlineMsg := RoomMessage{
			Type:    "presence",
			Ts:      time.Now().UnixMilli(),
			Source:  "server",
			Payload: json.RawMessage(`{"status":"offline"}`),
		}
		data, _ := json.Marshal(offlineMsg)
		room.UpdateSnapshot(data)
		api.logger.Info("writer disconnected", zap.String("room", room.code))
	}()

	// send online presence immediately
	onlineMsg := RoomMessage{
		Type:    "presence",
		Ts:      time.Now().UnixMilli(),
		Source:  "server",
		Payload: json.RawMessage(`{"status":"online"}`),
	}
	data, _ := json.Marshal(onlineMsg)
	room.UpdateSnapshot(data)

	for {
		var v json.RawMessage
		err := wsjson.Read(ctx, c, &v)
		if err != nil {
			break
		}

		room.UpdateSnapshot(v)
	}
}

func (api *API) handleSpectator(ctx context.Context, c *websocket.Conn, room *RoomState) {
	api.logger.Info("spectator connected", zap.String("room", room.code))

	spec := &SpectatorClient{
		ch:   make(chan []byte, 16),
		stop: make(chan struct{}),
	}

	lastSnap := room.AddSpectator(spec)
	defer room.RemoveSpectator(spec)

	// send last snapshot immediately if exists
	if lastSnap != nil {
		if err := wsjson.Write(ctx, c, json.RawMessage(lastSnap)); err != nil {
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-spec.stop:
			c.Close(websocket.StatusNormalClosure, "room deleted")
			return
		case msg := <-spec.ch:
			if err := wsjson.Write(ctx, c, json.RawMessage(msg)); err != nil {
				return
			}
		}
	}
}
