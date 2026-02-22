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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	messageChan := make(chan RoomMessage, 100)

	sseClient := NewSSEClient(config.NinjabrainURL+"/api/v1/stronghold/events", logger)
	go sseClient.Run(ctx, messageChan)

	wsClient := NewWSClient(config.ServerWSURL, roomRes.RoomCode, roomRes.WriteToken, config.TLSSkipVerify, time.Duration(config.HeartbeatIntvl)*time.Second, logger)
	go wsClient.Run(ctx, messageChan)

	<-ctx.Done()
	logger.Info("shutting down relay")
}
