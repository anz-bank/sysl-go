package core

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	test "github.com/anz-bank/sysl-go/core/testdata/proto"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const testPort = 8888

type TestServer struct{}

func (*TestServer) Test(ctx context.Context, req *test.TestRequest) (*test.TestReply, error) {
	return &test.TestReply{Field1: req.GetField1()}, nil
}

func localServer() config.CommonServerConfig {
	return config.CommonServerConfig{
		HostName: "localhost",
		Port:     testPort,
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
		Port:     testPort,
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

func (h *GrpcHandler) Interceptors() []grpc.UnaryServerInterceptor {
	h.methodsCalled["Interceptors"] = true

	return []grpc.UnaryServerInterceptor{}
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
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", testPort), securityOption, grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()
	client := test.NewTestServiceClient(conn)
	resp, err := client.Test(context.Background(), &test.TestRequest{Field1: "test"})
	require.NoError(t, err)
	require.Equal(t, "test", resp.GetField1())
}

func Test_makeGrpcListenFuncListens(t *testing.T) {
	ctx, _ := testutil.NewTestContextWithLoggerHook()

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := makeGrpcListenFunc(ctx, grpcServer, localServer())
	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
}

func Test_encryptionConfigUsed(t *testing.T) {
	t.Skip("Skipping as required certs not present")
	ctx, hook := testutil.NewTestContextWithLoggerHook()

	cfg := localSecureServer()

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := makeGrpcListenFunc(ctx, grpcServer, cfg)
	go func() {
		err := listener()
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

	ctx, hook := testutil.NewTestContextWithLoggerHook()

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	test.RegisterTestServiceServer(grpcServer, &TestServer{})

	listener := prepareGrpcServerListener(ctx, grpcServer, localServer())
	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", testPort), grpc.WithInsecure(), grpc.WithBlock())
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
	ctx, _ := testutil.NewTestContextWithLoggerHook()

	manager := &GrpcHandler{
		cfg: localServer(),
		reg: ServerReg{
			svr:           TestServer{},
			methodsCalled: make(map[string]bool),
		},
		methodsCalled: make(map[string]bool),
	}

	// Adapt deprecated GrpcManager type as GrpcServerManager struct
	grpcServerManager, err := newGrpcServerManagerFromGrpcManager(manager)
	require.NoError(t, err)

	listener := configurePublicGrpcServerListener(ctx, *grpcServerManager)
	require.NotNil(t, listener)

	go func() {
		err := listener()
		require.NoError(t, err)
	}()

	connectAndCheckReturn(t, grpc.WithInsecure())
	require.True(t, manager.methodsCalled["Interceptors"])
	require.True(t, manager.methodsCalled["EnabledGrpcHandlers"])
	require.True(t, manager.methodsCalled["GrpcPublicServerConfig"])
	require.True(t, manager.reg.methodsCalled["RegisterServer"])
}
