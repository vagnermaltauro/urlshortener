package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

type zerologLogger struct {
	logger zerolog.Logger
}

func New(level, environment string) Logger {

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	var output io.Writer = os.Stdout
	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Str("service", "url-shortener").
		Str("environment", environment).
		Logger()

	return &zerologLogger{logger: logger}
}

func (l *zerologLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug().Fields(toMap(fields)).Msg(msg)
}

func (l *zerologLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info().Fields(toMap(fields)).Msg(msg)
}

func (l *zerologLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn().Fields(toMap(fields)).Msg(msg)
}

func (l *zerologLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error().Fields(toMap(fields)).Msg(msg)
}

func (l *zerologLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal().Fields(toMap(fields)).Msg(msg)
}

func (l *zerologLogger) With(fields ...interface{}) Logger {
	return &zerologLogger{
		logger: l.logger.With().Fields(toMap(fields)).Logger(),
	}
}

func toMap(fields []interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		m[key] = fields[i+1]
	}
	return m
}
