package testutil

import (
	"context"
	"net/http"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/logconfig"
)

type TestHook struct {
	Entries []log.LogEntry
}

func (t *TestHook) OnLogged(entry *log.LogEntry) error {
	t.Entries = append(t.Entries, *entry)
	return nil
}

func (t *TestHook) LastEntry() *log.LogEntry {
	i := len(t.Entries) - 1
	if i < 0 {
		return nil
	}
	return &t.Entries[i]
}

func NewTestContextWithLoggerHook() (context.Context, *TestHook) {
	hook := &TestHook{}
	return TestContextWithLoggerHook(context.Background(), hook), hook
}

func TestContextWithLoggerHook(ctx context.Context, hook *TestHook) context.Context {
	ctx = logconfig.SetVerboseLogging(ctx, true)
	ctx = log.WithConfigs(log.SetVerboseMode(true)).Onto(ctx)
	ctx = log.WithConfigs(log.AddHooks(hook)).Onto(ctx)
	return ctx
}

func LoggerHookContextMiddleware() (func(next http.Handler) http.Handler, *TestHook) {
	hook := &TestHook{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(TestContextWithLoggerHook(r.Context(), hook))
			next.ServeHTTP(w, r)
		})
	}, hook
}
