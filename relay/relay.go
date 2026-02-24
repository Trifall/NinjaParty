package main

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Relay struct {
	config      *Config
	logger      *zap.Logger
	roomRes     *RoomResponse
	messageChan chan RoomMessage

	clientsMu  sync.Mutex
	sseClients []*SSEClient
	wsClient   *WSClient

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRelay(config *Config, logger *zap.Logger) *Relay {
	ctx, cancel := context.WithCancel(context.Background())
	return &Relay{
		config:      config,
		logger:      logger,
		messageChan: make(chan RoomMessage, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (r *Relay) StartRoom(roomCode, writeToken string) error {
	r.roomRes = &RoomResponse{
		RoomCode:   roomCode,
		WriteToken: writeToken,
	}

	r.logger.Info("starting clients for room", zap.String("code", roomCode))

	r.startClients()

	return nil
}

func (r *Relay) startClients() {
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()

	r.stopClientsLocked()

	r.ctx, r.cancel = context.WithCancel(context.Background())

	endpoints := map[string]string{
		"/api/v1/stronghold/events":       "stronghold_update",
		"/api/v1/boat/events":             "boat_update",
		"/api/v1/blind/events":            "blind_update",
		"/api/v1/divine/events":           "divine_update",
		"/api/v1/all-advancements/events": "all_advancements_update",
	}

	for endpoint, evType := range endpoints {
		sseClient := NewSSEClient(r.config.NinjabrainURL+endpoint, evType, r.logger)
		r.sseClients = append(r.sseClients, sseClient)
		go sseClient.Run(r.ctx, r.messageChan)
	}

	r.wsClient = NewWSClient(
		r.config.ServerWSURL,
		r.roomRes.RoomCode,
		r.roomRes.WriteToken,
		r.config.TLSSkipVerify,
		time.Duration(r.config.HeartbeatIntvl)*time.Second,
		r.logger,
	)
	go r.wsClient.Run(r.ctx, r.messageChan)
}

func (r *Relay) stopClientsLocked() {
	if r.cancel != nil {
		r.cancel()
	}
	r.sseClients = nil
	r.wsClient = nil
}

func (r *Relay) CreateNewRoom() (*RoomResponse, error) {
	r.logger.Info("creating new room")

	r.clientsMu.Lock()
	r.stopClientsLocked()

	roomRes, err := createRoom(r.config.ServerAPIURL)
	if err != nil {
		r.clientsMu.Unlock()
		return nil, err
	}

	r.roomRes = &RoomResponse{
		RoomCode:   roomRes.RoomCode,
		WriteToken: roomRes.WriteToken,
	}

	r.logger.Info("starting clients for new room", zap.String("code", roomRes.RoomCode))
	r.startClientsLocked()

	r.clientsMu.Unlock()

	return roomRes, nil
}

func (r *Relay) startClientsLocked() {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	endpoints := map[string]string{
		"/api/v1/stronghold/events":       "stronghold_update",
		"/api/v1/boat/events":             "boat_update",
		"/api/v1/blind/events":            "blind_update",
		"/api/v1/divine/events":           "divine_update",
		"/api/v1/all-advancements/events": "all_advancements_update",
	}

	for endpoint, evType := range endpoints {
		sseClient := NewSSEClient(r.config.NinjabrainURL+endpoint, evType, r.logger)
		r.sseClients = append(r.sseClients, sseClient)
		go sseClient.Run(r.ctx, r.messageChan)
	}

	r.wsClient = NewWSClient(
		r.config.ServerWSURL,
		r.roomRes.RoomCode,
		r.roomRes.WriteToken,
		r.config.TLSSkipVerify,
		time.Duration(r.config.HeartbeatIntvl)*time.Second,
		r.logger,
	)
	go r.wsClient.Run(r.ctx, r.messageChan)
}

func (r *Relay) RoomCode() string {
	if r.roomRes == nil {
		return ""
	}
	return r.roomRes.RoomCode
}

func (r *Relay) FullLink() string {
	return r.config.ObserverURL + r.RoomCode()
}

func (r *Relay) Shutdown() {
	r.logger.Info("shutting down relay")
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	r.stopClientsLocked()
}
