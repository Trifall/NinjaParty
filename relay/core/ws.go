package core

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
	onRoomClosed   func()
}

func NewWSClient(serverURL, room, runnerName, token string, tlsSkipVerify bool, heartbeatIntvl time.Duration, logger *zap.Logger, onRoomClosed func()) *WSClient {
	return &WSClient{
		url:            fmt.Sprintf("%s?room=%s&runnerName=%s&token=%s", serverURL, room, runnerName, token),
		tlsSkipVerify:  tlsSkipVerify,
		heartbeatIntvl: heartbeatIntvl,
		logger:         logger,
		onRoomClosed:   onRoomClosed,
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

		conn, resp, err := websocket.Dial(ctx, c.url, &websocket.DialOptions{
			HTTPClient: client,
		})
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				c.logger.Info("websocket dial returned 404, room is closed")
				if c.onRoomClosed != nil {
					c.onRoomClosed()
				}
				return backoff.Permanent(err)
			}
			return err
		}
		defer func() {
			if err := conn.CloseNow(); err != nil {
				c.logger.Debug("websocket CloseNow error", zap.Error(err))
			}
		}()

		readCtx, readCancel := context.WithCancel(ctx)
		defer readCancel()

		go func() {
			for {
				// we must read from the connection so that control frames (like Pong) are correctly handled
				// by discarding the data, coder/websocket's internal frame parser completes the ping
				_, _, err := conn.Read(readCtx)
				if err != nil {
					if websocket.CloseStatus(err) == websocket.StatusCode(4000) {
						c.logger.Info("received room closed close frame")
						if c.onRoomClosed != nil {
							c.onRoomClosed()
						}
					}
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
