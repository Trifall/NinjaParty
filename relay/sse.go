package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type SSEClient struct {
	url    string
	logger *zap.Logger
}

func NewSSEClient(url string, logger *zap.Logger) *SSEClient {
	return &SSEClient{
		url:    url,
		logger: logger,
	}
}

func (c *SSEClient) Run(ctx context.Context, out chan<- RoomMessage) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0 // loop forever

	_ = backoff.RetryNotify(func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		c.logger.Info("connected to ninjabrain SSE")
		b.Reset()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if data == "" {
					continue
				}

				// validate payload and package into RoomMessage
				var payload json.RawMessage
				if err := json.Unmarshal([]byte(data), &payload); err != nil {
					c.logger.Error("failed to decode sse payload", zap.Error(err), zap.String("data", data))
					continue
				}

				msg := RoomMessage{
					Type:    "stronghold_update",
					Ts:      time.Now().UnixMilli(),
					Source:  "ninjabrain-local",
					Payload: payload,
				}

				select {
				case out <- msg:
				case <-ctx.Done():
					return nil
				default:
					c.logger.Warn("channel full, dropping sse message")
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	}, b, func(err error, t time.Duration) {
		c.logger.Warn("ninjabrain connection lost, retrying", zap.Error(err), zap.Duration("delay", t))
	})
}
