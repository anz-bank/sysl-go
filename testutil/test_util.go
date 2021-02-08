package testutil

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/log"
)

type TestLogger struct {
	Level   log.Level
	Fields  map[string]interface{}
	entries *TestLogEntries // shared entries
}

type TestLogEntries struct {
	Entries []TestLogEntry
}

type TestLogEntry struct {
	Level   log.Level
	Message string
	Error   error
	Fields  map[string]interface{}
}

func NewTestLogger() TestLogger {
	return TestLogger{entries: &TestLogEntries{}}
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

func (l *TestLogger) Inject(ctx context.Context) context.Context {
	return ctx // unsupported
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

func NewTestContext() context.Context {
	ctx, _ := NewTestContextWithLogger()
	return ctx
}

func NewTestContextWithLogger() (context.Context, TestLogger) {
	return NewTestContextWithLoggerAtLevel(log.InfoLevel)
}

func NewTestContextWithLoggerAtLevel(level log.Level) (context.Context, TestLogger) {
	test := NewTestLogger()
	logger := test.WithLevel(level).(*TestLogger)
	return log.PutLogger(context.Background(), logger), *logger
}

func LoggerHookContextMiddleware() (func(next http.Handler) http.Handler, TestLogger) {
	logger := NewTestLogger()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(log.PutLogger(r.Context(), &logger))
			next.ServeHTTP(w, r)
		})
	}, logger
}
