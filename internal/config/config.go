package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	Port        int   `env:"PORT" envDefault:"8080"`
	MaxBodySize int64 `env:"MAX_BODY_SIZE" envDefault:"65536"` // 64KB
}

type RedisConfig struct {
	Host        string        `env:"REDIS_HOST"`
	Port        int           `env:"REDIS_PORT" envDefault:"6379"`
	DB          int           `env:"REDIS_DB" envDefault:"0"`
	Password    string        `env:"REDIS_PASSWORD"`
	SSLRequired bool          `env:"REDIS_SSL_ENABLED" envDefault:"false"`
	KeyTTL      time.Duration `env:"REDIS_KEY_TTL" envDefault:"720h"` // 30 days
}

type RateLimitConfig struct {
	MaxRPS     int     `env:"RATELIMIT_MAX" envDefault:"20"`
	RefillRate float32 `env:"RATELIMIT_REFILL" envDefault:"2.0"`
}

type CORSConfig struct {
	Enabled        bool   `env:"CORS_ENABLED" envDefault:"true"`
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`
}

// ReporterConfig controls the daily aggregate reporter schedule and retry behaviour.
// These settings apply to all notifiers (Slack, Discord, etc.).
type ReporterConfig struct {
	ScheduleHour  int           `env:"REPORT_SCHEDULE_HOUR"      envDefault:"10"`
	ScheduleMin   int           `env:"REPORT_SCHEDULE_MINUTE"    envDefault:"0"`
	MaxRetries    int           `env:"REPORT_MAX_RETRIES"        envDefault:"5"`
	RetryMinDelay time.Duration `env:"REPORT_RETRY_MIN_DELAY"    envDefault:"3s"`
	RetryMaxDelay time.Duration `env:"REPORT_RETRY_MAX_DELAY"    envDefault:"20s"`
}

// SlackConfig holds the Slack webhook URL.
type SlackConfig struct {
	WebhookURL string `env:"SLACK_WEBHOOK_URL"`
}

// DiscordConfig holds the Discord webhook URL.
type DiscordConfig struct {
	WebhookURL string `env:"DISCORD_WEBHOOK_URL"`
}

type Config struct {
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	Server    ServerConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Reporter  ReporterConfig
	Slack     SlackConfig
	Discord   DiscordConfig
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.migrateDeprecated()
	return cfg, nil
}

// migrateDeprecated copies values from deprecated SLACK_* retry env vars
// into ReporterConfig when the new REPORT_* equivalents are not set.
func (c *Config) migrateDeprecated() {
	if os.Getenv("REPORT_MAX_RETRIES") == "" {
		if v := os.Getenv("SLACK_MAX_RETRIES"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				c.Reporter.MaxRetries = n
			}
		}
	}
	if os.Getenv("REPORT_RETRY_MIN_DELAY") == "" {
		if v := os.Getenv("SLACK_RETRY_MIN_DELAY"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				c.Reporter.RetryMinDelay = d
			}
		}
	}
	if os.Getenv("REPORT_RETRY_MAX_DELAY") == "" {
		if v := os.Getenv("SLACK_RETRY_MAX_DELAY"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				c.Reporter.RetryMaxDelay = d
			}
		}
	}
}

func (c *Config) Validate() error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, errors.New("PORT must be between 1 and 65535"))
	}
	if c.Server.MaxBodySize <= 0 {
		errs = append(errs, errors.New("MAX_BODY_SIZE must be > 0"))
	}

	if strings.TrimSpace(c.Redis.Host) == "" {
		errs = append(errs, errors.New("REDIS_HOST is required"))
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		errs = append(errs, errors.New("REDIS_PORT must be between 1 and 65535"))
	}

	if c.RateLimit.MaxRPS <= 0 {
		errs = append(errs, errors.New("RATELIMIT_MAX must be > 0"))
	}
	if c.RateLimit.RefillRate <= 0 {
		errs = append(errs, errors.New("RATELIMIT_REFILL must be > 0"))
	}

	if url := strings.TrimSpace(c.Slack.WebhookURL); url != "" && !strings.HasPrefix(url, "https://") {
		errs = append(errs, errors.New("SLACK_WEBHOOK_URL must use https://"))
	}
	if url := strings.TrimSpace(c.Discord.WebhookURL); url != "" && !strings.HasPrefix(url, "https://") {
		errs = append(errs, errors.New("DISCORD_WEBHOOK_URL must use https://"))
	}

	if c.Reporter.ScheduleHour < 0 || c.Reporter.ScheduleHour > 23 {
		errs = append(errs, errors.New("REPORT_SCHEDULE_HOUR must be between 0 and 23"))
	}
	if c.Reporter.ScheduleMin < 0 || c.Reporter.ScheduleMin > 59 {
		errs = append(errs, errors.New("REPORT_SCHEDULE_MINUTE must be between 0 and 59"))
	}
	if c.Reporter.MaxRetries < 0 {
		errs = append(errs, errors.New("REPORT_MAX_RETRIES must be >= 0"))
	}
	if c.Reporter.RetryMinDelay > c.Reporter.RetryMaxDelay {
		errs = append(errs, errors.New("REPORT_RETRY_MIN_DELAY must be <= REPORT_RETRY_MAX_DELAY"))
	}

	return errors.Join(errs...)
}

// Deprecations returns warnings for deprecated env vars that are still set.
func (c *Config) Deprecations() []string {
	var warnings []string
	if os.Getenv("REPORT_INTERVAL") != "" {
		warnings = append(warnings, "REPORT_INTERVAL is deprecated; use REPORT_SCHEDULE_HOUR and REPORT_SCHEDULE_MINUTE instead")
	}
	deprecated := []struct{ old, replacement string }{
		{"SLACK_MAX_RETRIES", "REPORT_MAX_RETRIES"},
		{"SLACK_RETRY_MIN_DELAY", "REPORT_RETRY_MIN_DELAY"},
		{"SLACK_RETRY_MAX_DELAY", "REPORT_RETRY_MAX_DELAY"},
	}
	for _, d := range deprecated {
		if os.Getenv(d.old) != "" {
			warnings = append(warnings, fmt.Sprintf("%s is deprecated; use %s instead", d.old, d.replacement))
		}
	}
	return warnings
}
