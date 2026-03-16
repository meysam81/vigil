package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/meysam81/vigil/cmd/server"
	"github.com/meysam81/vigil/internal/logger"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cli.VersionPrinter = func(c *cli.Command) {
		fmt.Println(version)
	}

	cmd := &cli.Command{
		Name:                  "vigil",
		Usage:                 "Collect and store CSP violation reports.",
		Version:               version,
		EnableShellCompletion: true,
		Suggest:               true,
		DefaultCommand:        "server",
		Commands: []*cli.Command{
			server.Command(),
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show detailed version information",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Printf("version: %s\ncommit: %s\ndate: %s\nbuilt by: %s\n", version, commit, date, builtBy)
					return nil
				},
			},
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	if err := cmd.Run(ctx, os.Args); err != nil {
		log := logger.NewLogger("error", true)
		log.Error().Err(err).Msg("fatal error")
		cancel()
		os.Exit(1)
	}
	cancel()
}
