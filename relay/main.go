package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/browser"
	"go.uber.org/zap"

	"github.com/NinjaParty/relay/core"
	"github.com/NinjaParty/relay/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var logger *zap.Logger
	if os.Getenv("DEBUG") == "1" {
		logger, _ = zap.NewDevelopmentConfig().Build()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting NinjaParty Relay",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("date", date),
	)

	config, err := core.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	relay := core.NewRelay(config, logger)

	state, err := core.LoadRelayState()
	if err != nil {
		logger.Warn("failed to load relay state", zap.Error(err))
		state = &core.RelayState{}
	}

	if state.RunnerName != "" {
		relay.SetRunnerName(state.RunnerName)
	}

	// auto-rejoin if we have a saved room that still exists
	if state.RoomCode != "" && state.RunnerName != "" {
		if core.RoomExists(config.ServerAPIURL, state.RoomCode) {
			logger.Info("attempting to auto-rejoin saved room", zap.String("code", state.RoomCode))
			roomRes, err := relay.JoinRoom(state.RoomCode, state.Password, state.WriteToken)
			if err != nil {
				logger.Warn("failed to auto-rejoin room", zap.Error(err))
				state.RoomCode = ""
				state.WriteToken = ""
				if err := core.SaveRelayState(state); err != nil {
					logger.Warn("failed to save relay state", zap.Error(err))
				}
			} else {
				logger.Info("successfully auto-rejoined room", zap.String("code", roomRes.RoomCode))
				state.RoomCode = roomRes.RoomCode
				state.WriteToken = roomRes.WriteToken
				if err := core.SaveRelayState(state); err != nil {
					logger.Warn("failed to save relay state", zap.Error(err))
				}
			}
		} else {
			logger.Info("saved room no longer exists", zap.String("code", state.RoomCode))
			state.RoomCode = ""
			state.WriteToken = ""
			if err := core.SaveRelayState(state); err != nil {
				logger.Warn("failed to save relay state", zap.Error(err))
			}
		}
	}

	uiInstance := ui.New(config.LocalUIAddr, relay, logger, 5*time.Second)
	go func() {
		if err := uiInstance.Start(); err != nil {
			logger.Fatal("ui server failed", zap.Error(err))
		}
	}()

	if err := browser.OpenURL(fmt.Sprintf("http://%s", config.LocalUIAddr)); err != nil {
		logger.Warn("failed to open browser", zap.Error(err))
	}

	go runTray(relay, config.LocalUIAddr, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutting down relay")
	relay.Shutdown()
}
