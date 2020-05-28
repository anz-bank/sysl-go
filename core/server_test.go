package core

import (
	"context"
	"testing"

	tlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestServerStartWithoutHandlerPanic(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	serverError := make(chan error)

	require.Panics(t, func() {
		err := Server(context.Background(), "test", nil, nil, logger, nil)
		serverError <- err
	})
}
