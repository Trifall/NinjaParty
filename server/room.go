package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type RoomState struct {
	mu           sync.RWMutex
	code         string
	writeToken   string
	lastSnapshot []byte
	lastUpdated  time.Time

	// using channels to fan-out messages to spectators
	spectators map[*SpectatorClient]bool
}

type SpectatorClient struct {
	ch   chan []byte
	stop chan struct{}
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*RoomState
	ttl   time.Duration
}

func NewRoomManager(ttl time.Duration) *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*RoomState),
		ttl:   ttl,
	}
}

func GenerateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (rm *RoomManager) CreateRoom() (string, string, error) {
	code, err := GenerateRandomHex(3)
	if err != nil {
		return "", "", err
	}
	token, err := GenerateRandomHex(16)
	if err != nil {
		return "", "", err
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	for rm.rooms[code] != nil {
		code, err = GenerateRandomHex(3)
		if err != nil {
			return "", "", err
		}
	}

	rm.rooms[code] = &RoomState{
		code:        code,
		writeToken:  token,
		lastUpdated: time.Now(),
		spectators:  make(map[*SpectatorClient]bool),
	}

	return code, token, nil
}

func (rm *RoomManager) GetRoom(code string) *RoomState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

func (rm *RoomManager) DeleteRoom(code string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[code]; exists {
		room.mu.Lock()
		// close all spectator connections
		for spec := range room.spectators {
			close(spec.stop)
		}
		room.mu.Unlock()
		delete(rm.rooms, code)
	}
}

func (rm *RoomManager) StartEvictionTicker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rm.evictIdleRooms()
		}
	}
}

func (rm *RoomManager) evictIdleRooms() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	for code, room := range rm.rooms {
		room.mu.RLock()
		updated := room.lastUpdated
		room.mu.RUnlock()

		if now.Sub(updated) > rm.ttl {
			// Room is idle, delete it
			room.mu.Lock()
			for spec := range room.spectators {
				close(spec.stop)
			}
			room.mu.Unlock()
			delete(rm.rooms, code)
		}
	}
}

func (r *RoomState) ValidateToken(token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.writeToken == token
}

func (r *RoomState) UpdateSnapshot(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastSnapshot = data
	r.lastUpdated = time.Now()

	// Broadcast
	for spec := range r.spectators {
		select {
		case spec.ch <- data:
		default:
			// If buffer is full, drop to prevent blocking
		}
	}
}

func (r *RoomState) AddSpectator(spec *SpectatorClient) []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spectators[spec] = true

	return r.lastSnapshot
}

func (r *RoomState) RemoveSpectator(spec *SpectatorClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.spectators, spec)
}
