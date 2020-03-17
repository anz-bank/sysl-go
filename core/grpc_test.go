package core

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	test "github.com/anz-bank/sysl-go/core/testdata/proto"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/sirupsen/logrus"
	tlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

type TestServer struct{}

func (*TestServer) Test(ctx context.Context, req *test.TestRequest) (*test.TestReply, error) {
	return &test.TestReply{Field1: req.GetField1()}, nil
}

func localServer() config.CommonServerConfig {
	return config.CommonServerConfig{
		HostName: "localhost",
		Port:     8888,
	}
}

func localSecureServer() config.CommonServerConfig {
	minVer := "1.2"
	maxVer := "1.3"
	certPath := "testdata/creds/server1.pem"
	keyPath := "testdata/creds/server1.key"
	clientAuth := "NoClientCert"
	ciphers := []string{"TLS_RSA_WITH_AES_256_CBC_SHA"}
	return config.CommonServerConfig{
		HostName: "localhost",
		Port:     8888,
		TLS: &config.TLSConfig{
			MinVersion: &minVer,
			MaxVersion: &maxVer,
			ClientAuth: &clientAuth,
			Ciphers:    ciphers,
			ServerIdentity: &config.ServerIdentityConfig{
				CertKeyPair: &config.CertKeyPair{
					CertPath: &certPath,
					KeyPath:  &keyPath,
				},
			},
		},
	}
}

type ServerReg struct {
	svr           TestServer
	methodsCalled map[string]bool
}

func (r *ServerReg) RegisterServer(ctx context.Context, server *grpc.Server) {
	r.methodsCalled["RegisterServer"] = true
	test.RegisterTestServiceServer(server, &r.svr)
}

type GrpcHandler struct {
	cfg           config.CommonServerConfig
	reg           ServerReg
	methodsCalled map[string]bool
}

func (h *GrpcHandler) EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser {
	h.methodsCalled["EnabledGrpcHandlers"] = true

	return []handlerinitialiser.GrpcHandlerInitialiser{&h.reg}
}

func (h *GrpcHandler) GrpcAdminServerConfig() *config.CommonServerConfig {
	h.methodsCalled["GrpcAdminServerConfig"] = true
	return &h.cfg
}

func (h *GrpcHandler) GrpcPublicServerConfig() *config.CommonServerConfig {
	h.methodsCalled["GrpcPublicServerConfig"] = true
	return &h.cfg
}

func connectAndCheckReturn(t *testing.T, securityOption grpc.DialOption) {
	conn, err := grpc.Dial("localhost:8888", securityOption, grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	client := test.NewTestServiceClient(conn)
	resp, err := client.Test(context.Background(), &test.TestRequest{Field1: "test"})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetField1())
}

func Test_makeGrpcListenFuncListens(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := makeGrpcListenFunc(grpcServer, logger, localServer())
	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
}

func Test_encryptionConfigUsed(t *testing.T) {
	t.Skip("Skipping as required certs not present")
	logger, hook := tlog.NewNullLogger()
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(logger.WriterLevel(logrus.InfoLevel),
		logger.WriterLevel(logrus.WarnLevel), logger.WriterLevel(logrus.ErrorLevel)))

	cfg := localSecureServer()

	grpcServer, err := newGrpcServer(&cfg, logger)
	require.NoError(t, err)
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := makeGrpcListenFunc(grpcServer, logger, cfg)
	go func() {
		err = listener()
		require.NoError(t, err)
	}()

	creds, err := credentials.NewClientTLSFromFile("testdata/creds/ca.pem", "x.test.youtube.com")
	require.NoError(t, err)

	connectAndCheckReturn(t, grpc.WithTransportCredentials(creds))
	for _, entry := range hook.Entries {
		t.Log(entry.Message)
	}
}

func Test_serverUsesGivenLogger(t *testing.T) {
	os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "99")

	logger, hook := tlog.NewNullLogger()

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := prepareGrpcServerListener(logger, grpcServer, localServer())
	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()

	var connecting bool
	cre := regexp.MustCompile(`ClientConn switching balancer`)
	for _, entry := range hook.Entries {
		if connecting {
			break
		}
		connecting = cre.MatchString(entry.Message)
	}
	require.True(t, connecting)
}

func Test_libMakesCorrectHandlerCalls(t *testing.T) {
	logger, _ := tlog.NewNullLogger()

	handler := GrpcHandler{
		cfg: localServer(),
		reg: ServerReg{
			svr:           TestServer{},
			methodsCalled: make(map[string]bool),
		},
		methodsCalled: make(map[string]bool),
	}

	listener, err := configurePublicGrpcServerListener(context.Background(), &handler, logger)
	require.NoError(t, err)
	require.NotNil(t, listener)

	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
	require.True(t, handler.methodsCalled["EnabledGrpcHandlers"])
	require.True(t, handler.methodsCalled["GrpcPublicServerConfig"])
	require.True(t, handler.reg.methodsCalled["RegisterServer"])
}

func Test_NewGrpcServerWithValidConfig(t *testing.T) {
	cfg := config.CommonServerConfig{
		HostName: "host",
		Port:     3000,
	}
	logger, _ := tlog.NewNullLogger()
	newServer, err := newGrpcServer(&cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, newServer)
}
