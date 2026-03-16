package server

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Usage:   "Start the CSP report collector server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx)
		},
	}
}
