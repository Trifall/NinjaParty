package main

import (
	"encoding/json"
)

type RoomMessage struct {
	Type       string          `json:"type"`
	Ts         int64           `json:"ts"`
	Source     string          `json:"source"`
	RunnerName string          `json:"runnerName"`
	Payload    json.RawMessage `json:"payload"`
}

type StrongholdPayload struct {
	ResultType     string            `json:"resultType"`
	PlayerPosition PlayerPosition    `json:"playerPosition"`
	EyeThrows      []EyeThrow        `json:"eyeThrows"`
	Predictions    []ChunkPrediction `json:"predictions,omitempty"`
}

type PresencePayload struct {
	Status    string `json:"status"` // "online" or "offline"
	Name      string `json:"name,omitempty"`
	IsCreator bool   `json:"isCreator"`
}

type RoomInfoPayload struct {
	CreatedAt int64 `json:"createdAt"`
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

type BoatPayload struct {
	BoatAngle *float64 `json:"boatAngle,omitempty"`
	BoatState string   `json:"boatState"`
}

type BlindPayload struct {
	IsBlindModeEnabled bool        `json:"isBlindModeEnabled"`
	HasDivine          bool        `json:"hasDivine"`
	BlindResult        interface{} `json:"blindResult"`
}

type DivinePayload struct {
	IsDivineModeEnabled bool        `json:"isDivineModeEnabled"`
	DivineResult        interface{} `json:"divineResult"`
}

type AllAdvancementsPayload struct {
	IsAllAdvancementsModeEnabled bool        `json:"isAllAdvancementsModeEnabled"`
	Spawn                        interface{} `json:"spawn,omitempty"`
	Outpost                      interface{} `json:"outpost,omitempty"`
	Monument                     interface{} `json:"monument,omitempty"`
	Stronghold                   interface{} `json:"stronghold,omitempty"`
	DeepDark                     interface{} `json:"deepDark,omitempty"`
	CityQuery                    interface{} `json:"cityQuery,omitempty"`
	ShulkerTransport             interface{} `json:"shulkerTransport,omitempty"`
	GeneralLocation              interface{} `json:"generalLocation,omitempty"`
}

type InformationMessagesPayload struct {
	InformationMessages []InformationMessage `json:"informationMessages,omitempty"`
}

type InformationMessage struct {
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

type RoomClosedPayload struct {
	Message string `json:"message"`
}
