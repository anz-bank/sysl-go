package common

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/anz-bank/pkg/log"
	"github.com/stretchr/testify/mock"
)

func NewString(s string) *string {
	return &s
}

func NewBool(b bool) *bool {
	return &b
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

func NewTestCoreRequestContext() (*logrus.Logger, *test.Hook, context.Context) {
	logger, hook := test.NewNullLogger()

	ctx := NewTestCoreRequestContextWithLogger(logger)

	return logger, hook, ctx
}

func NewTestCoreRequestContextWithLogger(logger *logrus.Logger) context.Context {
	ctx := context.WithValue(context.Background(), coreRequestContextKey{},
		&coreRequestContext{
			logger: logger,
			entry:  logger.WithField("traceId", uuid.New().String()),
		})

	return ctx
}

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
	ctx := log.WithConfigs(log.AddHooks(&loghook)).Onto(context.Background())
	return ctx, &loghook
}
