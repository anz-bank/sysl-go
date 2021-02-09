package common

import (
	"net/http"

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
