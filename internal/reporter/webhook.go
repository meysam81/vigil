package reporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

// WebhookSender delivers JSON payloads to webhook URLs with retry and jitter.
type WebhookSender struct {
	client *http.Client
	log    *logger.Logger
	cfg    *config.ReporterConfig
}

// NewWebhookSender creates a WebhookSender with a TLS-hardened HTTP client.
func NewWebhookSender(log *logger.Logger, cfg *config.ReporterConfig) *WebhookSender {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
	}
	return &WebhookSender{
		client: &http.Client{Transport: transport},
		log:    log,
		cfg:    cfg,
	}
}

// Send posts body to url with retry and jitter on failure.
func (w *WebhookSender) Send(ctx context.Context, url string, body []byte) error {
	var lastErr error
	attempts := 1 + w.cfg.MaxRetries
	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = w.post(ctx, url, body)
		if lastErr == nil {
			return nil
		}

		if i < attempts-1 {
			delay := jitterDelay(w.cfg.RetryMinDelay, w.cfg.RetryMaxDelay)
			w.log.Warn().Err(lastErr).Int("attempt", i+1).Dur("retry_in", delay).Msg("webhook post failed, retrying")

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return fmt.Errorf("webhook delivery failed after %d attempts: %w", attempts, lastErr)
}

func (w *WebhookSender) post(ctx context.Context, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("posting to webhook: %w", err)
	}
	defer func() {
		if _, discardErr := io.Copy(io.Discard, resp.Body); discardErr != nil {
			w.log.Error().Err(discardErr).Msg("failed discarding http response")
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			w.log.Error().Err(closeErr).Msg("failed closing response body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// jitterDelay returns a random duration between min and max.
func jitterDelay(min, max time.Duration) time.Duration {
	minMs := min.Milliseconds()
	maxMs := max.Milliseconds()
	if maxMs <= minMs {
		return min
	}
	return time.Duration(rand.Int64N(maxMs-minMs+1)+minMs) * time.Millisecond
}
