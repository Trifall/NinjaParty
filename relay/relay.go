package main

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Relay struct {
	config        *Config
	logger        *zap.Logger
	runnerName    string
	roomRes       *RoomResponse
	roomCreatedAt time.Time
	messageChan   chan RoomMessage

	clientsMu             sync.Mutex
	sseClients            []*SSEClient
	wsClient              *WSClient
	ninjabrainConnected   bool
	ninjabrainClientsDone int

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

func (r *Relay) SetRunnerName(name string) {
	r.runnerName = name
}

func (r *Relay) StartRoom(roomCode, writeToken string) error {
	r.roomRes = &RoomResponse{
		RoomCode:   roomCode,
		WriteToken: writeToken,
		IsCreator:  true,
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
		sseClient := NewSSEClient(r.config.NinjabrainURL+endpoint, evType, r.config.TLSSkipVerify, r.logger)
		sseClient.SetCallbacks(func() {
			r.clientsMu.Lock()
			defer r.clientsMu.Unlock()
			wasConnected := r.ninjabrainConnected
			r.ninjabrainClientsDone++
			if !wasConnected {
				r.ninjabrainConnected = true
				r.logger.Info("ninjabrain API connected")
			}
		}, func() {
			r.clientsMu.Lock()
			defer r.clientsMu.Unlock()
			if r.ninjabrainClientsDone > 0 {
				r.ninjabrainClientsDone--
			}
			if r.ninjabrainClientsDone == 0 && r.ninjabrainConnected {
				r.ninjabrainConnected = false
				r.logger.Info("ninjabrain API disconnected")
			}
		})
		r.sseClients = append(r.sseClients, sseClient)
		go sseClient.Run(r.ctx, r.messageChan, r.runnerName)
	}

	r.wsClient = NewWSClient(
		r.config.ServerWSURL,
		r.roomRes.RoomCode,
		r.runnerName,
		r.roomRes.WriteToken,
		r.config.TLSSkipVerify,
		time.Duration(r.config.HeartbeatIntvl)*time.Second,
		r.logger,
		func() {
			r.logger.Info("room closed by server, shutting down relay")
			go r.Shutdown()
		},
	)
	go r.wsClient.Run(r.ctx, r.messageChan)
}

func (r *Relay) stopClientsLocked() {
	for _, client := range r.sseClients {
		client.Close()
	}
	r.sseClients = nil
	r.ninjabrainConnected = false
	r.ninjabrainClientsDone = 0
	if r.wsClient != nil {
		r.wsClient.Close()
		r.wsClient = nil
	}
}

func (r *Relay) stopAllClientsLocked() {
	if r.cancel != nil {
		r.cancel()
	}
	r.stopClientsLocked()
}

func (r *Relay) CreateNewRoom(password string) (*RoomResponse, error) {
	r.logger.Info("creating new room")

	r.clientsMu.Lock()
	r.stopClientsLocked()
	r.clientsMu.Unlock()

	roomRes, err := createRoom(r.config.ServerAPIURL, r.runnerName, password)
	if err != nil {
		return nil, err
	}

	r.roomRes = &RoomResponse{
		RoomCode:   roomRes.RoomCode,
		WriteToken: roomRes.WriteToken,
		CreatedAt:  roomRes.CreatedAt,
		IsCreator:  roomRes.IsCreator,
	}
	r.roomCreatedAt = time.UnixMilli(roomRes.CreatedAt)

	r.logger.Info("room created", zap.String("code", roomRes.RoomCode))

	r.startClients()

	return roomRes, nil
}

func (r *Relay) JoinRoom(roomCode, password, prevWriteToken string) (*RoomResponse, error) {
	r.logger.Info("joining room", zap.String("code", roomCode))

	r.clientsMu.Lock()
	r.stopClientsLocked()
	r.clientsMu.Unlock()

	roomRes, err := joinRoom(r.config.ServerAPIURL, roomCode, r.runnerName, password, prevWriteToken)
	if err != nil {
		return nil, err
	}

	r.roomRes = &RoomResponse{
		RoomCode:   roomRes.RoomCode,
		WriteToken: roomRes.WriteToken,
		CreatedAt:  roomRes.CreatedAt,
		IsCreator:  roomRes.IsCreator,
	}
	r.roomCreatedAt = time.UnixMilli(roomRes.CreatedAt)

	r.logger.Info("joined room", zap.String("code", roomRes.RoomCode))

	r.startClients()

	return roomRes, nil
}

func (r *Relay) RoomCode() string {
	if r.roomRes == nil {
		return ""
	}
	return r.roomRes.RoomCode
}

func (r *Relay) RunnerName() string {
	return r.runnerName
}

func (r *Relay) IsCreator() bool {
	return r.roomRes != nil && r.roomRes.IsCreator
}

func (r *Relay) NinjabrainConnected() bool {
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	return r.ninjabrainConnected
}

func (r *Relay) FullLink() string {
	return r.config.ObserverURL + r.RoomCode()
}

func (r *Relay) RoomCreatedAt() time.Time {
	return r.roomCreatedAt
}

func (r *Relay) Shutdown() {
	r.logger.Info("closing room", zap.String("code", r.RoomCode()))
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()
	r.stopAllClientsLocked()
	r.roomRes = nil
	r.roomCreatedAt = time.Time{}
}

func (r *Relay) ServerAPIURL() string {
	return r.config.ServerAPIURL
}
