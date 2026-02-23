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
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("fetching room from server", zap.String("api", config.ServerAPIURL))
	roomRes, err := createRoom(config.ServerAPIURL)
	if err != nil {
		logger.Fatal("could not create room", zap.Error(err))
	}

	logger.Info("room created", zap.String("code", roomRes.RoomCode))

	ui := &UI{
		addr:        config.LocalUIAddr,
		observerURL: config.ObserverURL,
		room:        roomRes.RoomCode,
		logger:      logger,
	}
	go func() {
		if err := ui.Start(); err != nil {
			logger.Fatal("ui server failed", zap.Error(err))
		}
	}()

	// automatically open browser to the local UI
	browser.OpenURL(fmt.Sprintf("http://%s", config.LocalUIAddr))

	// Setup system tray in a goroutine
	go func() {
		systray.Run(func() {
			systray.SetIcon(iconData)
			systray.SetTitle("Ninjabrain Party Relay")
			systray.SetTooltip("Ninjabrain Party Relay")

			mTitle := systray.AddMenuItem("Ninjabrain Party Relay", "")
			mTitle.Disable()
			systray.AddSeparator()
			mOpen := systray.AddMenuItem("Open UI", "Open the relay UI")
			mCopyCode := systray.AddMenuItem("Copy Room Code", "Copy the room code to clipboard")
			mQuit := systray.AddMenuItem("Quit", "Close the relay")
			mQuit.SetTooltip("Quit the relay application")

			for {
				select {
				case <-mOpen.ClickedCh:
					browser.OpenURL(fmt.Sprintf("http://%s", config.LocalUIAddr))
				case <-mCopyCode.ClickedCh:
					clipboard.WriteAll(roomRes.RoomCode)
				case <-mQuit.ClickedCh:
					logger.Info("shutting down relay via tray")
					os.Exit(0)
				}
			}
		}, func() {
			logger.Info("systray exited")
		})
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	messageChan := make(chan RoomMessage, 100)

	endpoints := map[string]string{
		"/api/v1/stronghold/events":       "stronghold_update",
		"/api/v1/boat/events":             "boat_update",
		"/api/v1/blind/events":            "blind_update",
		"/api/v1/divine/events":           "divine_update",
		"/api/v1/all-advancements/events": "all_advancements_update",
		// "/api/v1/information-messages/events": "information_messages_update", // this is not available in v1.5.1 API
	}

	for endpoint, evType := range endpoints {
		sseClient := NewSSEClient(config.NinjabrainURL+endpoint, evType, logger)
		go sseClient.Run(ctx, messageChan)
	}

	wsClient := NewWSClient(config.ServerWSURL, roomRes.RoomCode, roomRes.WriteToken, config.TLSSkipVerify, time.Duration(config.HeartbeatIntvl)*time.Second, logger)
	go wsClient.Run(ctx, messageChan)

	<-ctx.Done()
	logger.Info("shutting down relay")
}
