package main

import "encoding/json"

type RoomMessage struct {
	Type    string          `json:"type"`
	Ts      int64           `json:"ts"`
	Source  string          `json:"source"`
	Payload json.RawMessage `json:"payload"`
}

type StrongholdPayload struct {
	ResultType     string            `json:"resultType"`
	PlayerPosition PlayerPosition    `json:"playerPosition"`
	EyeThrows      []EyeThrow        `json:"eyeThrows"`
	Predictions    []ChunkPrediction `json:"predictions,omitempty"`
}

type PlayerPosition struct {
	XInOverworld    float64 `json:"xInOverworld"`
	ZInOverworld    float64 `json:"zInOverworld"`
	IsInOverworld   bool    `json:"isInOverworld"`
	IsInNether      bool    `json:"isInNether"`
	HorizontalAngle float64 `json:"horizontalAngle"`
}

type EyeThrow struct {
	XInOverworld           float64 `json:"xInOverworld"`
	ZInOverworld           float64 `json:"zInOverworld"`
	Angle                  float64 `json:"angle"`
	AngleWithoutCorrection float64 `json:"angleWithoutCorrection,omitempty"`
	Correction             float64 `json:"correction,omitempty"`
	CorrectionIncrements   int     `json:"correctionIncrements,omitempty"`
	Error                  float64 `json:"error"`
	Type                   string  `json:"type"`
}

type ChunkPrediction struct {
	Certainty         float64 `json:"certainty"`
	ChunkX            int     `json:"chunkX"`
	ChunkZ            int     `json:"chunkZ"`
	OverworldDistance float64 `json:"overworldDistance"`
}

type RoomResponse struct {
	RoomCode   string `json:"roomCode"`
	WriteToken string `json:"writeToken"`
}
