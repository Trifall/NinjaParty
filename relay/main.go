package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/pkg/browser"
	"go.uber.org/zap"
)

func main() {
	var logger *zap.Logger
	if os.Getenv("DEBUG") == "1" {
		logger, _ = zap.NewDevelopmentConfig().Build()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	relay := NewRelay(config, logger)

	state, err := loadRelayState()
	if err != nil {
		logger.Warn("failed to load relay state", zap.Error(err))
		state = &RelayState{}
	}

	if state.RunnerName != "" {
		relay.SetRunnerName(state.RunnerName)
	}

	// auto-rejoin if we have a saved room that still exists
	if state.RoomCode != "" && state.RunnerName != "" {
		if roomExists(config.ServerAPIURL, state.RoomCode) {
			logger.Info("attempting to auto-rejoin saved room", zap.String("code", state.RoomCode))
			roomRes, err := relay.JoinRoom(state.RoomCode, state.Password, state.WriteToken)
			if err != nil {
				logger.Warn("failed to auto-rejoin room", zap.Error(err))
				// clear the saved room code since we couldn't rejoin
				state.RoomCode = ""
				state.WriteToken = ""
				saveRelayState(state)
			} else {
				logger.Info("successfully auto-rejoined room", zap.String("code", roomRes.RoomCode))
				// update state with new tokens
				state.RoomCode = roomRes.RoomCode
				state.WriteToken = roomRes.WriteToken
				saveRelayState(state)
			}
		} else {
			logger.Info("saved room no longer exists", zap.String("code", state.RoomCode))
			state.RoomCode = ""
			state.WriteToken = ""
			saveRelayState(state)
		}
	}

	ui := &UI{
		addr:        config.LocalUIAddr,
		relay:       relay,
		logger:      logger,
		newRoomRate: NewRateLimiter(5 * time.Second),
	}
	go func() {
		if err := ui.Start(); err != nil {
			logger.Fatal("ui server failed", zap.Error(err))
		}
	}()

	browser.OpenURL(fmt.Sprintf("http://%s", config.LocalUIAddr))

	go func() {
		systray.Run(func() {
			systray.SetIcon(iconData)
			systray.SetTitle("NinjaParty Relay")
			systray.SetTooltip("NinjaParty Relay")

			mTitle := systray.AddMenuItem("NinjaParty Relay", "")
			mTitle.Disable()
			systray.AddSeparator()
			mOpen := systray.AddMenuItem("Open UI", "Open the relay UI")
			mCopyCode := systray.AddMenuItem("Copy Room Code", "Copy the room code to clipboard")
			mCopyLink := systray.AddMenuItem("Copy Link", "Copy the room link to clipboard")
			mQuit := systray.AddMenuItem("Quit", "Close the relay")
			mQuit.SetTooltip("Quit the relay application")

			for {
				select {
				case <-mOpen.ClickedCh:
					browser.OpenURL(fmt.Sprintf("http://%s", config.LocalUIAddr))
				case <-mCopyCode.ClickedCh:
					clipboard.WriteAll(relay.RoomCode())
				case <-mCopyLink.ClickedCh:
					clipboard.WriteAll(relay.FullLink())
				case <-mQuit.ClickedCh:
					logger.Info("shutting down relay via tray")
					relay.Shutdown()
					os.Exit(0)
				}
			}
		}, func() {
			logger.Info("systray exited")
		})
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutting down relay")
	relay.Shutdown()
}
