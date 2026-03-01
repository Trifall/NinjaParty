package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxWritersPerRoom   = 32
	gracePeriodDuration = 120 * time.Second
)

type WriterState struct {
	name       string
	writeToken string
	lastSeen   time.Time
}

type RoomState struct {
	mu            sync.RWMutex
	code          string
	passwordHash  string
	creatorName   string
	creatorToken  string
	lastSnapshots map[string]map[string][]byte // runnerName -> msgType -> data
	lastUpdated   time.Time
	writers       map[string]*WriterState

	gracePeriodTimer *time.Timer
	gracePeriodMu    sync.Mutex

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

func (rm *RoomManager) CreateRoom(creatorName, password string) (string, string, string, error) {
	const maxAttempts = 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := GenerateRandomHex(3)
		if err != nil {
			return "", "", "", err
		}
		writeToken, err := GenerateRandomHex(16)
		if err != nil {
			return "", "", "", err
		}
		creatorToken, err := GenerateRandomHex(16)
		if err != nil {
			return "", "", "", err
		}

		rm.mu.Lock()

		if rm.rooms[code] != nil {
			rm.mu.Unlock()
			continue
		}

		var hash string
		if password != "" {
			b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				rm.mu.Unlock()
				return "", "", "", err
			}
			hash = string(b)
		}

		room := &RoomState{
			code:          code,
			passwordHash:  hash,
			creatorName:   creatorName,
			creatorToken:  creatorToken,
			lastUpdated:   time.Now(),
			lastSnapshots: make(map[string]map[string][]byte),
			writers:       make(map[string]*WriterState),
			observers:     make(map[*ObserverClient]bool),
		}
		room.writers[creatorName] = &WriterState{
			name:       creatorName,
			writeToken: writeToken,
			lastSeen:   time.Now(),
		}

		rm.rooms[code] = room

		rm.mu.Unlock()
		rm.logRoomStats("room created")
		return code, writeToken, creatorToken, nil
	}

	return "", "", "", fmt.Errorf("failed to generate unique room code after %d attempts", maxAttempts)
}

func (rm *RoomManager) JoinRoom(code, runnerName, password string) (string, error) {
	rm.mu.Lock()
	room, exists := rm.rooms[code]
	rm.mu.Unlock()

	if !exists {
		return "", fmt.Errorf("room not found")
	}

	if room.passwordHash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(room.passwordHash), []byte(password)); err != nil {
			return "", fmt.Errorf("invalid password")
		}
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.writers) >= maxWritersPerRoom {
		return "", fmt.Errorf("room is full")
	}

	lowerName := strings.ToLower(runnerName)
	for name, writer := range room.writers {
		if strings.ToLower(name) == lowerName {
			if time.Since(writer.lastSeen) > gracePeriodDuration {
				delete(room.writers, name)
			} else {
				return "", fmt.Errorf("runner name already taken")
			}
		}
	}

	writeToken, err := GenerateRandomHex(16)
	if err != nil {
		return "", err
	}

	room.writers[runnerName] = &WriterState{
		name:       runnerName,
		writeToken: writeToken,
		lastSeen:   time.Now(),
	}

	if room.lastSnapshots[runnerName] == nil {
		room.lastSnapshots[runnerName] = make(map[string][]byte)
	}

	room.cancelGracePeriod()

	return writeToken, nil
}

func (rm *RoomManager) GetRoom(code string) *RoomState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[code]
}

func (rm *RoomManager) RoomExists(code string) bool {
	rm.mu.RLock()
	_, exists := rm.rooms[code]
	rm.mu.RUnlock()
	return exists
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

func (r *RoomState) ValidateToken(runnerName, token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	writer, exists := r.writers[runnerName]
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(writer.writeToken), []byte(token)) == 1
}

func (r *RoomState) IsCreator(runnerName, token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.creatorName == runnerName && subtle.ConstantTimeCompare([]byte(r.creatorToken), []byte(token)) == 1
}

func (r *RoomState) cancelGracePeriod() {
	r.gracePeriodMu.Lock()
	defer r.gracePeriodMu.Unlock()
	if r.gracePeriodTimer != nil {
		r.gracePeriodTimer.Stop()
		r.gracePeriodTimer = nil
	}
}

func (r *RoomState) startGracePeriod(closeCallback func()) {
	r.gracePeriodMu.Lock()
	defer r.gracePeriodMu.Unlock()

	if r.gracePeriodTimer != nil {
		r.gracePeriodTimer.Stop()
	}

	r.gracePeriodTimer = time.AfterFunc(gracePeriodDuration, func() {
		closeCallback()
	})
}

func (r *RoomState) UpdateSnapshot(runnerName, msgType string, data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastSnapshots[runnerName] == nil {
		r.lastSnapshots[runnerName] = make(map[string][]byte)
	}
	r.lastSnapshots[runnerName][msgType] = data
	r.lastUpdated = time.Now()

	writer, exists := r.writers[runnerName]
	if exists {
		writer.lastSeen = time.Now()
	}

	for spec := range r.observers {
		select {
		case spec.ch <- data:
		default:
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
	for _, runnerSnaps := range r.lastSnapshots {
		for _, snap := range runnerSnaps {
			snaps = append(snaps, snap)
		}
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

func (r *RoomState) WriterDisconnected(runnerName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	writer, exists := r.writers[runnerName]
	if !exists {
		return
	}
	writer.lastSeen = time.Now()

	// count writers seen recently
	connectedCount := 0
	for name, w := range r.writers {
		if name != runnerName && time.Since(w.lastSeen) < gracePeriodDuration {
			connectedCount++
		}
	}

	if connectedCount == 0 {
		r.startGracePeriod(func() {
			BroadcastRoomClosed(r.code, "All runners have disconnected")
		})
	}
}

func BroadcastRoomClosed(code, message string) {
	RoomManagerInstance.mu.Lock()
	room, exists := RoomManagerInstance.rooms[code]
	if exists {
		delete(RoomManagerInstance.rooms, code)
	}
	RoomManagerInstance.mu.Unlock()

	if exists {
		roomClosedMsg := RoomMessage{
			Type:       "room_closed",
			Ts:         time.Now().UnixMilli(),
			Source:     "server",
			RunnerName: "",
			Payload:    json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, message)),
		}
		data, _ := json.Marshal(roomClosedMsg)
		room.BroadcastToObservers(data)
		room.closeObservers()
		RoomManagerInstance.logRoomStats("room closed (grace period)")
	}
}

var RoomManagerInstance *RoomManager
