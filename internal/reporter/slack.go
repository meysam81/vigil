package reporter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

const (
	slackTimeout = 10 * time.Second
	topCount     = 5
	timeFormat   = time.RFC1123
)

type slackPayload struct {
	Text string `json:"text"`
}

func (r *Reporter) sendSlack(ctx context.Context, report *Report) error {
	msg := formatReport(report)

	body, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("marshaling slack payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, slackTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to slack: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			r.log.Error().Err(err).Msg("failed closing response body")
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	return nil
}

func formatReport(report *Report) string {
	var b bytes.Buffer

	window := report.Until.Sub(report.Since).Truncate(time.Minute)
	fmt.Fprintf(&b, "*Vigil CSP Report* — %s to %s (%s)\n", report.Since.Format(timeFormat), report.Until.Format(timeFormat), window)
	fmt.Fprintf(&b, "Total violations: *%d*\n", report.Total)

	if top := TopN(report.Directives, topCount); len(top) > 0 {
		b.WriteString("\n*Top Violated Directives:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", e.Key, e.Count)
		}
	}

	if top := TopN(report.Origins, topCount); len(top) > 0 {
		b.WriteString("\n*Top Blocked Origins:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", e.Key, e.Count)
		}
	}

	if top := TopN(report.Pages, topCount); len(top) > 0 {
		b.WriteString("\n*Top Pages:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • `%s` — %d\n", e.Key, e.Count)
		}
	}

	if top := TopN(report.Browsers, topCount); len(top) > 0 {
		b.WriteString("\n*Top Browsers:*\n")
		for _, e := range top {
			fmt.Fprintf(&b, "  • %s — %d\n", e.Key, e.Count)
		}
	}

	return b.String()
}
