package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidate_MinReportInterval(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{Host: "localhost"},
		Slack: SlackConfig{
			ReportInterval: 30 * time.Second,
			RetryMinDelay:  3 * time.Second,
			RetryMaxDelay:  20 * time.Second,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for ReportInterval < 1m")
	}
	if !strings.Contains(err.Error(), "REPORT_INTERVAL") {
		t.Errorf("error should mention REPORT_INTERVAL, got: %v", err)
	}
}

func TestValidate_RetryBounds(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{Host: "localhost"},
		Slack: SlackConfig{
			ReportInterval: 24 * time.Hour,
			RetryMinDelay:  30 * time.Second,
			RetryMaxDelay:  5 * time.Second,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for RetryMinDelay > RetryMaxDelay")
	}
	if !strings.Contains(err.Error(), "SLACK_RETRY_MIN_DELAY") {
		t.Errorf("error should mention SLACK_RETRY_MIN_DELAY, got: %v", err)
	}
}

func TestValidate_NegativeMaxRetries(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{Host: "localhost"},
		Slack: SlackConfig{
			ReportInterval: 24 * time.Hour,
			MaxRetries:     -1,
			RetryMinDelay:  3 * time.Second,
			RetryMaxDelay:  20 * time.Second,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for MaxRetries < 0")
	}
	if !strings.Contains(err.Error(), "SLACK_MAX_RETRIES") {
		t.Errorf("error should mention SLACK_MAX_RETRIES, got: %v", err)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{Host: "localhost"},
		Slack: SlackConfig{
			WebhookURL:     "https://hooks.slack.com/test",
			ReportInterval: 24 * time.Hour,
			MaxRetries:     5,
			RetryMinDelay:  3 * time.Second,
			RetryMaxDelay:  20 * time.Second,
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
