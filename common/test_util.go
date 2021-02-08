package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/config"

	"github.com/anz-bank/sysl-go/log"

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

func NewTestCoreRequestContext() (log.Logger, context.Context) {
	logger := log.NewDefaultLogger()
	ctx := log.PutLogger(context.Background(), logger)
	ctx = config.PutDefaultConfig(ctx, &config.DefaultConfig{})
	return logger, ctx
}
