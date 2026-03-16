package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/csp-report-collector/internal/config"
)

func New(ctx context.Context, cfg *config.RedisConfig) (*goredis.Client, error) {
	opts := &goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.SSLRequired {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	client := goredis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
