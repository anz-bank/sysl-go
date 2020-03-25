package simplegrpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/codegen/tests/simple"
	pb "github.com/anz-bank/sysl-go/codegen/tests/simplegrpc/simplepb"
	"github.com/anz-bank/sysl-go/codegen/tests/simplegrpc/simplesg"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
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

func GetStuffStub(ctx context.Context, req *pb.GetStuffRequest, client simplesg.GetStuffClient) (*pb.GetStuffResponse, error) {
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

	si := simplesg.GrpcServiceInterface{
		GetStuff: GetStuffStub,
	}

	client := simple.NewClient(server.Client(), server.URL)
	serviceHandler := simplesg.NewGrpcServiceHandler(cb, &si, client)

	serverHolder := ServerHolder{}

	handlerManager := TestGrpcHandler{
		cfg:      localServerConfig(),
		handlers: []handlerinitialiser.GrpcHandlerInitialiser{serviceHandler, &serverHolder},
	}

	serverError := make(chan error)

	go func() {
		err := core.Server(context.Background(), "test", nil, &handlerManager, logger, nil, nil)
		serverError <- err
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
	serverHolder.svr.GracefulStop()
	require.NoError(t, <-serverError)
}
