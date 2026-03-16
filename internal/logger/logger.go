package logger

import (
	"github.com/rs/zerolog"

	"github.com/meysam81/x/logging"
)

type Logger = zerolog.Logger

func NewLogger(logLevel string, noColor bool) *zerolog.Logger {
	l := logging.NewLogger(
		logging.WithLogLevel(logLevel),
		logging.WithColorsEnabled(!noColor),
	)
	return &l
}
