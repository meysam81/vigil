package reporter

import (
	"testing"
)

func TestParseReport_Legacy(t *testing.T) {
	rpt := &report{
		Directives: make(map[string]int),
		Origins:    make(map[string]int),
		Pages:      make(map[string]int),
		Browsers:   make(map[string]int),
	}

	raw := `{
		"csp-report": {
			"blocked-uri": "http://example.com/css/style.css",
			"document-uri": "http://example.com/signup.html",
			"effective-directive": "style-src-elem",
			"violated-directive": "style-src-elem"
		}
	}`

	parseReport(raw, rpt)

	if rpt.Total != 1 {
		t.Fatalf("total: want 1, got %d", rpt.Total)
	}
	if rpt.Directives["style-src-elem"] != 1 {
		t.Fatalf("directive: want style-src-elem=1, got %v", rpt.Directives)
	}
	if rpt.Origins["example.com"] != 1 {
		t.Fatalf("origin: want example.com=1, got %v", rpt.Origins)
	}
	if rpt.Pages["/signup.html"] != 1 {
		t.Fatalf("page: want /signup.html=1, got %v", rpt.Pages)
	}
	if rpt.Browsers["Unknown"] != 1 {
		t.Fatalf("browser: want Unknown=1, got %v", rpt.Browsers)
	}
}

func TestParseReport_Modern(t *testing.T) {
	rpt := &report{
		Directives: make(map[string]int),
		Origins:    make(map[string]int),
		Pages:      make(map[string]int),
		Browsers:   make(map[string]int),
	}

	raw := `{
		"body": {
			"blockedURL": "https://cdn.evil.com/tracker.js",
			"documentURL": "https://example.com/app",
			"effectiveDirective": "script-src-elem"
		},
		"user_agent": "Mozilla/5.0 Chrome/127.0.0.0 Safari/537.36"
	}`

	parseReport(raw, rpt)

	if rpt.Total != 1 {
		t.Fatalf("total: want 1, got %d", rpt.Total)
	}
	if rpt.Directives["script-src-elem"] != 1 {
		t.Fatalf("directive: want script-src-elem=1, got %v", rpt.Directives)
	}
	if rpt.Origins["cdn.evil.com"] != 1 {
		t.Fatalf("origin: want cdn.evil.com=1, got %v", rpt.Origins)
	}
	if rpt.Browsers["Chrome"] != 1 {
		t.Fatalf("browser: want Chrome=1, got %v", rpt.Browsers)
	}
}

func TestExtractOrigin(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"http URL", "http://example.com/path", "example.com"},
		{"https URL", "https://cdn.example.com/script.js", "cdn.example.com"},
		{"data URI", "data:text/css;base64,abc", "data:"},
		{"blob URI", "blob:https://example.com/uuid", "blob:"},
		{"inline", "inline", "inline"},
		{"eval", "eval", "eval"},
		{"self", "self", "self"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOrigin(tt.in)
			if got != tt.want {
				t.Errorf("extractOrigin(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://example.com/page", "/page"},
		{"https://example.com/", "/"},
		{"https://example.com", "/"},
		{"https://example.com/a/b/c", "/a/b/c"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := extractPath(tt.in)
			if got != tt.want {
				t.Errorf("extractPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBrowserFamily(t *testing.T) {
	tests := []struct {
		ua   string
		want string
	}{
		{"Mozilla/5.0 Chrome/127.0 Safari/537.36", "Chrome"},
		{"Mozilla/5.0 Firefox/128.0", "Firefox"},
		{"Mozilla/5.0 Safari/605.1.15", "Safari"},
		{"Mozilla/5.0 Edg/127.0", "Edge"},
		{"Mozilla/5.0 OPR/113.0", "Opera"},
		{"SomeUnknownBot/1.0", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := browserFamily(tt.ua)
			if got != tt.want {
				t.Errorf("browserFamily(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func TestTopN(t *testing.T) {
	m := map[string]int{
		"a": 10,
		"b": 5,
		"c": 20,
		"d": 1,
		"e": 15,
	}

	top3 := topN(m, 3)
	if len(top3) != 3 {
		t.Fatalf("len: want 3, got %d", len(top3))
	}
	if top3[0].Key != "c" || top3[0].Count != 20 {
		t.Errorf("top3[0]: want c=20, got %s=%d", top3[0].Key, top3[0].Count)
	}
	if top3[1].Key != "e" || top3[1].Count != 15 {
		t.Errorf("top3[1]: want e=15, got %s=%d", top3[1].Key, top3[1].Count)
	}
	if top3[2].Key != "a" || top3[2].Count != 10 {
		t.Errorf("top3[2]: want a=10, got %s=%d", top3[2].Key, top3[2].Count)
	}

	// n > len(m)
	all := topN(m, 100)
	if len(all) != 5 {
		t.Fatalf("all len: want 5, got %d", len(all))
	}

	// empty map
	empty := topN(map[string]int{}, 5)
	if len(empty) != 0 {
		t.Fatalf("empty len: want 0, got %d", len(empty))
	}
}
