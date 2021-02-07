package log

import (
	"context"
	"fmt"
	"time"

	"github.com/anz-bank/pkg/logging"

	"github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

type loggerKey struct{}

// Logger is a component used to perform logging.
type Logger interface {
	Error(err error, message string)
	Info(message string)
	Debug(message string)

	// WithStr returns a new logger that persists the given key/value.
	WithStr(key string, value string) Logger

	// WithInt returns a new logger that persists the given key/value.
	WithInt(key string, value int) Logger

	// WithDuration returns a new logger that persists the given key/value.
	WithDuration(key string, value time.Duration) Logger

	// WithLevel returns a new logger that logs the given level or below.
	WithLevel(level Level) Logger

	// Inject adds the logger directly into the context (if applicable).
	// In order to support different logging implementations, Sysl-go wraps the implementation and
	// makes it available through the context. In addition to this, some applications are already
	// using a particular logging implementation by accessing it directly through the context. In
	// order to support both workflows, whenever Sysl-go injects its wrapped logged into the context,
	// it also provides the logging implementation, through this method, with an opportunity to
	// inject itself directly into the context also.
	Inject(ctx context.Context) context.Context
}

// Level represents the level at which a logger will log.
// The currently supported values are:
// 2 - Error
// 4 - Info
// 5 - Debug
type Level int

const (
	ErrorLevel = Level(logrus.ErrorLevel) // 2
	InfoLevel  = Level(logrus.InfoLevel)  // 4
	DebugLevel = Level(logrus.DebugLevel) // 5
)

// Error logs the given error and message against the context found in the logger.
func Error(ctx context.Context, err error, message string) {
	GetLogger(ctx).Error(err, message)
}

// Errorf logs the given error and message against the context found in the logger.
func Errorf(ctx context.Context, err error, format string, args ...interface{}) {
	GetLogger(ctx).Error(err, fmt.Sprintf(format, args...))
}

// Info logs the given message against the context found in the logger.
func Info(ctx context.Context, message string) {
	GetLogger(ctx).Info(message)
}

// Infof logs the given message against the context found in the logger.
func Infof(ctx context.Context, format string, args ...interface{}) {
	GetLogger(ctx).Info(fmt.Sprintf(format, args...))
}

// Debug logs the given message against the context found in the logger.
func Debug(ctx context.Context, message string) {
	GetLogger(ctx).Debug(message)
}

// Debugf logs the given message against the context found in the logger.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	GetLogger(ctx).Debug(fmt.Sprintf(format, args...))
}

// WithStr returns the given context with a logger that persists the given key/value.
func WithStr(ctx context.Context, key string, value string) context.Context {
	return PutLogger(ctx, GetLogger(ctx).WithStr(key, value))
}

// WithInt returns the given context with a logger that persists the given key/value.
func WithInt(ctx context.Context, key string, value int) context.Context {
	return PutLogger(ctx, GetLogger(ctx).WithInt(key, value))
}

// WithDuration returns the given context with a logger that persists the given key/value.
func WithDuration(ctx context.Context, key string, value time.Duration) context.Context {
	return PutLogger(ctx, GetLogger(ctx).WithDuration(key, value))
}

// GetLogger returns the logger from the context, or nil if no logger can be found.
func GetLogger(ctx context.Context) Logger {
	m, _ := ctx.Value(loggerKey{}).(Logger)
	return m
}

// PutLogger puts the given logger in the context.
func PutLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(logger.Inject(ctx), loggerKey{}, logger)
}

// NewDefaultLogger returns a logger that is regarded as the default logger to use within an
// application when no logger configuration is provided.
func NewDefaultLogger() Logger {
	return NewPkgLogger(log.Fields{})
}

// NewPkgLogger returns is an implementation of Logger that uses the pkg/log logger.
func NewPkgLogger(fields log.Fields) Logger {
	return &PkgLogger{fields}
}

type PkgLogger struct {
	Fields log.Fields
}

func (l *PkgLogger) logger() log.Logger              { return l.Fields.From(context.Background()) }
func (l *PkgLogger) Error(err error, message string) { l.logger().Error(err, message) }
func (l *PkgLogger) Info(message string)             { l.logger().Info(message) }
func (l *PkgLogger) Debug(message string)            { l.logger().Debug(message) }

