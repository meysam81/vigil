package server

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/meysam81/vigil/internal/logger"
)

var log *logger.Logger

func Command() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "Start the CSP report collector server",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			log = logger.NewLogger("info", false)
			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx)
		},
	}
}
