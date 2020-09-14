package simplegrpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/codegen/tests/simple"
	pb "github.com/anz-bank/sysl-go/codegen/tests/simplepb"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/sethvargo/go-retry"
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

type TestGrpcHandler struct {
	cfg      config.CommonServerConfig
	handlers []handlerinitialiser.GrpcHandlerInitialiser
}

func (h *TestGrpcHandler) Interceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{}
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

type Callbacks struct {
	timeout time.Duration
}

func (c Callbacks) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

func GetStuffStub(ctx context.Context, req *pb.GetStuffRequest, client GetStuffClient) (*pb.GetStuffResponse, error) {
	resp := pb.GetStuffResponse{
		Data: []*pb.Item{{Name: "test"}},
	}
	return &resp, nil
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
		err := core.Server(context.Background(), "test", nil, &handlerManager, logger, nil) //nolint
		serverError <- err
	}()

	getStuff := func(ctx context.Context, req *pb.GetStuffRequest) (*pb.GetStuffResponse, error) {
		conn, err := grpc.DialContext(ctx, "localhost:8888", grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client := pb.NewSimpleGrpcClient(conn)
		return client.GetStuff(ctx, req)
	}

	// Setup: wait for the service to come up
	ctx := context.Background()
	backoff, err := retry.NewFibonacci(10 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := getStuff(ctx, &pb.GetStuffRequest{InnerStuff: "test"})
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	require.Nil(t, err)

	// Test the service
	response, err := getStuff(ctx, &pb.GetStuffRequest{InnerStuff: "test"})
	require.Nil(t, err)
	require.Equal(t, "test", response.GetData()[0].GetName())

	serverHolder.svr.GracefulStop()
	require.NoError(t, <-serverError)
}

func TestBuildRestHandlerInitialiser(t *testing.T) {
	t.Parallel()
	testConfig := NewDefaultConfig()
	downstreamConfig := testConfig.GenCode.Downstream.(*DownstreamConfig)
	downstreamConfig.Simple.ServiceURL = "http://localhost:8080/simple"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	cb := Callbacks{
		timeout: 1 * time.Second,
	}

	si := GrpcServiceInterface{
		GetStuff: GetStuffStub,
	}

	client := simple.NewClient(server.Client(), server.URL)
	serviceHandler := NewGrpcServiceHandler(cb, &si, client)
	clients, err := BuildDownstreamClients(&testConfig)
	handler := BuildGrpcHandlerInitialiser(si, cb, clients)
	require.Nil(t, err)
	grpcRouter, ok := (handler).(*GrpcServiceHandler)
	require.True(t, ok)
	reflect.DeepEqual(cb, grpcRouter.genCallback)
	reflect.DeepEqual(serviceHandler, grpcRouter.serviceInterface)
}

func TestBuildDownstreamClients(t *testing.T) {
	t.Parallel()
	testConfig := NewDefaultConfig()
	downstreamConfig := testConfig.GenCode.Downstream.(*DownstreamConfig)
	downstreamConfig.Simple.ServiceURL = "http://localhost:8080/simple"
	handlers, err := BuildDownstreamClients(&testConfig)
	require.Nil(t, err)
	require.NotNil(t, handlers.simpleClient)
}

func TestClientIsService(t *testing.T) {
	t.Parallel()
	client := simple.NewClient(nil, "")
	var i interface{} = client
	_, ok := i.(simple.Service)
	require.True(t, ok)
}
