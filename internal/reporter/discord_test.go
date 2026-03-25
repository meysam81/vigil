package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

func TestFormatDiscordReport(t *testing.T) {
	rpt := &report{
		Total: 42,
		Since: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Until: time.Date(2025, 1, 2, 1, 30, 0, 0, time.UTC),
		Directives: map[string]int{
			"script-src-elem": 30,
			"style-src-elem":  12,
		},
		Dispositions: map[string]int{
			"enforce": 35,
			"report":  7,
		},
		Origins: map[string]int{
			"cdn.evil.com": 25,
			"tracker.io":   17,
		},
		Pages: map[string]int{
			"/app":   35,
			"/login": 7,
		},
		Browsers: map[string]int{
			"Chrome":  30,
			"Firefox": 12,
		},
		SourceFiles: map[string]int{
			"https://example.com/main.js": 20,
		},
		Samples: []sampleEntry{
			{
				Directive:  "script-src-elem",
				Sample:     "alert('xss')",
				SourceFile: "https://example.com/main.js",
				Line:       42,
				Col:        10,
			},
		},
	}

	payload := formatDiscordReport(rpt)

	if len(payload.Embeds) != 1 {
		t.Fatalf("embeds count: want 1, got %d", len(payload.Embeds))
	}

	embed := payload.Embeds[0]

	if embed.Title != "Vigil CSP Report" {
		t.Errorf("title: want 'Vigil CSP Report', got %q", embed.Title)
	}
	if embed.Color != discordEmbedColor {
		t.Errorf("color: want %d, got %d", discordEmbedColor, embed.Color)
	}
	if embed.Footer == nil || embed.Footer.Text != "Vigil CSP Monitor" {
		t.Error("footer missing or incorrect")
	}
	if embed.Timestamp != "2025-01-02T01:30:00Z" {
		t.Errorf("timestamp: want 2025-01-02T01:30:00Z, got %s", embed.Timestamp)
	}

	// Check expected fields exist
	fieldNames := make(map[string]string)
	for _, f := range embed.Fields {
		fieldNames[f.Name] = f.Value
	}

	expectedFields := []string{
		"Time Window",
		"Total Violations",
		"Top Violated Directives",
		"Top Blocked Origins",
		"Top Pages",
		"Top Browsers",
		"Disposition",
		"Top Source Files",
		"Recent Samples",
	}
	for _, name := range expectedFields {
		if _, ok := fieldNames[name]; !ok {
			t.Errorf("missing field %q", name)
		}
	}

	// Verify field content
	if v := fieldNames["Total Violations"]; !strings.Contains(v, "42") {
		t.Errorf("Total Violations should contain 42, got: %s", v)
	}
	if v := fieldNames["Top Violated Directives"]; !strings.Contains(v, "script-src-elem") {
		t.Errorf("Top Violated Directives should contain script-src-elem, got: %s", v)
	}
	if v := fieldNames["Top Blocked Origins"]; !strings.Contains(v, "cdn.evil.com") {
		t.Errorf("Top Blocked Origins should contain cdn.evil.com, got: %s", v)
	}
	if v := fieldNames["Recent Samples"]; !strings.Contains(v, "alert('xss')") || !strings.Contains(v, ":42:10") {
		t.Errorf("Recent Samples should contain sample and location, got: %s", v)
	}
}

func TestFormatDiscordReport_EmptyReport(t *testing.T) {
	rpt := &report{
		Total:        0,
		Since:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Until:        time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		Directives:   map[string]int{},
		Dispositions: map[string]int{},
		Origins:      map[string]int{},
		Pages:        map[string]int{},
		Browsers:     map[string]int{},
		SourceFiles:  map[string]int{},
	}

	payload := formatDiscordReport(rpt)
	embed := payload.Embeds[0]

	// Should only have Time Window and Total Violations
	if len(embed.Fields) != 2 {
		t.Errorf("empty report fields: want 2, got %d", len(embed.Fields))
	}
}

func TestFormatDiscordList(t *testing.T) {
	entries := []rankedEntry{
		{"script-src", 30},
		{"style-src", 12},
	}

	got := formatDiscordList(entries, true)
	if !strings.Contains(got, "`script-src`") {
		t.Errorf("code formatting missing backticks: %s", got)
	}

	got = formatDiscordList(entries, false)
	if strings.Contains(got, "`") {
		t.Errorf("non-code should not have backticks: %s", got)
	}
}

func TestDiscordNotifier_Send(t *testing.T) {
	var receivedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		if _, err := r.Body.Read(buf); err != nil && err.Error() != "EOF" {
			t.Errorf("reading body: %v", err)
		}
		receivedBody = buf
		w.WriteHeader(http.StatusNoContent) // Discord returns 204 on success
	}))
	defer srv.Close()

	log := logger.NewLogger("error", true)
	cfg := &config.ReporterConfig{
		MaxRetries:    0,
		RetryMinDelay: time.Millisecond,
		RetryMaxDelay: time.Millisecond,
	}
	ws := NewWebhookSender(log, cfg)
	notifier := NewDiscordNotifier(srv.URL, ws)

	rpt := &report{
		Total:        5,
		Since:        time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		Until:        time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		Directives:   map[string]int{"script-src": 5},
		Dispositions: map[string]int{"enforce": 5},
		Origins:      map[string]int{"evil.com": 5},
		Pages:        map[string]int{"/": 5},
		Browsers:     map[string]int{"Chrome": 5},
		SourceFiles:  map[string]int{},
	}

	err := notifier.Send(context.Background(), rpt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the payload is valid Discord webhook JSON
	var payload discordPayload
	if err := json.Unmarshal(receivedBody, &payload); err != nil {
		t.Fatalf("received body is not valid JSON: %v", err)
	}
	if len(payload.Embeds) != 1 {
		t.Fatalf("embeds: want 1, got %d", len(payload.Embeds))
	}
	if payload.Embeds[0].Title != "Vigil CSP Report" {
		t.Errorf("title: want 'Vigil CSP Report', got %q", payload.Embeds[0].Title)
	}
}

func TestDiscordNotifier_Name(t *testing.T) {
	d := &DiscordNotifier{}
	if d.Name() != "discord" {
		t.Errorf("Name: want discord, got %s", d.Name())
	}
}
