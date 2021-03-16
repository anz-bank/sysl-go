package log

import (
	"context"
	"fmt"
	"time"

	zero "github.com/anz-bank/pkg/logging"

	pkg "github.com/anz-bank/pkg/log"
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

	// Inject puts the logger into the context, returning the new context and a function that
	// can be later used to restore the logger from the context.
	Inject(ctx context.Context) (context.Context, func(ctx context.Context) Logger)
}

// Level represents the level at which a logger will log.
// The currently supported values are:
// 2 - Error
// 4 - Info
// 5 - Debug.
type Level int

const (
	ErrorLevel = Level(logrus.ErrorLevel) // 2
	InfoLevel  = Level(logrus.InfoLevel)  // 4
	DebugLevel = Level(logrus.DebugLevel) // 5
)

func (l Level) String() string {
	switch l {
	case ErrorLevel:
		return "error"
	case InfoLevel:
		return "info"
	default:
		return "debug"
	}
}

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

// WithLevel returns the given context with a logger that logs at the given level.
func WithLevel(ctx context.Context, level Level) context.Context {
	return PutLogger(ctx, GetLogger(ctx).WithLevel(level))
}

// GetLogger returns the logger from the context, or nil if no logger can be found.
func GetLogger(ctx context.Context) Logger {
	fn, _ := ctx.Value(loggerKey{}).(func(ctx context.Context) Logger)
	if fn != nil {
		return fn(ctx)
	}
	return nil
}

// PutLogger puts the given logger in the context.
func PutLogger(ctx context.Context, logger Logger) context.Context {
	ctx, fn := logger.Inject(ctx)
	return context.WithValue(ctx, loggerKey{}, fn)
}

// NewDefaultLogger returns a logger that is regarded as the default logger to use within an
// application when no logger configuration is provided.
func NewDefaultLogger() Logger {
	return NewPkgLogger(pkg.Fields{})
}

// NewPkgLogger returns is an implementation of Logger that uses the pkg/log logger.
func NewPkgLogger(fields pkg.Fields) Logger {
	return &pkgLogger{fields}
}

type pkgLogger struct {
	fields pkg.Fields
}

func (l *pkgLogger) logger() pkg.Logger              { return l.fields.From(context.Background()) }
func (l *pkgLogger) Error(err error, message string) { l.logger().Error(err, message) }
func (l *pkgLogger) Info(message string)             { l.logger().Info(message) }
func (l *pkgLogger) Debug(message string)            { l.logger().Debug(message) }

func (l *pkgLogger) WithStr(key string, value string) Logger {
	return &pkgLogger{l.fields.With(key, value)}
}

func (l *pkgLogger) WithInt(key string, value int) Logger {
	return &pkgLogger{l.fields.With(key, value)}
}

func (l *pkgLogger) WithDuration(key string, value time.Duration) Logger {
	return &pkgLogger{l.fields.With(key, value)}
}

func (l *pkgLogger) WithLevel(level Level) Logger {
	return &pkgLogger{l.fields.WithConfigs(pkg.SetVerboseMode(level == DebugLevel))}
}

func (l *pkgLogger) Inject(ctx context.Context) (context.Context, func(ctx context.Context) Logger) {
	// Put and restore the logger natively. Rather than referencing the instance directly for the
	// purpose of restoration, this approach has the benefit of ensuring that any fields added
	// directly to the native logger aren't lost if the application uses both a native a wrapped logger.
	return l.fields.Onto(ctx), func(c context.Context) Logger { return &pkgLogger{pkg.FieldsFrom(c)} }
}

// NewZeroPkgLogger returns is an implementation of Logger that uses the pkg/logging logger.
func NewZeroPkgLogger(logger *zero.Logger) Logger {
	return &zeroPkgLogger{logger}
}

type zeroPkgLogger struct {
	logger *zero.Logger
}

func (l *zeroPkgLogger) Error(err error, message string) { l.logger.Error(err).Msg(message) }
func (l *zeroPkgLogger) Info(message string)             { l.logger.Info().Msg(message) }
func (l *zeroPkgLogger) Debug(message string)            { l.logger.Debug().Msg(message) }

