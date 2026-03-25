package reporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

const discordEmbedColor = 0xFF6600 // orange

type discordPayload struct {
	Embeds []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title     string         `json:"title"`
	Color     int            `json:"color"`
	Fields    []discordField `json:"fields,omitempty"`
	Timestamp string         `json:"timestamp"`
	Footer    *discordFooter `json:"footer,omitempty"`
}

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordFooter struct {
	Text string `json:"text"`
}

// DiscordNotifier sends formatted CSP reports to a Discord webhook as rich embeds.
type DiscordNotifier struct {
	webhook *WebhookSender
	url     string
}

// NewDiscordNotifier creates a DiscordNotifier backed by the shared webhook sender.
func NewDiscordNotifier(url string, ws *WebhookSender) *DiscordNotifier {
	return &DiscordNotifier{webhook: ws, url: url}
}

func (d *DiscordNotifier) Name() string { return "discord" }

// Send formats and delivers the report to Discord as a rich embed.
func (d *DiscordNotifier) Send(ctx context.Context, rpt *report) error {
	payload := formatDiscordReport(rpt)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling discord payload: %w", err)
	}

	if err := d.webhook.Send(ctx, d.url, body); err != nil {
		return fmt.Errorf("discord: %w", err)
	}
	return nil
}

func formatDiscordReport(rpt *report) discordPayload {
	window := rpt.Until.Sub(rpt.Since).Truncate(time.Minute)

	var fields []discordField

	fields = append(fields, discordField{
		Name:   "Time Window",
		Value:  fmt.Sprintf("%s to %s (%s)", rpt.Since.Format(timeFormat), rpt.Until.Format(timeFormat), humanDuration(window)),
		Inline: false,
	})

	fields = append(fields, discordField{
		Name:   "Total Violations",
		Value:  fmt.Sprintf("**%d**", rpt.Total),
		Inline: true,
	})

	if top := topN(rpt.Directives, topCount); len(top) > 0 {
		fields = append(fields, discordField{
			Name:  "Top Violated Directives",
			Value: formatDiscordList(top, true),
		})
	}

	if top := topN(rpt.Origins, topCount); len(top) > 0 {
		fields = append(fields, discordField{
			Name:  "Top Blocked Origins",
			Value: formatDiscordList(top, true),
		})
	}

	if top := topN(rpt.Pages, topCount); len(top) > 0 {
		fields = append(fields, discordField{
			Name:  "Top Pages",
			Value: formatDiscordList(top, true),
		})
	}

	if top := topN(rpt.Browsers, topCount); len(top) > 0 {
		fields = append(fields, discordField{
			Name:  "Top Browsers",
			Value: formatDiscordList(top, false),
		})
	}

	if len(rpt.Dispositions) > 0 {
		fields = append(fields, discordField{
			Name:  "Disposition",
			Value: formatDiscordList(topN(rpt.Dispositions, topCount), false),
		})
	}

	if top := topN(rpt.SourceFiles, topCount); len(top) > 0 {
		fields = append(fields, discordField{
			Name:  "Top Source Files",
			Value: formatDiscordList(top, true),
		})
	}

	if len(rpt.Samples) > 0 {
		var b strings.Builder
		for _, s := range rpt.Samples {
			sample := s.Sample
			if len(sample) > 60 {
				sample = sample[:60] + "…"
			}
			loc := ""
			if s.SourceFile != "" {
				loc = fmt.Sprintf(" (%s", s.SourceFile)
				if s.Line > 0 {
					loc += fmt.Sprintf(":%d", s.Line)
					if s.Col > 0 {
						loc += fmt.Sprintf(":%d", s.Col)
					}
				}
				loc += ")"
			}
			fmt.Fprintf(&b, "• `%s` `%s`%s\n", s.Directive, sample, loc)
		}
		fields = append(fields, discordField{
			Name:  "Recent Samples",
			Value: b.String(),
		})
	}

	return discordPayload{
		Embeds: []discordEmbed{{
			Title:     "Vigil CSP Report",
			Color:     discordEmbedColor,
			Fields:    fields,
			Timestamp: rpt.Until.Format(time.RFC3339),
			Footer:    &discordFooter{Text: "Vigil CSP Monitor"},
		}},
	}
}

func formatDiscordList(entries []rankedEntry, code bool) string {
	var b strings.Builder
	for _, e := range entries {
		if code {
			fmt.Fprintf(&b, "• `%s` — %d\n", e.Key, e.Count)
		} else {
			fmt.Fprintf(&b, "• %s — %d\n", e.Key, e.Count)
		}
	}
	return b.String()
}
