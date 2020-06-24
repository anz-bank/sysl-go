package core

import (
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestServerStartWithoutHandlerPanic(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetReportCaller(true)
	serverError := make(chan error)

	require.Panics(t, func() {
		err := NewServerParams(context.Background(), "test", WithLogrusLogger(logger)).Start()
		serverError <- err
	})
	logValue := hook.LastEntry().Data["error_message"]
	require.True(t, strings.Contains(logValue.(string), "REST and gRPC servers cannot both be nil"))
	caller := hook.LastEntry().Data["caller"]
	require.True(t, strings.Contains(caller.(string), "server.go"))
}
