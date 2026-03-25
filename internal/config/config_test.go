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
		Reporter: ReporterConfig{
			ScheduleHour:  10,
			ScheduleMin:   0,
			MaxRetries:    5,
			RetryMinDelay: 3 * time.Second,
			RetryMaxDelay: 20 * time.Second,
		},
		Slack:   SlackConfig{WebhookURL: "https://hooks.slack.com/test"},
		Discord: DiscordConfig{},
	}
}

func TestValidate_ScheduleHourBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		hour int
	}{
		{"negative", -1},
		{"too high", 24},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validBase()
			cfg.Reporter.ScheduleHour = tc.hour
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected error for ScheduleHour=%d", tc.hour)
			}
			if !strings.Contains(err.Error(), "REPORT_SCHEDULE_HOUR") {
				t.Errorf("error should mention REPORT_SCHEDULE_HOUR, got: %v", err)
			}
		})
	}
}

func TestValidate_ScheduleMinuteBounds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		minute int
	}{
		{"negative", -1},
		{"too high", 60},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validBase()
			cfg.Reporter.ScheduleMin = tc.minute
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected error for ScheduleMin=%d", tc.minute)
			}
			if !strings.Contains(err.Error(), "REPORT_SCHEDULE_MINUTE") {
				t.Errorf("error should mention REPORT_SCHEDULE_MINUTE, got: %v", err)
			}
		})
	}
}

func TestDeprecations_ReportInterval(t *testing.T) {
	t.Setenv("REPORT_INTERVAL", "24h")
	cfg := validBase()
	warnings := cfg.Deprecations()
	if len(warnings) == 0 {
		t.Fatal("expected deprecation warning for REPORT_INTERVAL")
	}
	if !strings.Contains(warnings[0], "REPORT_INTERVAL") {
		t.Errorf("warning should mention REPORT_INTERVAL, got: %s", warnings[0])
	}
}

func TestDeprecations_NoWarnings(t *testing.T) {
	cfg := validBase()
	warnings := cfg.Deprecations()
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got: %v", warnings)
	}
}

func TestDeprecations_SlackRetryVars(t *testing.T) {
	t.Setenv("SLACK_MAX_RETRIES", "3")
	t.Setenv("SLACK_RETRY_MIN_DELAY", "1s")
	t.Setenv("SLACK_RETRY_MAX_DELAY", "10s")

	cfg := validBase()
	warnings := cfg.Deprecations()
	if len(warnings) != 3 {
		t.Fatalf("expected 3 deprecation warnings, got %d: %v", len(warnings), warnings)
	}
	for _, w := range warnings {
		if !strings.Contains(w, "deprecated") {
			t.Errorf("expected 'deprecated' in warning, got: %s", w)
		}
	}
}

func TestMigrateDeprecated_SlackRetryFallback(t *testing.T) {
	t.Setenv("SLACK_MAX_RETRIES", "7")
	t.Setenv("SLACK_RETRY_MIN_DELAY", "2s")
	t.Setenv("SLACK_RETRY_MAX_DELAY", "30s")

	cfg := validBase()
	cfg.migrateDeprecated()

	if cfg.Reporter.MaxRetries != 7 {
		t.Errorf("MaxRetries: want 7, got %d", cfg.Reporter.MaxRetries)
	}
	if cfg.Reporter.RetryMinDelay != 2*time.Second {
		t.Errorf("RetryMinDelay: want 2s, got %s", cfg.Reporter.RetryMinDelay)
	}
	if cfg.Reporter.RetryMaxDelay != 30*time.Second {
		t.Errorf("RetryMaxDelay: want 30s, got %s", cfg.Reporter.RetryMaxDelay)
	}
}

func TestMigrateDeprecated_NewVarsTakePrecedence(t *testing.T) {
	t.Setenv("REPORT_MAX_RETRIES", "10")
	t.Setenv("SLACK_MAX_RETRIES", "3")

	cfg := validBase()
	cfg.Reporter.MaxRetries = 10 // as if parsed from REPORT_MAX_RETRIES
	cfg.migrateDeprecated()

	if cfg.Reporter.MaxRetries != 10 {
		t.Errorf("MaxRetries: want 10 (new var), got %d", cfg.Reporter.MaxRetries)
	}
}

func TestValidate_RetryBounds(t *testing.T) {
	cfg := validBase()
	cfg.Reporter.RetryMinDelay = 30 * time.Second
	cfg.Reporter.RetryMaxDelay = 5 * time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for RetryMinDelay > RetryMaxDelay")
	}
	if !strings.Contains(err.Error(), "REPORT_RETRY_MIN_DELAY") {
		t.Errorf("error should mention REPORT_RETRY_MIN_DELAY, got: %v", err)
	}
}

func TestValidate_NegativeMaxRetries(t *testing.T) {
	cfg := validBase()
	cfg.Reporter.MaxRetries = -1

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for MaxRetries < 0")
	}
	if !strings.Contains(err.Error(), "REPORT_MAX_RETRIES") {
		t.Errorf("error should mention REPORT_MAX_RETRIES, got: %v", err)
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

func TestValidate_DiscordWebhookURL(t *testing.T) {
	cfg := validBase()
	cfg.Discord.WebhookURL = "http://discord.com/api/webhooks/123/abc"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for non-HTTPS Discord webhook")
	}
	if !strings.Contains(err.Error(), "DISCORD_WEBHOOK_URL") {
		t.Errorf("expected DISCORD_WEBHOOK_URL in error, got: %v", err)
	}
}

func TestValidate_SlackWebhookHTTP(t *testing.T) {
	cfg := validBase()
	cfg.Slack.WebhookURL = "http://hooks.slack.com/test"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for non-HTTPS Slack webhook")
	}
	if !strings.Contains(err.Error(), "SLACK_WEBHOOK_URL") {
		t.Errorf("expected SLACK_WEBHOOK_URL in error, got: %v", err)
	}
}
