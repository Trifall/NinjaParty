package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var logger *zap.Logger
	if os.Getenv("DEBUG") == "1" {
		logger, _ = zap.NewDevelopmentConfig().Build(zap.AddStacktrace(zapcore.FatalLevel))
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("Starting NinjaParty Server",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("date", date),
	)

	config, err := LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	rm := NewRoomManager(config.RoomTTL, logger)
	api := NewAPI(config, logger, rm)

	srv := &http.Server{
		Addr:    config.ListenAddr,
		Handler: api.Router(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go rm.StartEvictionTicker(ctx, 5*time.Minute) // check for idle rooms every 5 min

	logger.Info("starting server", zap.String("addr", config.ListenAddr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exiting")
}
