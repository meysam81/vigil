package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

func TestFormatSlackReport(t *testing.T) {
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

	msg := formatSlackReport(rpt)

	checks := []string{
		"*Vigil CSP Report*",
		"Total violations: *42*",
		"25h 30m",
		"script-src-elem",
		"cdn.evil.com",
		"/app",
		"Chrome",
		"*Disposition:*",
		"enforce",
		"report",
		"*Top Source Files:*",
		"example.com/main.js",
		"*Recent Samples:*",
		"alert('xss')",
		":42:10",
	}
	for _, c := range checks {
		if !strings.Contains(msg, c) {
			t.Errorf("formatSlackReport missing %q in:\n%s", c, msg)
		}
	}
}

func TestEscapeMrkdwn(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"<@U123>", "&lt;@U123&gt;"},
		{"foo & bar", "foo &amp; bar"},
		{"no special chars", "no special chars"},
		{"<script>", "&lt;script&gt;"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := escapeMrkdwn(tt.in)
			if got != tt.want {
				t.Errorf("escapeMrkdwn(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{25*time.Hour + 3*time.Minute, "25h 3m"},
		{45 * time.Minute, "45m"},
		{0, "0m"},
		{24 * time.Hour, "24h 0m"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := humanDuration(tt.d)
			if got != tt.want {
				t.Errorf("humanDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestSlackNotifier_Send(t *testing.T) {
	var received bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: want application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	log := logger.NewLogger("error", true)
	cfg := &config.ReporterConfig{
		MaxRetries:    0,
		RetryMinDelay: time.Millisecond,
		RetryMaxDelay: time.Millisecond,
	}
	ws := NewWebhookSender(log, cfg)
	notifier := NewSlackNotifier(srv.URL, ws)

	rpt := &report{
		Total:      1,
		Since:      time.Now().Add(-time.Hour),
		Until:      time.Now(),
		Directives: map[string]int{"script-src": 1},
		Origins:    map[string]int{"example.com": 1},
		Pages:      map[string]int{"/": 1},
		Browsers:   map[string]int{"Chrome": 1},
	}

	err := notifier.Send(context.Background(), rpt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !received {
		t.Fatal("server did not receive the request")
	}
}