func (l *zeroPkgLogger) WithStr(key string, value string) Logger {
	return &zeroPkgLogger{l.logger.WithStr(key, value)}
}

func (l *zeroPkgLogger) WithInt(key string, value int) Logger {
	return &zeroPkgLogger{l.logger.WithInt(key, value)}
}

func (l *zeroPkgLogger) WithDuration(key string, value time.Duration) Logger {
	return &zeroPkgLogger{l.logger.WithDur(key, value)}
}

func (l *zeroPkgLogger) WithLevel(level Level) Logger {
	var lvl zero.Level
	switch level {
	case ErrorLevel:
		lvl = zero.ErrorLevel
	case InfoLevel:
		lvl = zero.InfoLevel
	case DebugLevel:
		lvl = zero.DebugLevel
	}
	return &zeroPkgLogger{l.logger.WithLevel(lvl)}
}

func (l *zeroPkgLogger) Inject(ctx context.Context) (context.Context, func(ctx context.Context) Logger) {
	// Put and restore the logger natively. Rather than referencing the instance directly for the
	// purpose of restoration, this approach has the benefit of ensuring that any fields added
	// directly to the native logger aren't lost if the application uses both a native a wrapped logger.
	return l.logger.ToContext(ctx), func(c context.Context) Logger { return &zeroPkgLogger{zero.FromContext(c)} }
}

// NewLogrusLogger returns an implementation of Logger that uses the Logrus logger.
func NewLogrusLogger(logger *logrus.Logger) Logger {
	return &logrusLogger{logger: logger}
}

type logrusLogger struct {
	logger *logrus.Logger
	fields logrus.Fields
}

func (l *logrusLogger) entry() *logrus.Entry { return l.logger.WithFields(l.fields) }

func (l *logrusLogger) Error(err error, message string) { l.entry().WithError(err).Error(message) }
func (l *logrusLogger) Info(message string)             { l.entry().Info(message) }
func (l *logrusLogger) Debug(message string)            { l.entry().Debug(message) }

func (l *logrusLogger) WithStr(key string, value string) Logger { return l.withField(key, value) }
func (l *logrusLogger) WithInt(key string, value int) Logger    { return l.withField(key, value) }
func (l *logrusLogger) WithDuration(key string, value time.Duration) Logger {
	return l.withField(key, value)
}

func (l *logrusLogger) withField(key string, value interface{}) Logger {
	fields := make(map[string]interface{})
	for key, value := range l.fields {
		fields[key] = value
	}
	fields[key] = value
	return &logrusLogger{l.logger, fields}
}

func (l *logrusLogger) WithLevel(level Level) Logger {
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
	l.logger.SetLevel(lvl)
	return l
}

func (l *logrusLogger) Inject(ctx context.Context) (context.Context, func(ctx context.Context) Logger) {
	// Note: Logrus does not provide a native ability to add itself to the context, however,
	// historically Sysl-go has provided utility methods to inject a logrus logger into the context.
	// This approach is deprecated but is presently included in order to support legacy applications
	// that continue to use Logrus directly. The obvious downside of using Logrus directly is that
	// there is no built-in mechanism to persist key/value pairs within the context.
	return LogrusLoggerToContext(ctx, l.logger, GetLogrusLogEntryFromContext(ctx)), func(ctx context.Context) Logger { return l }
}

type logrusRequestContextKey struct{}

type logrusRequestContext struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// Deprecated: Use GetLogger, Error, Info or Debug methods.
func LogrusLoggerToContext(ctx context.Context, logger *logrus.Logger, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logrusRequestContextKey{}, logrusRequestContext{logger, entry})
}

// Deprecated: Use log.GetLogger.
func GetLogrusLogEntryFromContext(ctx context.Context) *logrus.Entry {
	core := ctx.Value(logrusRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(logrusRequestContext).entry
}

// Deprecated: Use log.GetLogger.
func GetLogrusLoggerFromContext(ctx context.Context) *logrus.Logger {
	core := ctx.Value(logrusRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(logrusRequestContext).logger
}
