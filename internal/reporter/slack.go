package reporter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"
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
	webhook *WebhookSender
	url     string
}

// NewSlackNotifier creates a SlackNotifier backed by the shared webhook sender.
func NewSlackNotifier(url string, ws *WebhookSender) *SlackNotifier {
	return &SlackNotifier{webhook: ws, url: url}
}

func (s *SlackNotifier) Name() string { return "slack" }

// Send formats and delivers the report to Slack.
func (s *SlackNotifier) Send(ctx context.Context, rpt *report) error {
	msg := formatSlackReport(rpt)

	body, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("marshaling slack payload: %w", err)
	}

	if err := s.webhook.Send(ctx, s.url, body); err != nil {
		return fmt.Errorf("slack: %w", err)
	}
	return nil
}

func formatSlackReport(rpt *report) string {
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
