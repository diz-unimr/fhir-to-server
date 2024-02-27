package config

import "github.com/rs/zerolog/log"

type AppLogger struct {
}

func DefaultLogger() AppLogger {
	return AppLogger{}
}

func (l AppLogger) Errorf(format string, v ...interface{}) {
	log.Error().Msgf(format, v...)
}
func (l AppLogger) Warnf(format string, v ...interface{}) {
	log.Warn().Msgf(format, v...)
}
func (l AppLogger) Debugf(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}
