package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAPI_HandleCreateRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	config := &Config{ObserverURL: "http://localhost:5173"}
	api := NewAPI(config, logger, rm)
	router := api.Router()

	// 1. success create room
	reqBody := `{"runnerName": "runner_test", "password": "pass"}`
	req, _ := http.NewRequest("POST", "/api/v1/rooms", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp["code"] == "" || resp["writeToken"] == "" || resp["createdAt"] == nil {
		t.Errorf("missing expected fields in response: %v", resp)
	}

	// 2. invalid Payload
	req, _ = http.NewRequest("POST", "/api/v1/rooms", strings.NewReader(`"invalid"`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected bad request, got %v", rr.Code)
	}

	// 3. invalid runner name
	req, _ = http.NewRequest("POST", "/api/v1/rooms", strings.NewReader(`{"runnerName": ""}`))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected bad request for missing runner name, got %v", rr.Code)
	}
}

func TestAPI_HandleJoinRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	config := &Config{ObserverURL: "http://localhost:5173"}
	api := NewAPI(config, logger, rm)
	router := api.Router()

	code, _, _, _ := rm.CreateRoom("runner_test", "pass")

	// 1. success join room
	reqBody := `{"runnerName": "runner_test2", "password": "pass"}`
	req, _ := http.NewRequest("POST", "/api/v1/rooms/"+code+"/join", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["writeToken"] == "" || resp["isCreator"] == true {
		t.Errorf("unexpected response: %v", resp)
	}

	// 2. invalid password
	reqBody = `{"runnerName": "Charlie", "password": "wrong"}`
	req, _ = http.NewRequest("POST", "/api/v1/rooms/"+code+"/join", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected unauthorized, got %v", rr.Code)
	}

	// 3. room not found
	reqBody = `{"runnerName": "Charlie", "password": "pass"}`
	req, _ = http.NewRequest("POST", "/api/v1/rooms/invalid/join", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected not found, got %v", rr.Code)
	}
}

func TestAPI_HandleRoomExists(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	config := &Config{ObserverURL: "http://localhost:5173"}
	api := NewAPI(config, logger, rm)
	router := api.Router()

	code, _, _, _ := rm.CreateRoom("runner_test", "")

	// 1. existing room
	req, _ := http.NewRequest("GET", "/api/v1/rooms/"+code+"/exists", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected OK, got %v", rr.Code)
	}

	// 2. non-existent room
	req, _ = http.NewRequest("GET", "/api/v1/rooms/invalid/exists", nil)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected Not Found, got %v", rr.Code)
	}
}

func TestAPI_HandleForceCloseRoom(t *testing.T) {
	logger := zap.NewNop()
	rm := NewRoomManager(time.Hour, logger)
	config := &Config{ObserverURL: "http://localhost:5173"}
	api := NewAPI(config, logger, rm)
	router := api.Router()

	code, writeToken, _, _ := rm.CreateRoom("runner_test", "")
	runner2, _, _, _ := rm.JoinRoom(code, "runner_test2", "", "")

	// 1. unauthorized user
	body, _ := json.Marshal(map[string]string{"writeToken": runner2})
	req, _ := http.NewRequest("POST", "/api/v1/rooms/"+code+"/force-close", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected Unauthorized, got %v", rr.Code)
	}

	// 2. authorized user (creator)
	body, _ = json.Marshal(map[string]string{"writeToken": writeToken})
	req, _ = http.NewRequest("POST", "/api/v1/rooms/"+code+"/force-close", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected OK, got %v", rr.Code)
	}

	// ensure the room was closed
	if rm.GetRoom(code) != nil {
		t.Errorf("expected room to be deleted")
	}
}
