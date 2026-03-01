package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
	ctxCancel      context.CancelFunc
}

func NewWSClient(serverURL, room, token string, tlsSkipVerify bool, heartbeatIntvl time.Duration, logger *zap.Logger) *WSClient {
	return &WSClient{
		url:            fmt.Sprintf("%s?room=%s&token=%s", serverURL, room, token),
		tlsSkipVerify:  tlsSkipVerify,
		heartbeatIntvl: heartbeatIntvl,
		logger:         logger,
	}
}

func (c *WSClient) Close() {
	if c.ctxCancel != nil {
		c.ctxCancel()
	}
}

func (c *WSClient) Run(ctx context.Context, in <-chan RoomMessage) {
	ctx, c.ctxCancel = context.WithCancel(ctx)
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0 // loop forever

	_ = backoff.RetryNotify(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
		}

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

		readCtx, readCancel := context.WithCancel(ctx)
		defer readCancel()

		go func() {
			for {
				// we must read from the connection so that control frames (like Pong) are correctly handled
				// by discarding the data, coder/websocket's internal frame parser completes the ping
				_, _, err := conn.Read(readCtx)
				if err != nil {
					readCancel()
					return
				}
			}
		}()

		c.logger.Info("connected to central websocket")
		b.Reset()

		heartbeat := time.NewTicker(c.heartbeatIntvl)
		defer heartbeat.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-readCtx.Done():
				return fmt.Errorf("read failed or connection closed")
			case <-heartbeat.C:
				if err := conn.Ping(ctx); err != nil {
					return err
				}
			case msg := <-in:
				if err := wsjson.Write(ctx, conn, msg); err != nil {
					return err
				}

				payloadBytes, _ := json.Marshal(msg.Payload)
				c.logger.Info("sent payload for " + msg.Type + " to central server")
				c.logger.Debug("payload: ", zap.String("payload", string(payloadBytes)))
			}
		}
	}, b, func(err error, t time.Duration) {
		c.logger.Warn("websocket connection lost, retrying", zap.Error(err), zap.Duration("delay", t))
	})
}
