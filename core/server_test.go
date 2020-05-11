package core

import (
	"context"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	tlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type Callbacks struct {
	timeout time.Duration
}
type TestGrpcHandler struct {
	cfg      config.CommonServerConfig
	handlers []handlerinitialiser.GrpcHandlerInitialiser
}

type TestHttpHandler struct {
	cfg      config.CommonServerConfig
	handlers []handlerinitialiser.HandlerInitialiser
}

// EnabledHandlers() []handlerinitialiser.HandlerInitialiser
// LibraryConfig() *config.LibraryConfig
// AdminServerConfig() *config.CommonHTTPServerConfig
// PublicServerConfig() *config.CommonHTTPServerConfig

func (hl *TestHttpHandler) EnabledHandlers() []handlerinitialiser.HandlerInitialiser {
	return nil
}
func (hl *TestHttpHandler) LibraryConfig() *config.LibraryConfig {
	return nil
}
func (hl *TestHttpHandler) AdminServerConfig() *config.CommonHTTPServerConfig {
	return nil
}
func (hl *TestHttpHandler) PublicServerConfig() *config.CommonHTTPServerConfig {
	return &config.CommonHTTPServerConfig{
		BasePath: "/test",
		Common: config.CommonServerConfig{
			HostName: "",
			Port:     8080,
			TLS:      nil,
		},
	}
}

func TestServerStartWithoutHandlerPanic(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	serverError := make(chan error)
	params := &ServerParams{
		Ctx:  context.Background(),
		Name: "test",
	}

	require.Panics(t, func() {
		err := params.Start(WithLogger(logger))
		serverError <- err
	})
}

func TestServerStartWithRestManagerHandler(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	restManager := &TestHttpHandler{}
	params := &ServerParams{
		Ctx:  context.Background(),
		Name: "test",
	}
	go func() {
		params.Start(WithLogger(logger), WithRestManager(restManager))
		require.NotEmpty(t, params.restManager)
	}()

}
