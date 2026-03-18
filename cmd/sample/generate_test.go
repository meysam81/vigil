package sample

import (
	"testing"
)

func TestGenerateLegacyDeterministic(t *testing.T) {
	report := GenerateLegacy(false)

	cspReport, ok := report["csp-report"].(map[string]any)
	if !ok {
		t.Fatal("missing csp-report key")
	}

	checks := map[string]any{
		"blocked-uri":         "https://cdn.example.com/script.js",
		"disposition":         "enforce",
		"document-uri":        "https://example.com/page",
		"effective-directive": "script-src-elem",
		"violated-directive":  "script-src-elem",
		"status-code":         200,
		"line-number":         42,
		"column-number":       15,
	}
	for key, want := range checks {
		got, exists := cspReport[key]
		if !exists {
			t.Errorf("missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("key %q: got %v, want %v", key, got, want)
		}
	}
}

func TestGenerateModernDeterministic(t *testing.T) {
	report := GenerateModern(false)

	if report["type"] != "csp-violation" {
		t.Errorf("type: got %v, want csp-violation", report["type"])
	}
	if report["age"] != 500 {
		t.Errorf("age: got %v, want 500", report["age"])
	}
	if report["url"] != "https://example.com/page" {
		t.Errorf("url: got %v, want https://example.com/page", report["url"])
	}

	body, ok := report["body"].(map[string]any)
	if !ok {
		t.Fatal("missing body key")
	}

	checks := map[string]any{
		"blockedURL":         "https://cdn.example.com/script.js",
		"disposition":        "enforce",
		"documentURL":        "https://example.com/page",
		"effectiveDirective": "script-src-elem",
		"statusCode":         200,
		"lineNumber":         42,
		"columnNumber":       15,
	}
	for key, want := range checks {
		got, exists := body[key]
		if !exists {
			t.Errorf("missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("key %q: got %v, want %v", key, got, want)
		}
	}
}

func TestGenerateLegacyRandom(t *testing.T) {
	report := GenerateLegacy(true)

	cspReport, ok := report["csp-report"].(map[string]any)
	if !ok {
		t.Fatal("missing csp-report key")
	}

	requiredKeys := []string{
		"blocked-uri", "disposition", "document-uri",
		"effective-directive", "original-policy", "referrer",
		"status-code", "violated-directive", "source-file",
		"line-number", "column-number",
	}
	for _, key := range requiredKeys {
		if _, exists := cspReport[key]; !exists {
			t.Errorf("missing key %q", key)
		}
	}
}

func TestGenerateModernRandom(t *testing.T) {
	report := GenerateModern(true)

	if report["type"] != "csp-violation" {
		t.Errorf("type: got %v, want csp-violation", report["type"])
	}
	if _, ok := report["age"]; !ok {
		t.Error("missing age key")
	}
	if _, ok := report["user_agent"]; !ok {
		t.Error("missing user_agent key")
	}

	body, ok := report["body"].(map[string]any)
	if !ok {
		t.Fatal("missing body key")
	}

	requiredKeys := []string{
		"blockedURL", "columnNumber", "disposition",
		"documentURL", "effectiveDirective", "lineNumber",
		"originalPolicy", "referrer", "sample",
		"sourceFile", "statusCode",
	}
	for _, key := range requiredKeys {
		if _, exists := body[key]; !exists {
			t.Errorf("missing key %q", key)
		}
	}
}

func TestGenerateLegacyDeterministicIsStable(t *testing.T) {
	a := GenerateLegacy(false)
	b := GenerateLegacy(false)

	aReport := a["csp-report"].(map[string]any)
	bReport := b["csp-report"].(map[string]any)

	for key, va := range aReport {
		if vb := bReport[key]; va != vb {
			t.Errorf("key %q differs between calls: %v vs %v", key, va, vb)
		}
	}
}

func TestGenerateModernDeterministicIsStable(t *testing.T) {
	a := GenerateModern(false)
	b := GenerateModern(false)

	if a["age"] != b["age"] {
		t.Error("age differs between deterministic calls")
	}
	if a["url"] != b["url"] {
		t.Error("url differs between deterministic calls")
	}
	if a["user_agent"] != b["user_agent"] {
		t.Error("user_agent differs between deterministic calls")
	}
}
