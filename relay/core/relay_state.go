package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const relayStateFile = "relay_state.json"

func LoadRelayState() (*RelayState, error) {
	data, err := os.ReadFile(relayStateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &RelayState{}, nil
		}
		return nil, err
	}
	var state RelayState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveRelayState(state *RelayState) error {
	dir := filepath.Dir(relayStateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(relayStateFile, data, 0600)
}

func ClearRelayState() error {
	if _, err := os.Stat(relayStateFile); err == nil {
		return os.Remove(relayStateFile)
	}
	return nil
}
