package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	Port int `env:"PORT" envDefault:"8080"`
}

type RedisConfig struct {
	Host        string `env:"REDIS_HOST"`
	Port        int    `env:"REDIS_PORT" envDefault:"6379"`
	DB          int    `env:"REDIS_DB" envDefault:"0"`
	Password    string `env:"REDIS_PASSWORD"`
	SSLRequired bool   `env:"REDIS_SSL_ENABLED" envDefault:"false"`
}

type RateLimitConfig struct {
	MaxRPS     int     `env:"RATELIMIT_MAX" envDefault:"20"`
	RefillRate float32 `env:"RATELIMIT_REFILL" envDefault:"2.0"`
}

type CORSConfig struct {
	Enabled        bool   `env:"CORS_ENABLED" envDefault:"true"`
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`
}

type Config struct {
	LogLevel  string          `env:"LOG_LEVEL" envDefault:"info"`
	Server    ServerConfig    `envPrefix:""`
	Redis     RedisConfig     `envPrefix:""`
	RateLimit RateLimitConfig `envPrefix:""`
	CORS      CORSConfig      `envPrefix:""`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() []string {
	var errs []string
	if strings.TrimSpace(c.Redis.Host) == "" {
		errs = append(errs, "REDIS_HOST is required")
	}
	return errs
}
