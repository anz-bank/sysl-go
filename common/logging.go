package common

import (
	"context"

	"github.com/anz-bank/pkg/logging"

	"github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

type loggerKey struct{}

// Logger is a component used to perform logging.
type Logger interface {
	Error(message string, err error)
	Info(message string)
	Debug(message string)
	WithStr(key string, value string) Logger // return a new logger that persists the given key/value
	WithLevel(level Level) Logger            // return a new logger that logs the given level or below
}

// Level represents the level at which a logger will log.
type Level int

const (
	ErrorLevel Level = iota
	InfoLevel
	DebugLevel
)

// Error logs the given error and message against the context found in the logger.
func Error(ctx context.Context, message string, err error) {
	GetLogger(ctx).Error(message, err)
}

// Info logs the given message against the context found in the logger.
func Info(ctx context.Context, message string) {
	GetLogger(ctx).Info(message)
}

// Debug logs the given message against the context found in the logger.
func Debug(ctx context.Context, message string) {
	GetLogger(ctx).Debug(message)
}

// WithStr returns the given context with a logger that persists the given key/value.
func WithStr(ctx context.Context, key string, value string) context.Context {
	return PutLogger(ctx, GetLogger(ctx).WithStr(key, value))
}

// GetLogger returns the logger from the context, or nil if no logger can be found.
func GetLogger(ctx context.Context) Logger {
	m, _ := ctx.Value(loggerKey{}).(Logger)
	return m
}

// PutLogger puts the given logger in the context.
func PutLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// NewPkgLogger returns is an implementation of Logger that uses the pkg/log logger.
func NewPkgLogger(fields log.Fields) Logger {
	return &PkgLogger{fields}
}

type PkgLogger struct {
	Fields log.Fields
}

func (l *PkgLogger) logger() log.Logger              { return l.Fields.From(context.Background()) }
func (l *PkgLogger) Error(message string, err error) { l.logger().Error(err, message) }
func (l *PkgLogger) Info(message string)             { l.logger().Info(message) }
func (l *PkgLogger) Debug(message string)            { l.logger().Debug(message) }

func (l *PkgLogger) WithStr(key string, value string) Logger {
	return &PkgLogger{l.Fields.With(key, value)}
}

func (l *PkgLogger) WithLevel(level Level) Logger {
	return &PkgLogger{l.Fields.WithConfigs(log.SetVerboseMode(level == DebugLevel))}
}

// NewZeroPkgLogger returns is an implementation of Logger that uses the pkg/logging logger.
func NewZeroPkgLogger(logger *logging.Logger) Logger {
	return &ZeroPkgLogger{logger}
}

type ZeroPkgLogger struct {
	Logger *logging.Logger
}

func (l *ZeroPkgLogger) Error(message string, err error) { l.Logger.Error(err).Msg(message) }
func (l *ZeroPkgLogger) Info(message string)             { l.Logger.Info().Msg(message) }
func (l *ZeroPkgLogger) Debug(message string)            { l.Logger.Debug().Msg(message) }

func (l *ZeroPkgLogger) WithStr(key string, value string) Logger {
	return &ZeroPkgLogger{l.Logger.WithStr(key, value)}
}

func (l *ZeroPkgLogger) WithLevel(level Level) Logger {
	var lvl logging.Level
	switch level {
	case ErrorLevel:
		lvl = logging.ErrorLevel
	case InfoLevel:
		lvl = logging.InfoLevel
	case DebugLevel:
		lvl = logging.DebugLevel
	}
	return &ZeroPkgLogger{l.Logger.WithLevel(lvl)}
}

// NewLogrusLogger returns an implementation of Logger that uses the Logrus logger.
func NewLogrusLogger(logger *logrus.Logger) Logger {
	return &LogrusLogger{Logger: logger}
}

type LogrusLogger struct {
	Logger *logrus.Logger
	Fields logrus.Fields
}

func (l *LogrusLogger) logger() *logrus.Entry           { return l.Logger.WithFields(l.Fields) }
func (l *LogrusLogger) Error(message string, err error) { l.logger().WithError(err).Error(message) }
func (l *LogrusLogger) Info(message string)             { l.logger().Info(message) }
func (l *LogrusLogger) Debug(message string)            { l.logger().Debug(message) }

func (l *LogrusLogger) WithStr(key string, value string) Logger {
	fields := make(map[string]interface{})
	for key, value := range l.Fields {
		fields[key] = value
	}
	fields[key] = value
	return &LogrusLogger{l.Logger, fields}
}

func (l *LogrusLogger) WithLevel(level Level) Logger {
	// Note: This method returns the same logger instance because the logrus logger mutates
	// the logger instance itself when setting the log level.
	var lvl logrus.Level
	switch level {
	case ErrorLevel:
		lvl = logrus.ErrorLevel
	case InfoLevel:
		lvl = logrus.InfoLevel
	case DebugLevel:
		lvl = logrus.DebugLevel
	}
	l.Logger.SetLevel(lvl)
	return l
}
