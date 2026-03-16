package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log" //nolint:depguard // logger package is the only legitimate user of the global zerolog logger

	"github.com/meysam81/x/logging"
)

type Logger = zerolog.Logger

var Nop = zerolog.Nop

func NewLogger(logLevel string, noColor bool) *zerolog.Logger {
	l := logging.NewLogger(
		logging.WithLogLevel(logLevel),
		logging.WithColorsEnabled(!noColor),
	)
	log.Logger = l
	return &l
}
