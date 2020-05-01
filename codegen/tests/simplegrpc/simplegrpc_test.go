package simplegrpc

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/codegen/tests/simple"
	pb "github.com/anz-bank/sysl-go/codegen/tests/simplepb"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/anz-bank/sysl-go/validator"
	"github.com/prometheus/client_golang/prometheus"
	tlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type ServerHolder struct {
	svr *grpc.Server
}

func (s *ServerHolder) RegisterServer(ctx context.Context, server *grpc.Server) {
	s.svr = server
}

func (s ServerHolder) Config() validator.Validator {
	return nil
}

func (s ServerHolder) Name() string {
	return ""
}

type TestGrpcHandler struct {
	cfg      config.CommonServerConfig
	handlers []handlerinitialiser.GrpcHandlerInitialiser
}

func (h *TestGrpcHandler) EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser {
	return h.handlers
}

func (h *TestGrpcHandler) GrpcAdminServerConfig() *config.CommonServerConfig {
	return &h.cfg
}

func (h *TestGrpcHandler) GrpcPublicServerConfig() *config.CommonServerConfig {
	return &h.cfg
}

func localServerConfig() config.CommonServerConfig {
	return config.CommonServerConfig{
		HostName: "localhost",
		Port:     8888,
	}
}

func localAdminServerConfig(adminPort int) *config.CommonHTTPServerConfig {
	return &config.CommonHTTPServerConfig{
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     adminPort,
		},
		BasePath: "/admin",
	}
}

type Callbacks struct {
	timeout time.Duration
}

func (c Callbacks) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

func (c Callbacks) Config() validator.Validator {
	return nil
}

func GetStuffStub(ctx context.Context, req *pb.GetStuffRequest, client GetStuffClient) (*pb.GetStuffResponse, error) {
	resp := pb.GetStuffResponse{
		Data: []*pb.Item{{Name: "test"}},
	}
	return &resp, nil
}

func connectAndCheckReturn(t *testing.T, securityOption grpc.DialOption) {
	conn, err := grpc.Dial("localhost:8888", securityOption, grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewSimpleGrpcClient(conn)
	resp, err := client.GetStuff(context.Background(), &pb.GetStuffRequest{InnerStuff: "test"})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetData()[0].GetName())
}

func TestValidRequestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	logger, _ := tlog.NewNullLogger()

	cb := Callbacks{
		timeout: 1 * time.Second,
	}

	si := GrpcServiceInterface{
		GetStuff: GetStuffStub,
	}

	client := simple.NewClient(server.Client(), server.URL)
	serviceHandler := NewGrpcServiceHandler(cb, &si, client)

	serverHolder := ServerHolder{}

	handlerManager := TestGrpcHandler{
		cfg:      localServerConfig(),
		handlers: []handlerinitialiser.GrpcHandlerInitialiser{serviceHandler, &serverHolder},
	}

	serverError := make(chan error)

	go func() {
		err := core.Server(context.Background(), "test", nil, nil, &handlerManager, logger, nil)
		serverError <- err
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
	serverHolder.svr.GracefulStop()
	require.NoError(t, <-serverError)
}

func TestValidRequestResponseAdminServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	adminPort := rand.Intn(999) + 8000

	logger, _ := tlog.NewNullLogger()
	adminConfig := config.LibraryConfig{
		Log:         config.LogConfig{},
		Profiling:   false,
		AdminServer: localAdminServerConfig(adminPort),
	}

	cb := Callbacks{
		timeout: 1 * time.Second,
	}

	si := GrpcServiceInterface{
		GetStuff: GetStuffStub,
	}

	client := simple.NewClient(server.Client(), server.URL)
	serviceHandler := NewGrpcServiceHandler(cb, &si, client)

	serverHolder := ServerHolder{}

	handlerManager := TestGrpcHandler{
		cfg:      localServerConfig(),
		handlers: []handlerinitialiser.GrpcHandlerInitialiser{serviceHandler, &serverHolder},
	}

	serverError := make(chan error)

	go func() {
		err := core.Server(context.Background(), "admin", &adminConfig, nil, &handlerManager, logger, prometheus.NewRegistry())
		serverError <- err
	}()

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/status", adminPort), nil)
	require.Nil(t, err)
	resp, err1 := http.DefaultClient.Do(req)
	require.Nil(t, err1)
	defer resp.Body.Close()
}
