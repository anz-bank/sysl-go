package testutil

import (
	"context"
	"time"

	"github.com/anz-bank/sysl-go/config"

	"github.com/anz-bank/sysl-go/log"
)

type TestLogEntries struct {
	Entries []TestLogEntry
}

// The TestLogger is an implementation of log.Logger suitable for use within unit tests.
// Provision with NewTestLogger or NewTestContext.
type TestLogger struct {
	Level   log.Level
	Fields  map[string]interface{}
	entries *TestLogEntries // shared entries
}

type TestLogEntry struct {
	Level   log.Level
	Message string
	Error   error
	Fields  map[string]interface{}
}

func NewTestLogger() *TestLogger {
	return &TestLogger{Level: log.InfoLevel, entries: &TestLogEntries{}}
}

func (l *TestLogger) Entries() []TestLogEntry {
	return l.entries.Entries
}

func (l *TestLogger) EntryCount() int {
	return len(l.entries.Entries)
}

func (l *TestLogger) Error(err error, message string) { l.log(log.ErrorLevel, err, message) }
func (l *TestLogger) Info(message string)             { l.log(log.InfoLevel, nil, message) }
func (l *TestLogger) Debug(message string)            { l.log(log.DebugLevel, nil, message) }

func (l *TestLogger) log(level log.Level, err error, message string) {
	if l.Level >= level {
		l.entries.Entries = append(l.entries.Entries, TestLogEntry{
			Level:   level,
			Message: message,
			Error:   err,
			Fields:  l.copyFields(),
		})
	}
}

func (l *TestLogger) WithStr(key string, value string) log.Logger {
	return l.withField(key, value)
}

func (l *TestLogger) WithInt(key string, value int) log.Logger {
	return l.withField(key, value)
}

func (l *TestLogger) WithDuration(key string, value time.Duration) log.Logger {
	return l.withField(key, value)
}

func (l *TestLogger) withField(key string, value interface{}) log.Logger {
	fields := l.copyFields()
	fields[key] = value
	return &TestLogger{l.Level, fields, l.entries}
}

func (l *TestLogger) WithLevel(level log.Level) log.Logger {
	return &TestLogger{level, l.Fields, l.entries}
}

func (l *TestLogger) Inject(ctx context.Context) (context.Context, func(ctx context.Context) log.Logger) {
	return ctx, func(_ context.Context) log.Logger { return l } // return single instance
}

func (l *TestLogger) copyFields() map[string]interface{} {
	fields := make(map[string]interface{})
	for key, value := range l.Fields {
		fields[key] = value
	}
	return fields
}

func (l *TestLogger) LastEntry() *TestLogEntry {
	if len(l.entries.Entries) == 0 {
		return nil
	}
	return &l.entries.Entries[len(l.entries.Entries)-1]
}

// TestContextOpt is an interface to help configure the test context.
type TestContextOpt interface {
	Apply(ctx context.Context) context.Context
}

// NewTestContext returns a context suitable for use within unit tests.
// The context comes equipped with the following:
// 1. Logger
// To modify the test context further, pass additional TestContextOpt values.
func NewTestContext(opts ...TestContextOpt) context.Context {
	ctx, _ := NewTestContextWithLogger(opts...)
	return ctx
}

// NewTestContext returns a context (and logger) suitable for use within unit tests.
// The context comes equipped with the following:
// 1. Logger
// To modify the test context further, pass additional TestContextOpt values.
func NewTestContextWithLogger(opts ...TestContextOpt) (context.Context, *TestLogger) {
	ctx := log.PutLogger(context.Background(), NewTestLogger())
	for _, o := range opts {
		ctx = o.Apply(ctx)
	}
	logger := log.GetLogger(ctx)
	return ctx, logger.(*TestLogger)
}

// WithLogLevel returns a TestContextOpt that sets the log level within the test context.
func WithLogLevel(level log.Level) TestContextOpt {
	return &withLogLevel{level}
}

type withLogLevel struct {
	level log.Level
}

func (w *withLogLevel) Apply(ctx context.Context) context.Context {
	return log.WithLevel(ctx, w.level)
}

// WithConfig returns a TestContextOpt that adds the given configuration into the test context.
func WithConfig(cfg *config.DefaultConfig) TestContextOpt {
	return &withConfig{cfg}
}

type withConfig struct {
	cfg *config.DefaultConfig
}

func (w *withConfig) Apply(ctx context.Context) context.Context {
	return config.PutDefaultConfig(ctx, w.cfg)
}

// WithLogPayloadContents returns a TestContextOpt that sets the configuration option within the
// test context to log the payload contents or not.
func WithLogPayloadContents(log bool) TestContextOpt {
	return &withLogPayloadContents{log}
}

type withLogPayloadContents struct {
	log bool
}

func (w *withLogPayloadContents) Apply(ctx context.Context) context.Context {
	cfg := config.GetDefaultConfig(ctx)
	if cfg == nil {
		cfg = &config.DefaultConfig{}
	}
	if cfg.Development == nil {
		cfg.Development = &config.DevelopmentConfig{}
	}
	cfg.Development.LogPayloadContents = w.log
	return config.PutDefaultConfig(ctx, cfg)
}
