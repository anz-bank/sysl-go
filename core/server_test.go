package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerStartWithoutHandlerPanic(t *testing.T) {
	serverError := make(chan error)

	require.Panics(t, func() {
		err := Server(context.Background(), "test", nil, nil, nil, nil)
		serverError <- err
	})
}
