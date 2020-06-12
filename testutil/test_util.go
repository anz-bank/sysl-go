package testutil

import (
	"context"

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
	loghook := TestHook{}
	ctxWithValue := context.WithValue(context.Background(), logconfig.IsVerboseLoggingKey{},
		&logconfig.IsVerboseLogging{
			Flag: true,
		})
	ctx := log.WithConfigs(log.AddHooks(&loghook)).Onto(ctxWithValue)
	return ctx, &loghook
}
