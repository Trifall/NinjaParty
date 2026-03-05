package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CreateRoom(serverAPI, runnerName, password string) (*RoomResponse, error) {
	reqBody := JoinRoomRequest{
		RunnerName: runnerName,
		Password:   password,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverAPI+"/rooms", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

func JoinRoom(serverAPI, roomCode, runnerName, password, prevToken string) (*RoomResponse, error) {
	reqBody := JoinRoomRequest{
		RunnerName: runnerName,
		Password:   password,
		PrevToken:  prevToken,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverAPI+"/rooms/"+roomCode+"/join", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// read plain text error message first
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil && len(bodyBytes) > 0 {
			errorMsg := string(bodyBytes)
			var errResp JoinRoomResponse
			if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error != "" {
				return nil, fmt.Errorf("%s", errResp.Error)
			}
			return nil, fmt.Errorf("%s", errorMsg)
		}
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var parsed JoinRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if parsed.WriteToken == "" {
		return nil, fmt.Errorf("invalid response: missing writeToken")
	}

	return &RoomResponse{
		RoomCode:   roomCode,
		WriteToken: parsed.WriteToken,
		CreatedAt:  parsed.CreatedAt,
		IsCreator:  parsed.IsCreator,
	}, nil
}

func RoomExists(serverAPI, roomCode string) bool {
	resp, err := http.Get(serverAPI + "/rooms/" + roomCode + "/exists")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}
