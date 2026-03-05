package ui

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NinjaParty/relay/core"
)

func (u *UI) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	state, err := core.LoadRelayState()
	if err != nil {
		state = &core.RelayState{}
	}

	isActive := u.relay.RoomCode() != ""
	roomCode := u.relay.RoomCode()
	fullLink := u.relay.FullLink()

	showSetup := (state.RunnerName == "" || state.RoomCode == "") && !isActive
	staleRoom := false
	showRejoin := false

	if roomCode != "" && isActive {
		if !core.RoomExists(u.relay.ServerAPIURL(), roomCode) {
			staleRoom = true
			isActive = false
			u.relay.Shutdown()
			if err := core.ClearRelayState(); err != nil {
				u.logger.Warn("failed to clear relay state", zap.Error(err))
			}
		}
	} else if state.RoomCode != "" && !isActive && state.RunnerName != "" {
		if core.RoomExists(u.relay.ServerAPIURL(), state.RoomCode) {
			showRejoin = true
		} else {
			staleRoom = true
		}
	}

	data := PageData{
		RunnerName:    state.RunnerName,
		RoomCode:      roomCode,
		FullLink:      fullLink,
		IsActive:      isActive,
		IsCreator:     u.relay.IsCreator(),
		StaleRoom:     staleRoom,
		ShowSetup:     showSetup,
		ShowRejoin:    showRejoin,
		HasRunnerName: state.RunnerName != "",
		SavedRoomCode: state.RoomCode,
	}

	if err := u.tmpl.Execute(w, data); err != nil {
		u.logger.Error("failed to render ui template", zap.Error(err))
	}
}

func (u *UI) handleClose(w http.ResponseWriter, r *http.Request) {
	u.logger.Info("close requested via UI")
	u.relay.Shutdown()
}

func (u *UI) handleForceClose(w http.ResponseWriter, r *http.Request) {
	u.logger.Info("force close requested via UI")
	roomCode := u.relay.RoomCode()
	writeToken := ""
	if u.relay.IsCreator() {
		state, err := core.LoadRelayState()
		if err == nil {
			writeToken = state.WriteToken
		}
	}

	if writeToken != "" {
		reqBody, _ := json.Marshal(map[string]string{"writeToken": writeToken})
		resp, err := http.Post(u.relay.ServerAPIURL()+"/rooms/"+roomCode+"/force-close", "application/json", bytes.NewReader(reqBody))
		if err != nil || resp.StatusCode != http.StatusOK {
			u.logger.Warn("server force-close failed, closing locally anyway")
		}
	}

	u.relay.Shutdown()
	state, _ := core.LoadRelayState()
	if state != nil {
		state.RoomCode = ""
		state.WriteToken = ""
		state.Password = ""
		if err := core.SaveRelayState(state); err != nil {
			u.logger.Warn("failed to save relay state", zap.Error(err))
		}
	}
}

func (u *UI) handleCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encodeErr := func(status int, msg string) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(SetupResponse{Success: false, Error: msg}); err != nil {
			u.logger.Error("failed to encode error response", zap.Error(err))
		}
	}

	if !u.newRoomRate.Allow() {
		encodeErr(http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		encodeErr(http.StatusBadRequest, "invalid request")
		return
	}

	if req.RunnerName == "" {
		encodeErr(http.StatusBadRequest, "runner name is required")
		return
	}

	u.relay.SetRunnerName(req.RunnerName)

	roomRes, err := u.relay.CreateNewRoom(req.Password)
	if err != nil {
		u.logger.Error("failed to create room", zap.Error(err))
		encodeErr(http.StatusInternalServerError, "Failed to create room: "+err.Error())
		return
	}

	state := &core.RelayState{
		RunnerName: req.RunnerName,
		RoomCode:   roomRes.RoomCode,
		WriteToken: roomRes.WriteToken,
		Password:   req.Password,
	}
	if err := core.SaveRelayState(state); err != nil {
		u.logger.Warn("failed to save relay state", zap.Error(err))
	}

	if err := json.NewEncoder(w).Encode(SetupResponse{Success: true, RoomCode: roomRes.RoomCode}); err != nil {
		u.logger.Error("failed to encode create response", zap.Error(err))
	}
}

func (u *UI) handleJoin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encodeErr := func(status int, msg string) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(SetupResponse{Success: false, Error: msg}); err != nil {
			u.logger.Error("failed to encode error response", zap.Error(err))
		}
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		encodeErr(http.StatusBadRequest, "invalid request")
		return
	}

	if req.RunnerName == "" || req.RoomCode == "" {
		encodeErr(http.StatusBadRequest, "runner name and room code are required")
		return
	}

	u.relay.SetRunnerName(req.RunnerName)

	roomRes, err := u.relay.JoinRoom(req.RoomCode, req.Password, "")
	if err != nil {
		u.logger.Error("failed to join room", zap.Error(err))
		encodeErr(http.StatusBadRequest, err.Error())
		return
	}

	state := &core.RelayState{
		RunnerName: req.RunnerName,
		RoomCode:   roomRes.RoomCode,
		WriteToken: roomRes.WriteToken,
		Password:   req.Password,
	}
	if err := core.SaveRelayState(state); err != nil {
		u.logger.Warn("failed to save relay state", zap.Error(err))
	}

	if err := json.NewEncoder(w).Encode(SetupResponse{Success: true, RoomCode: roomRes.RoomCode}); err != nil {
		u.logger.Error("failed to encode join response", zap.Error(err))
	}
}

func (u *UI) handleRejoin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encodeErr := func(status int, msg string) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(SetupResponse{Success: false, Error: msg}); err != nil {
			u.logger.Error("failed to encode error response", zap.Error(err))
		}
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		encodeErr(http.StatusBadRequest, "invalid request")
		return
	}

	state, err := core.LoadRelayState()
	if err != nil || state.RunnerName == "" {
		encodeErr(http.StatusBadRequest, "no saved runner name found")
		return
	}

	roomRes, err := u.relay.JoinRoom(req.RoomCode, state.Password, state.WriteToken)
	if err != nil {
		u.logger.Error("failed to rejoin room", zap.Error(err))
		encodeErr(http.StatusBadRequest, err.Error())
		return
	}

	state.RoomCode = roomRes.RoomCode
	state.WriteToken = roomRes.WriteToken
	if err := core.SaveRelayState(state); err != nil {
		u.logger.Warn("failed to save relay state", zap.Error(err))
	}

	if err := json.NewEncoder(w).Encode(SetupResponse{Success: true, RoomCode: roomRes.RoomCode}); err != nil {
		u.logger.Error("failed to encode rejoin response", zap.Error(err))
	}
}

func (u *UI) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	isActive := u.relay.RoomCode() != ""
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"ninjabrainConnected": u.relay.NinjabrainConnected(),
		"roomCreatedAt":       u.relay.RoomCreatedAt().UnixMilli(),
		"isActive":            isActive,
	}); err != nil {
		u.logger.Error("failed to encode status response", zap.Error(err))
	}
}
