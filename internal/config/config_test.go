package config

import (
	"strings"
	"testing"
	"time"
)

// validBase returns a Config with all fields set to valid values.
func validBase() *Config {
	return &Config{
		Server:    ServerConfig{Port: 8080, MaxBodySize: 65536},
		Redis:     RedisConfig{Host: "localhost", Port: 6379},
		RateLimit: RateLimitConfig{MaxRPS: 20, RefillRate: 2.0},
		Slack: SlackConfig{
			WebhookURL:     "https://hooks.slack.com/test",
			ReportInterval: 24 * time.Hour,
			MaxRetries:     5,
			RetryMinDelay:  3 * time.Second,
			RetryMaxDelay:  20 * time.Second,
		},
	}
}

func TestValidate_MinReportInterval(t *testing.T) {
	cfg := validBase()
	cfg.Slack.ReportInterval = 30 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for ReportInterval < 1m")
	}
	if !strings.Contains(err.Error(), "REPORT_INTERVAL") {
		t.Errorf("error should mention REPORT_INTERVAL, got: %v", err)
	}
}

func TestValidate_RetryBounds(t *testing.T) {
	cfg := validBase()
	cfg.Slack.RetryMinDelay = 30 * time.Second
	cfg.Slack.RetryMaxDelay = 5 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for RetryMinDelay > RetryMaxDelay")
	}
	if !strings.Contains(err.Error(), "SLACK_RETRY_MIN_DELAY") {
		t.Errorf("error should mention SLACK_RETRY_MIN_DELAY, got: %v", err)
	}
}

func TestValidate_NegativeMaxRetries(t *testing.T) {
	cfg := validBase()
	cfg.Slack.MaxRetries = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for MaxRetries < 0")
	}
	if !strings.Contains(err.Error(), "SLACK_MAX_RETRIES") {
		t.Errorf("error should mention SLACK_MAX_RETRIES, got: %v", err)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validBase()

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_PortBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		port int
		msg  string
	}{
		{"zero", 0, "PORT"},
		{"too high", 70000, "PORT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validBase()
			cfg.Server.Port = tc.port
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected error for port=%d", tc.port)
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Errorf("expected %q in error, got: %v", tc.msg, err)
			}
		})
	}
}

func TestValidate_MaxBodySize(t *testing.T) {
	cfg := validBase()
	cfg.Server.MaxBodySize = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for MaxBodySize=0")
	}
	if !strings.Contains(err.Error(), "MAX_BODY_SIZE") {
		t.Errorf("expected MAX_BODY_SIZE in error, got: %v", err)
	}
}

func TestValidate_RedisPort(t *testing.T) {
	cfg := validBase()
	cfg.Redis.Port = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for Redis.Port=0")
	}
	if !strings.Contains(err.Error(), "REDIS_PORT") {
		t.Errorf("expected REDIS_PORT in error, got: %v", err)
	}
}

func TestValidate_RateLimitBounds(t *testing.T) {
	t.Run("MaxRPS zero", func(t *testing.T) {
		cfg := validBase()
		cfg.RateLimit.MaxRPS = 0
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for MaxRPS=0")
		}
		if !strings.Contains(err.Error(), "RATELIMIT_MAX") {
			t.Errorf("expected RATELIMIT_MAX in error, got: %v", err)
		}
	})

	t.Run("RefillRate zero", func(t *testing.T) {
		cfg := validBase()
		cfg.RateLimit.RefillRate = 0
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected error for RefillRate=0")
		}
		if !strings.Contains(err.Error(), "RATELIMIT_REFILL") {
			t.Errorf("expected RATELIMIT_REFILL in error, got: %v", err)
		}
	})
}
