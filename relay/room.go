package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func createRoom(serverAPI string) (*RoomResponse, error) {
	resp, err := http.Post(serverAPI+"/rooms", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var parsed RoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if parsed.RoomCode == "" || parsed.WriteToken == "" {
		return nil, fmt.Errorf("invalid response: missing roomCode or writeToken")
	}

	if len(parsed.RoomCode) < 3 || len(parsed.WriteToken) < 16 {
		return nil, fmt.Errorf("invalid response: roomCode or writeToken too short")
	}

	return &parsed, nil
}
