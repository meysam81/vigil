package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
)

func New(ctx context.Context, cfg *config.RedisConfig, log *logger.Logger) (*goredis.Client, error) {
	log.Debug().Str("host", cfg.Host).Int("port", cfg.Port).Msg("connecting to redis")

	opts := &goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.SSLRequired {
		log.Debug().Bool("tls", cfg.SSLRequired).Msg("redis TLS configuration")
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	client := goredis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Info().Str("addr", opts.Addr).Msg("redis connected")

	return client, nil
}
