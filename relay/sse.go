package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

type SSEClient struct {
	url            string
	msgType        string
	runnerName     string
	tlsSkipVerify  bool
	logger         *zap.Logger
	ctxCancel      context.CancelFunc
	onConnected    func()
	onDisconnected func()
}

func NewSSEClient(url, msgType string, tlsSkipVerify bool, logger *zap.Logger) *SSEClient {
	return &SSEClient{
		url:           url,
		msgType:       msgType,
		tlsSkipVerify: tlsSkipVerify,
		logger:        logger,
	}
}

func (c *SSEClient) SetCallbacks(onConnected, onDisconnected func()) {
	c.onConnected = onConnected
	c.onDisconnected = onDisconnected
}

func (c *SSEClient) fetchInitialState(ctx context.Context, defaultEndpoint string, out chan<- RoomMessage) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.tlsSkipVerify},
		},
	}

	retries := 10
	retryInterval := 10 * time.Second

	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, defaultEndpoint, nil)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			c.logger.Warn("failed to fetch initial state, retrying", zap.Error(err), zap.Int("attempt", i+1))
			time.Sleep(retryInterval)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var payload json.RawMessage
			if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
				msg := RoomMessage{
					Type:       c.msgType,
					Ts:         time.Now().UnixMilli(),
					Source:     "ninjabrain-local",
					RunnerName: c.runnerName,
					Payload:    payload,
				}
				select {
				case out <- msg:
					c.logger.Info("fetched initial state from Ninjabrain")
					c.logger.Debug("initial state value: ", zap.String("type", c.msgType))
				case <-ctx.Done():
				default:
				}
				return
			} else {
				c.logger.Warn("failed to decode initial state response, retrying", zap.Error(err), zap.Int("attempt", i+1))
				time.Sleep(retryInterval)
				continue
			}
		}

		c.logger.Warn("failed to fetch initial state, retrying", zap.Int("status", resp.StatusCode), zap.Int("attempt", i+1))
		time.Sleep(retryInterval)
	}

	c.logger.Warn("failed to fetch initial state after all retries", zap.String("type", c.msgType))
}

func (c *SSEClient) Close() {
	if c.ctxCancel != nil {
		c.ctxCancel()
	}
}

func (c *SSEClient) Run(ctx context.Context, out chan<- RoomMessage, runnerName string) {
	c.runnerName = runnerName
	ctx, c.ctxCancel = context.WithCancel(ctx)
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0 // loop forever

	_ = backoff.RetryNotify(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "text/event-stream")

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.tlsSkipVerify},
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		c.logger.Info("connected to ninjabrain SSE")
		if c.onConnected != nil {
			c.onConnected()
		}
		b.Reset()

		if c.msgType != "information_messages_update" {
			baseURL := strings.TrimSuffix(c.url, "/events")
			c.fetchInitialState(ctx, baseURL, out)
		}

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

				c.logger.Debug("received ninjabrain sse payload")

				msg := RoomMessage{
					Type:       c.msgType,
					Ts:         time.Now().UnixMilli(),
					Source:     "ninjabrain-local",
					RunnerName: c.runnerName,
					Payload:    payload,
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
		if c.onDisconnected != nil {
			c.onDisconnected()
		}
	})
}
