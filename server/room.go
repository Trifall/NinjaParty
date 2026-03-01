package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type RoomState struct {
	mu            sync.RWMutex
	code          string
	writeToken    string
	lastSnapshots map[string][]byte
	lastUpdated   time.Time

	// using channels to fan-out messages to observers
	observers map[*ObserverClient]bool
}

type ObserverClient struct {
	ch   chan []byte
	stop chan struct{}
}

type RoomManager struct {
	mu     sync.RWMutex
	rooms  map[string]*RoomState
	ttl    time.Duration
	logger *zap.Logger
}

func NewRoomManager(ttl time.Duration, logger *zap.Logger) *RoomManager {
	return &RoomManager{
		rooms:  make(map[string]*RoomState),
		ttl:    ttl,
		logger: logger,
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
	const maxAttempts = 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := GenerateRandomHex(3)
		if err != nil {
			return "", "", err
		}
		token, err := GenerateRandomHex(16)
		if err != nil {
			return "", "", err
		}

		rm.mu.Lock()

		if rm.rooms[code] != nil {
			rm.mu.Unlock()
			continue
		}

		rm.rooms[code] = &RoomState{
			code:          code,
			writeToken:    token,
			lastUpdated:   time.Now(),
			lastSnapshots: make(map[string][]byte),
			observers:     make(map[*ObserverClient]bool),
		}

		rm.mu.Unlock()
		rm.logRoomStats("room created")
		return code, token, nil
	}

	return "", "", fmt.Errorf("failed to generate unique room code after %d attempts", maxAttempts)
}

func (rm *RoomManager) GetRoom(code string) *RoomState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

func (rm *RoomManager) getCounts() (int, int) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	roomCount := len(rm.rooms)
	observerCount := 0
	for _, room := range rm.rooms {
		room.mu.RLock()
		observerCount += len(room.observers)
		room.mu.RUnlock()
	}
	return roomCount, observerCount
}

func (rm *RoomManager) logRoomStats(action string) {
	roomCount, observerCount := rm.getCounts()
	rm.logger.Info(action,
		zap.Int("total_rooms", roomCount),
		zap.Int("total_observers", observerCount),
	)
}

func (rm *RoomManager) DeleteRoom(code string) {
	rm.mu.Lock()
	room, exists := rm.rooms[code]
	if exists {
		delete(rm.rooms, code)
	}
	rm.mu.Unlock()

	if exists {
		room.closeObservers()
		rm.logRoomStats("room closed")
	}
}

func (rm *RoomManager) CloseRoom(code string) {
	rm.mu.Lock()
	room, exists := rm.rooms[code]
	if exists {
		delete(rm.rooms, code)
	}
	rm.mu.Unlock()

	if exists {
		rm.logger.Debug("CloseRoom: closing observers", zap.String("code", code))
		room.closeObservers()
		rm.logRoomStats("room closed")
	}
}

func (r *RoomState) closeObservers() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for spec := range r.observers {
		close(spec.stop)
	}
	r.observers = make(map[*ObserverClient]bool)
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
	now := time.Now()
	var toEvict []*RoomState
	var toEvictCodes []string
	for code, room := range rm.rooms {
		room.mu.RLock()
		updated := room.lastUpdated
		room.mu.RUnlock()

		if now.Sub(updated) > rm.ttl {
			toEvict = append(toEvict, room)
			toEvictCodes = append(toEvictCodes, code)
			delete(rm.rooms, code)
		}
	}
	rm.mu.Unlock()

	for i, room := range toEvict {
		room.closeObservers()
		rm.logger.Info("room evicted", zap.String("code", toEvictCodes[i]))
	}
	if len(toEvict) > 0 {
		rm.logRoomStats("eviction complete")
	}
}

func (r *RoomState) ValidateToken(token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return subtle.ConstantTimeCompare([]byte(r.writeToken), []byte(token)) == 1
}

func (r *RoomState) UpdateSnapshot(msgType string, data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastSnapshots[msgType] = data
	r.lastUpdated = time.Now()

	// Broadcast
	for spec := range r.observers {
		select {
		case spec.ch <- data:
		default:
			// If buffer is full, drop to prevent blocking
		}
	}
}

func (r *RoomState) AddObserver(spec *ObserverClient) [][]byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.observers[spec] {
		return nil
	}

	r.observers[spec] = true

	var snaps [][]byte
	for _, snap := range r.lastSnapshots {
		snaps = append(snaps, snap)
	}

	return snaps
}

func (rm *RoomManager) AddObserver(roomCode string, spec *ObserverClient) [][]byte {
	rm.mu.RLock()
	room := rm.rooms[roomCode]
	rm.mu.RUnlock()

	if room == nil {
		return nil
	}

	snaps := room.AddObserver(spec)
	room.mu.RLock()
	count := len(room.observers)
	room.mu.RUnlock()

	rm.logger.Info("observer joined",
		zap.String("room", roomCode),
		zap.Int("room_observers", count),
	)
	rm.logRoomStats("observer joined")

	return snaps
}

func (r *RoomState) BroadcastToObservers(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for spec := range r.observers {
		select {
		case spec.ch <- data:
		default:
		}
	}
}

func (r *RoomState) RemoveObserver(spec *ObserverClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.observers, spec)
}

func (rm *RoomManager) RemoveObserver(roomCode string, spec *ObserverClient) {
	rm.mu.RLock()
	room := rm.rooms[roomCode]
	rm.mu.RUnlock()

	if room == nil {
		return
	}

	room.RemoveObserver(spec)
	room.mu.RLock()
	count := len(room.observers)
	room.mu.RUnlock()

	rm.logger.Info("observer left",
		zap.String("room", roomCode),
		zap.Int("room_observers", count),
	)
	rm.logRoomStats("observer left")
}
