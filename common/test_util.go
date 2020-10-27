package common

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

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
		coreRequestContext{
			logger: logger,
			entry:  logger.WithField("traceId", uuid.New().String()),
		})

	return ctx
}
