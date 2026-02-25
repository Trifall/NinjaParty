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

	relay := NewRelay(config, logger)
	if err := relay.StartRoom(roomRes.RoomCode, roomRes.WriteToken); err != nil {
		logger.Fatal("failed to start relay", zap.Error(err))
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
					clipboard.WriteAll(relay.RoomCode())
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