func (l *PkgLogger) WithStr(key string, value string) Logger {
	return &PkgLogger{l.Fields.With(key, value)}
}

func (l *PkgLogger) WithInt(key string, value int) Logger {
	return &PkgLogger{l.Fields.With(key, value)}
}

func (l *PkgLogger) WithDuration(key string, value time.Duration) Logger {
	return &PkgLogger{l.Fields.With(key, value)}
}

func (l *PkgLogger) WithLevel(level Level) Logger {
	return &PkgLogger{l.Fields.WithConfigs(log.SetVerboseMode(level == DebugLevel))}
}

func (l *PkgLogger) Inject(ctx context.Context) context.Context {
	return l.Fields.Onto(ctx)
}

// NewZeroPkgLogger returns is an implementation of Logger that uses the pkg/logging logger.
func NewZeroPkgLogger(logger *logging.Logger) Logger {
	return &ZeroPkgLogger{logger}
}

type ZeroPkgLogger struct {
	Logger *logging.Logger
}

func (l *ZeroPkgLogger) Error(err error, message string) { l.Logger.Error(err).Msg(message) }
func (l *ZeroPkgLogger) Info(message string)             { l.Logger.Info().Msg(message) }
func (l *ZeroPkgLogger) Debug(message string)            { l.Logger.Debug().Msg(message) }

func (l *ZeroPkgLogger) WithStr(key string, value string) Logger {
	return &ZeroPkgLogger{l.Logger.WithStr(key, value)}
}

func (l *ZeroPkgLogger) WithInt(key string, value int) Logger {
	return &ZeroPkgLogger{l.Logger.WithInt(key, value)}
}

func (l *ZeroPkgLogger) WithDuration(key string, value time.Duration) Logger {
	return &ZeroPkgLogger{l.Logger.WithDur(key, value)}
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

func (l *ZeroPkgLogger) Inject(ctx context.Context) context.Context {
	return l.Logger.ToContext(ctx)
}

// NewLogrusLogger returns an implementation of Logger that uses the Logrus logger.
func NewLogrusLogger(logger *logrus.Logger) Logger {
	return &LogrusLogger{Logger: logger}
}

type LogrusLogger struct {
	Logger *logrus.Logger
	Fields logrus.Fields
}

func (l *LogrusLogger) logger() *logrus.Entry { return l.Logger.WithFields(l.Fields) }

func (l *LogrusLogger) Error(err error, message string) { l.logger().WithError(err).Error(message) }
func (l *LogrusLogger) Info(message string)             { l.logger().Info(message) }
func (l *LogrusLogger) Debug(message string)            { l.logger().Debug(message) }

func (l *LogrusLogger) WithStr(key string, value string) Logger { return l.withField(key, value) }
func (l *LogrusLogger) WithInt(key string, value int) Logger    { return l.withField(key, value) }
func (l *LogrusLogger) WithDuration(key string, value time.Duration) Logger {
	return l.withField(key, value)
}

func (l *LogrusLogger) withField(key string, value interface{}) Logger {
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

func (l *LogrusLogger) Inject(ctx context.Context) context.Context {
	// Note: Logrus does not provide a native ability to add itself to the context, however,
	// historically Sysl-go has provided utility methods to inject a logrus logger into the context.
	// This approach is deprecated and will soon be removed, after which this method will return
	// the passed-in context value.
	return LoggerToContext(ctx, l.Logger, nil)
}

// Deprecated
type coreRequestContextKey struct{}

// Deprecated
type coreRequestContext struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// Deprecated: Use GetLogger, Error, Info or Debug methods.
func LoggerToContext(ctx context.Context, logger *logrus.Logger, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, coreRequestContextKey{}, coreRequestContext{logger, entry})
}

// Deprecated: Use log.GetLogger.
func GetLogEntryFromContext(ctx context.Context) *logrus.Entry {
	core := ctx.Value(coreRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(coreRequestContext).entry
}

// Deprecated: Use log.GetLogger.
func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	core := ctx.Value(coreRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(coreRequestContext).logger
}
