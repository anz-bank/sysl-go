package log

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	zero "github.com/anz-bank/pkg/logging"

	"github.com/stretchr/testify/require"

	pkg "github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

func TestPkgLogger(t *testing.T) {
	newLogger := func(buf *bytes.Buffer) Logger { return NewPkgLogger(pkg.Fields{}.WithConfigs(pkg.SetOutput(buf))) }
	nativePersist := func(ctx context.Context, k string, v string) context.Context { return pkg.With(k, v).Onto(ctx) }
	nativeLog := func(ctx context.Context, buf *bytes.Buffer, str string) {
		pkg.FieldsFrom(ctx).WithConfigs(pkg.SetOutput(buf)).Info(ctx, str)
	}
	testLoggerEvents(t, newLogger)
	testLoggerLevel(t, newLogger)
	testLoggerPersistence(t, newLogger)
	testLoggerInterleave(t, newLogger, nativePersist, nativeLog)
}

func TestZeroPkgLogger(t *testing.T) {
	newLogger := func(buf *bytes.Buffer) Logger { return NewZeroPkgLogger(zero.New(buf)) }
	nativePersist := func(ctx context.Context, k string, v string) context.Context {
		return zero.FromContext(ctx).WithStr(k, v).ToContext(ctx)
	}
	nativeLog := func(ctx context.Context, buf *bytes.Buffer, str string) {
		zero.FromContext(ctx).WithOutput(buf).Info().Msg(str)
	}
	testLoggerEvents(t, newLogger)
	testLoggerLevel(t, newLogger)
	testLoggerPersistence(t, newLogger)
	testLoggerInterleave(t, newLogger, nativePersist, nativeLog)
}

func TestLogrusLogger(t *testing.T) {
	newLogger := func(buf *bytes.Buffer) Logger { lrs := logrus.New(); lrs.Out = buf; return NewLogrusLogger(lrs) }
	testLoggerEvents(t, newLogger)
	testLoggerLevel(t, newLogger)
	testLoggerPersistence(t, newLogger)
}

// Test that a logger logs its events appropriately.
func testLoggerEvents(t *testing.T, newLogger func(*bytes.Buffer) Logger) {
	buf := &bytes.Buffer{}
	ctx := PutLogger(context.Background(), newLogger(buf).WithLevel(DebugLevel))

	// Verify that an error level log is logged
	Error(ctx, errors.New("error"), "format")
	require.Contains(t, buf.String(), "error")
	require.Contains(t, buf.String(), "format")

	// Verify that an info level log is logged
	Info(ctx, "info")
	require.Contains(t, buf.String(), "info")

	// Verify that a debug level log is logged
	Debug(ctx, "debug")
	require.Contains(t, buf.String(), "debug")
}

// Test that a logger persists fields between calls.
func testLoggerPersistence(t *testing.T, newLogger func(*bytes.Buffer) Logger) {
	buf := bytes.Buffer{}
	ctx := PutLogger(context.Background(), newLogger(&buf).WithLevel(DebugLevel))

	// Verify that a string is persisted within the context
	ctx = WithStr(ctx, "string_key", "string_value")
	Info(ctx, "string_event")
	require.Contains(t, buf.String(), "string_event")
	require.Contains(t, buf.String(), "string_key")
	require.Contains(t, buf.String(), "string_value")

	// Verify that an int is persisted within the context
	ctx = WithInt(ctx, "int_key", 12)
	Info(ctx, "int_event")
	require.Contains(t, buf.String(), "int_event")
	require.Contains(t, buf.String(), "int_key")
	require.Contains(t, buf.String(), "12")

	// Verify that a duration is persisted within the context
	ctx = WithDuration(ctx, "duration_key", time.Hour)
	Info(ctx, "duration_event")
	require.Contains(t, buf.String(), "duration_event")
	require.Contains(t, buf.String(), "duration_key")
	require.True(t, strings.Contains(buf.String(), "3600000") || strings.Contains(buf.String(), "1h0m0s"))
}

// Test that a logger logs, or ignores, log levels appropriately.
func testLoggerLevel(t *testing.T, newLogger func(*bytes.Buffer) Logger) {
	buf := bytes.Buffer{}
	logger := newLogger(&buf).WithLevel(DebugLevel)
	ctx := PutLogger(context.Background(), logger)

	// Verify that a debug level log is ignored with a lower level set
	ctx = PutLogger(ctx, logger.WithLevel(InfoLevel))
	Debug(ctx, "ignore-debug")
	require.NotContains(t, buf.String(), "ignore-debug")

	// Verify that an info level log is ignored with a lower level set
	ctx = PutLogger(ctx, logger.WithLevel(ErrorLevel))
	Debug(ctx, "ignore-info")
	require.NotContains(t, buf.String(), "ignore-info")
}

// Test the usage of the native logger can be interleaved with the wrapped logger.
func testLoggerInterleave(t *testing.T,
	newLogger func(*bytes.Buffer) Logger,
	nativePersist func(context.Context, string, string) context.Context,
	nativeLog func(context.Context, *bytes.Buffer, string)) {
	buf := bytes.Buffer{}
	wrapped := newLogger(&buf).WithLevel(DebugLevel)
	ctx := PutLogger(context.Background(), wrapped)
	ctx = nativePersist(ctx, "native-key", "native-value")
	ctx = WithStr(ctx, "wrapped-key", "wrapped-value")
	Info(ctx, "info")
	require.Contains(t, buf.String(), "native-key")
	require.Contains(t, buf.String(), "wrapped-key")
	require.Contains(t, buf.String(), "info")

	buf = bytes.Buffer{}
	wrapped = newLogger(&buf).WithLevel(DebugLevel)
	ctx = PutLogger(context.Background(), wrapped)
	ctx = nativePersist(ctx, "native-key", "native-value")
	ctx = WithStr(ctx, "wrapped-key", "wrapped-value")
	nativeLog(ctx, &buf, "info")
	require.Contains(t, buf.String(), "native-key")
	require.Contains(t, buf.String(), "wrapped-key")
	require.Contains(t, buf.String(), "info")
}
