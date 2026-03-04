package main

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRoomManager_CreateRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)

	code, token, createdAt, err := rm.CreateRoom("runner445", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(code) != 6 { // hex of 3 bytes is 6 chars
		t.Errorf("expected code length 6, got %d", len(code))
	}

	if len(token) != 32 { // hex of 16 bytes is 32 chars
		t.Errorf("expected token length 32, got %d", len(token))
	}

	if createdAt == 0 {
		t.Errorf("expected createdAt to be non-zero")
	}

	room := rm.GetRoom(code)
	if room == nil {
		t.Fatalf("expected room to exist")
	}
	if room.creatorName != "runner445" {
		t.Errorf("expected creatorName to be runner445, got %s", room.creatorName)
	}
}

func TestRoomManager_JoinRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)

	code, creatorToken, _, err := rm.CreateRoom("runner445", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 1. join with wrong password
	_, _, _, err = rm.JoinRoom(code, "Bob", "wrong_secret", "")
	if err == nil || !strings.Contains(err.Error(), "invalid password") {
		t.Errorf("expected invalid password error, got %v", err)
	}

	// 2. join with correct password
	bobToken, isCreator, _, err := rm.JoinRoom(code, "Bob", "secret", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if isCreator {
		t.Errorf("expected Bob not to be creator")
	}
	if len(bobToken) != 32 {
		t.Errorf("expected bobToken length 32, got %d", len(bobToken))
	}

	// 3. join with existing name (Bob again, different token)
	_, _, _, err = rm.JoinRoom(code, "Bob", "secret", "")
	if err == nil || !strings.Contains(err.Error(), "runner name already taken") {
		t.Errorf("expected runner name already taken error, got %v", err)
	}

	// 4. join with existing name (Bob again, same token)
	bobToken2, _, _, err := rm.JoinRoom(code, "Bob", "secret", bobToken)
	if err != nil {
		t.Fatalf("expected no error for reconnection, got %v", err)
	}
	if bobToken != bobToken2 {
		t.Errorf("expected identical token on reconnect")
	}

	// 5. creator rejoins
	runner445Token, isCreatorA, _, err := rm.JoinRoom(code, "runner445", "secret", creatorToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !isCreatorA {
		t.Errorf("expected runner445 to remain creator")
	}
	if runner445Token != creatorToken {
		t.Errorf("expected identical token on creator reconnect")
	}

	// 6. non-existent room
	_, _, _, err = rm.JoinRoom("invalid_code", "Charlie", "secret", "")
	if err == nil || !strings.Contains(err.Error(), "room not found") {
		t.Errorf("expected room not found error, got %v", err)
	}
}

func TestRoomManager_JoinRoomFull(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)

	code, _, _, err := rm.CreateRoom("runner445", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// fill the room up to maxWritersPerRoom (32 initially, 1 is runner445, loop 31 times)
	for i := 1; i < maxWritersPerRoom; i++ {
		name := string(rune('A' + i)) // 'B', 'C', ...
		_, _, _, err := rm.JoinRoom(code, name, "", "")
		if err != nil {
			t.Fatalf("expected no error on join %d, got %v", i, err)
		}
	}

	// one more should fail
	_, _, _, err = rm.JoinRoom(code, "Overflow", "", "")
	if err == nil || !strings.Contains(err.Error(), "room is full") {
		t.Errorf("expected room is full error, got %v", err)
	}
}

func TestRoomManager_ForceCloseRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)

	code, creatorToken, _, err := rm.CreateRoom("runner445", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bobToken, _, _, err := rm.JoinRoom(code, "Bob", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Bob tries to force close
	err = rm.ForceCloseRoom(code, bobToken)
	if err == nil || !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected unauthorized error, got %v", err)
	}

	// runner445 tries to force close
	err = rm.ForceCloseRoom(code, creatorToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check room is closed
	room := rm.GetRoom(code)
	if room != nil {
		t.Errorf("expected room to be closed")
	}
}

func TestRoomManager_EvictIdleRooms(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Millisecond*50, logger)

	code, _, _, err := rm.CreateRoom("runner445", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// sleep for more than the TTL
	time.Sleep(time.Millisecond * 100)

	rm.evictIdleRooms()

	room := rm.GetRoom(code)
	if room != nil {
		t.Errorf("expected room to be evicted")
	}
}

func TestRoomState_UpdateSnapshot(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	code, _, _, _ := rm.CreateRoom("runner445", "")
	room := rm.GetRoom(code)

	room.UpdateSnapshot("runner445", "position", []byte("data1"))

	room.mu.RLock()
	data := room.lastSnapshots["runner445"]["position"]
	room.mu.RUnlock()

	if string(data) != "data1" {
		t.Errorf("expected data1, got %s", string(data))
	}
}

func TestRoomState_AddObserver(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	code, _, _, _ := rm.CreateRoom("runner445", "")
	room := rm.GetRoom(code)

	room.UpdateSnapshot("runner445", "position", []byte("data1"))

	observer := &ObserverClient{
		ch:   make(chan []byte, 10),
		stop: make(chan struct{}),
	}

	snapshots := rm.AddObserver(code, observer)
	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	} else if string(snapshots[0]) != "data1" {
		t.Errorf("expected data1, got %s", string(snapshots[0]))
	}

	room.BroadcastToObservers([]byte("data2"))

	select {
	case msg := <-observer.ch:
		if string(msg) != "data2" {
			t.Errorf("expected data2, got %s", string(msg))
		}
	case <-time.After(time.Second):
		t.Errorf("timeout waiting for observer message")
	}

	rm.RemoveObserver(code, observer)
}
