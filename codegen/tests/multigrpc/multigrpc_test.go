package multigrpc

import (
	"context"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/codegen/tests/cards"
	pb "github.com/anz-bank/sysl-go/codegen/tests/cardspb"
	"github.com/anz-bank/sysl-go/codegen/tests/wallet"
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

func GetCardsStub(ctx context.Context, req *pb.GetCardsRequest, client cards.GetCardsClient) (*pb.GetCardsResponse, error) {
	respString := "error"

	if req.PersonaId == "1" {
		respString = "Card1"
	}

	resp := pb.GetCardsResponse{
		Cards: []*pb.Card{{Name: respString}},
	}
	return &resp, nil
}
func AppleStub(ctx context.Context, req *pb.AppleRequest, client wallet.AppleClient) (*pb.AppleResponse, error) {
	respString := "error"

	if req.Fpan == "1" {
		respString = "Activate!"
	}

	resp := pb.AppleResponse{
		ActivationData: respString,
	}
	return &resp, nil
}

func TestEndToEndValidRequestResponse(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	cb := Callbacks{
		timeout: 1 * time.Second,
	}

	walletSH := wallet.NewGrpcServiceHandler(cb, &wallet.GrpcServiceInterface{Apple: AppleStub})
	cardsSH := cards.NewGrpcServiceHandler(cb, &cards.GrpcServiceInterface{GetCards: GetCardsStub})
	serverHolder := ServerHolder{}

	handlerManager := TestGrpcHandler{
		cfg: localServerConfig(),
		handlers: []handlerinitialiser.GrpcHandlerInitialiser{
			cardsSH, walletSH, &serverHolder},
	}

	serverError := make(chan error)

	go func() {
		err := core.Server(context.Background(), "test",
			nil, &handlerManager, logger, nil)
		serverError <- err
	}()

	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	{
		client := pb.NewCardsClient(conn)
		resp, err := client.GetCards(context.Background(), &pb.GetCardsRequest{PersonaId: "1"})
		require.NoError(t, err)
		require.Equal(t, "Card1", resp.GetCards()[0].GetName())
	}
	{
		client := pb.NewWalletClient(conn)
		resp, err := client.Apple(context.Background(), &pb.AppleRequest{Fpan: "1"})
		require.NoError(t, err)
		require.Equal(t, "Activate!", resp.GetActivationData())
	}

	serverHolder.svr.GracefulStop()
	require.NoError(t, <-serverError)
}
