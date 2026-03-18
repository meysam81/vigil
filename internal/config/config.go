package config

import (
	"errors"
	"fmt"
	"os"
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

type SlackConfig struct {
	WebhookURL         string        `env:"SLACK_WEBHOOK_URL"`
	ReportScheduleHour int           `env:"REPORT_SCHEDULE_HOUR"  envDefault:"10"`
	ReportScheduleMin  int           `env:"REPORT_SCHEDULE_MINUTE" envDefault:"0"`
	MaxRetries         int           `env:"SLACK_MAX_RETRIES"     envDefault:"5"`
	RetryMinDelay      time.Duration `env:"SLACK_RETRY_MIN_DELAY" envDefault:"3s"`
	RetryMaxDelay      time.Duration `env:"SLACK_RETRY_MAX_DELAY" envDefault:"20s"`
}

type Config struct {
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	Server    ServerConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Slack     SlackConfig
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
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
	if c.Slack.ReportScheduleHour < 0 || c.Slack.ReportScheduleHour > 23 {
		errs = append(errs, errors.New("REPORT_SCHEDULE_HOUR must be between 0 and 23"))
	}
	if c.Slack.ReportScheduleMin < 0 || c.Slack.ReportScheduleMin > 59 {
		errs = append(errs, errors.New("REPORT_SCHEDULE_MINUTE must be between 0 and 59"))
	}
	if c.Slack.MaxRetries < 0 {
		errs = append(errs, errors.New("SLACK_MAX_RETRIES must be >= 0"))
	}
	if c.Slack.RetryMinDelay > c.Slack.RetryMaxDelay {
		errs = append(errs, errors.New("SLACK_RETRY_MIN_DELAY must be <= SLACK_RETRY_MAX_DELAY"))
	}

	return errors.Join(errs...)
}

// Deprecations returns warnings for deprecated env vars that are still set.
func (c *Config) Deprecations() []string {
	var warnings []string
	if os.Getenv("REPORT_INTERVAL") != "" {
		warnings = append(warnings, "REPORT_INTERVAL is deprecated; use REPORT_SCHEDULE_HOUR and REPORT_SCHEDULE_MINUTE instead")
	}
	return warnings
}
