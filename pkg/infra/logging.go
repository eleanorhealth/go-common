package infra

import (
	"os"
	"time"

	"github.com/eleanorhealth/go-common/pkg/env"
	"github.com/rs/zerolog"
)

func Logger() zerolog.Logger {
	// Configure Zerolog to work with Google Cloud Logging.
	// https://github.com/rs/zerolog/issues/174
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.LevelFieldName = "severity"

	logger := zerolog.New(os.Stdout)

	if env.Get("ZEROLOG_CONSOLE_WRITER", false) {
		logger = ConsoleLogger()
	}

	return logger
}

func ConsoleLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter())
}
