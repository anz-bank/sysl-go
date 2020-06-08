package internal

import (
	"context"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/logconfig"
)

type TestHook struct {
	Entries []log.LogEntry
}

func (h *TestHook) OnLogged(entry *log.LogEntry) error {
	h.Entries = append(h.Entries, *entry)
	return nil
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
