package sample

import (
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	"github.com/urfave/cli/v3"
)

// Command returns the "sample" subcommand.
func Command() *cli.Command {
	return &cli.Command{
		Name:    "sample",
		Aliases: []string{"sa"},
		Usage:   "Generate sample CSP violation reports for testing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "modern",
				Usage:   "Report format: legacy or modern",
			},
			&cli.IntFlag{
				Name:    "count",
				Aliases: []string{"c"},
				Value:   1,
				Usage:   "Number of reports to generate",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Write to file instead of stdout",
			},
			&cli.BoolFlag{
				Name:    "random",
				Aliases: []string{"r"},
				Usage:   "Randomize report data",
			},
			&cli.BoolFlag{
				Name:    "pretty",
				Aliases: []string{"p"},
				Usage:   "Pretty-print JSON output",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(cmd)
		},
	}
}

func run(cmd *cli.Command) error {
	format := cmd.String("format")
	count := cmd.Int("count")
	output := cmd.String("output")
	random := cmd.Bool("random")
	pretty := cmd.Bool("pretty")

	if format != "legacy" && format != "modern" {
		return fmt.Errorf("invalid format %q: must be legacy or modern", format)
	}
	if count < 1 {
		return fmt.Errorf("count must be at least 1, got %d", count)
	}

	generate := GenerateModern
	if format == "legacy" {
		generate = GenerateLegacy
	}

	reports := make([]map[string]any, count)
	for i := range count {
		reports[i] = generate(random)
	}

	// Modern: always array. Legacy: single object for count=1, array for count>1.
	var payload any
	if format == "modern" || count > 1 {
		payload = reports
	} else {
		payload = reports[0]
	}

	var data []byte
	var err error
	if pretty {
		data, err = json.MarshalIndent(payload, "", "  ")
	} else {
		data, err = json.Marshal(payload)
	}
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	data = append(data, '\n')

	if output != "" {
		if err := os.WriteFile(output, data, 0o644); err != nil {
			return fmt.Errorf("writing to %s: %w", output, err)
		}
		return nil
	}

	_, err = os.Stdout.Write(data)
	return err
}
