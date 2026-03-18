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
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

const (
	topCount   = 5
	timeFormat = time.RFC3339
)

type slackPayload struct {
	Text string `json:"text"`
}

// SlackNotifier sends formatted CSP reports to a Slack webhook.
type SlackNotifier struct {
	cfg    *config.SlackConfig
	log    *logger.Logger
	client *http.Client
}

// NewSlackNotifier creates a SlackNotifier with a dedicated HTTP client.
func NewSlackNotifier(cfg *config.SlackConfig, log *logger.Logger) *SlackNotifier {
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
	return &SlackNotifier{
		cfg: cfg,
		log: log,
		client: &http.Client{
			Transport: transport,
		},
	}
}

func (s *SlackNotifier) Name() string { return "slack" }

// Send formats and delivers the report to Slack with retry and jitter.
func (s *SlackNotifier) Send(ctx context.Context, rpt *report) error {
	s.log.Debug().Int("violations", rpt.Total).Msg("preparing slack report")

	msg := formatReport(rpt)

	body, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("marshaling slack payload: %w", err)
	}

	var lastErr error
	attempts := 1 + s.cfg.MaxRetries
	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = s.post(ctx, body)
		if lastErr == nil {
			s.log.Info().Int("violations", rpt.Total).Msg("slack report delivered")
			return nil
		}

		if i < attempts-1 {
			delay := jitterDelay(s.cfg.RetryMinDelay, s.cfg.RetryMaxDelay)
			s.log.Warn().Err(lastErr).Int("attempt", i+1).Dur("retry_in", delay).Msg("slack post failed, retrying")

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return fmt.Errorf("slack delivery failed after %d attempts: %w", attempts, lastErr)
}

func (s *SlackNotifier) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("posting to slack: %w", err)
	}
	defer func() {
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			s.log.Error().Err(err).Msg("failed discarding http response")
		}
		err = resp.Body.Close()
		if err != nil {
			s.log.Error().Err(err).Msg("failed closing response body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
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

func formatReport(rpt *report) string {
	var b bytes.Buffer

	window := rpt.Until.Sub(rpt.Since).Truncate(time.Minute)
	fmt.Fprintf(&b, "*Vigil CSP Report* — %s to %s (%s)\n", rpt.Since.Format(timeFormat), rpt.Until.Format(timeFormat), humanDuration(window))
	fmt.Fprintf(&b, "Total violations: *%d*\n", rpt.Total)

	if top := topN(rpt.Directives, topCount); len(top) > 0 {
		b.WriteString("\n*Top Violated Directives:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if top := topN(rpt.Origins, topCount); len(top) > 0 {
		b.WriteString("\n*Top Blocked Origins:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if top := topN(rpt.Pages, topCount); len(top) > 0 {
		b.WriteString("\n*Top Pages:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if top := topN(rpt.Browsers, topCount); len(top) > 0 {
		b.WriteString("\n*Top Browsers:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • %s — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if len(rpt.Dispositions) > 0 {
		b.WriteString("\n*Disposition:*\n")
		for _, e := range topN(rpt.Dispositions, topCount) {
			fmt.Fprintf(&b, "  • %s — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if top := topN(rpt.SourceFiles, topCount); len(top) > 0 {
		b.WriteString("\n*Top Source Files:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", escapeMrkdwn(e.Key), e.Count)
		}
	}

	if len(rpt.Samples) > 0 {
		b.WriteString("\n*Recent Samples:*\n")
		for _, s := range rpt.Samples {
			sample := s.Sample
			if len(sample) > 60 {
				sample = sample[:60] + "…"
			}
			loc := ""
			if s.SourceFile != "" {
				loc = fmt.Sprintf(" (%s", escapeMrkdwn(s.SourceFile))
				if s.Line > 0 {
					loc += fmt.Sprintf(":%d", s.Line)
					if s.Col > 0 {
						loc += fmt.Sprintf(":%d", s.Col)
					}
				}
				loc += ")"
			}
			fmt.Fprintf(&b, "  • `%s` `%s`%s\n", escapeMrkdwn(s.Directive), escapeMrkdwn(sample), loc)
		}
	}

	return b.String()
}

// escapeMrkdwn escapes Slack mrkdwn special characters in user-controlled text.
func escapeMrkdwn(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// humanDuration formats a duration as "Xh Ym" for readability.
func humanDuration(d time.Duration) string {
	d = d.Truncate(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
