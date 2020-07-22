package core

import (
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"google.golang.org/grpc"

	"github.com/anz-bank/sysl-go/config"
)

func BuildDownstreamHTTPClient(serviceName string, cfg *config.CommonDownstreamData) (*http.Client, error) {
	if cfg == nil {
		return buildDefaultHTTPClient(serviceName)
	}

	client, err := config.DefaultHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	client.Transport = common.NewLoggingRoundTripper(serviceName, client.Transport)

	return client, nil
}

func buildDefaultHTTPClient(serviceName string) (*http.Client, error) {
	client, err := config.DefaultHTTPClient(nil)
	if err != nil {
		return nil, err
	}
	client.Timeout = time.Minute
	client.Transport = common.NewLoggingRoundTripper(serviceName, client.Transport)

	return client, nil
}

func BuildDownstreamGRPCClient(_ string, cfg *config.CommonGRPCDownstreamData) (*grpc.ClientConn, error) {
	return config.DefaultGRPCClient(cfg)
}
