package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"go.uber.org/zap"
)

type WSClient struct {
	url            string
	tlsSkipVerify  bool
	heartbeatIntvl time.Duration
	logger         *zap.Logger
}

func NewWSClient(serverURL, room, token string, tlsSkipVerify bool, heartbeatIntvl time.Duration, logger *zap.Logger) *WSClient {
	return &WSClient{
		url:            fmt.Sprintf("%s?room=%s&token=%s", serverURL, room, token),
		tlsSkipVerify:  tlsSkipVerify,
		heartbeatIntvl: heartbeatIntvl,
		logger:         logger,
	}
}

func (c *WSClient) Run(ctx context.Context, in <-chan RoomMessage) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0 // loop forever

	_ = backoff.RetryNotify(func() error {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.tlsSkipVerify},
			},
		}

		conn, _, err := websocket.Dial(ctx, c.url, &websocket.DialOptions{
			HTTPClient: client,
		})
		if err != nil {
			return err
		}
		defer conn.CloseNow()

		c.logger.Info("connected to central websocket")
		b.Reset()

		heartbeat := time.NewTicker(c.heartbeatIntvl)
		defer heartbeat.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-heartbeat.C:
				if err := conn.Ping(ctx); err != nil {
					return err
				}
			case msg := <-in:
				if err := wsjson.Write(ctx, conn, msg); err != nil {
					return err
				}
			}
		}
	}, b, func(err error, t time.Duration) {
		c.logger.Warn("websocket connection lost, retrying", zap.Error(err), zap.Duration("delay", t))
	})
}
