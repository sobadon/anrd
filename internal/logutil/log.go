package logutil

import (
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger() zerolog.Logger {
	// zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	zerolog.CallerMarshalFunc = func(file string, line int) string {
		filename := filepath.Base(file)
		return filename + ":" + strconv.Itoa(line)
	}

	logger := log.With().Caller().Logger()

	return logger
}
