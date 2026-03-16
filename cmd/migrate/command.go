package migrate

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/meysam81/vigil/internal/config"
	"github.com/meysam81/vigil/internal/logger"
	"github.com/meysam81/vigil/internal/migration"
	iredis "github.com/meysam81/vigil/internal/redis"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "migrate",
		Aliases: []string{"m"},
		Usage:   "Run database migrations (backfill vigil:timeline from existing csp:* keys)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx)
		},
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.NewLogger(cfg.LogLevel, false)

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	log.Info().Msg("configuration loaded")

	redisClient, err := iredis.New(ctx, &cfg.Redis, log)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error().Err(err).Msg("failed closing redis connection")
		}
	}()

	if err := migration.Run(ctx, redisClient, log); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
