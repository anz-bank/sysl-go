package core

import (
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"google.golang.org/grpc"

	"github.com/anz-bank/sysl-go/config"
)

func BuildDownstreamHTTPClient(serviceName string, hooks *Hooks, cfg *config.CommonDownstreamData) (*http.Client, error) {
	client, err := config.DefaultHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		client.Timeout = time.Minute
	}

	client.Transport = common.NewLoggingRoundTripper(serviceName, client.Transport)
	if hooks.DownstreamRoundTripper != nil {
		serviceURL := ""
		if cfg != nil {
			serviceURL = cfg.ServiceURL
		}
		client.Transport = hooks.DownstreamRoundTripper(serviceName, serviceURL, client.Transport)
	}

	return client, nil
}

// BuildDownstreamGRPCClient creates a grpc client connection to the target indicated by cfg.ServiceAddress.
// The dial options can be customised by cfg or by hooks, see ResolveGrpcDialOptions for details. The
// serviceName is the name of the target service. This function is intended to be called from generated code.
func BuildDownstreamGRPCClient(serviceName string, hooks *Hooks, cfg *config.CommonGRPCDownstreamData) (*grpc.ClientConn, error) {
	opts, err := ResolveGrpcDialOptions(serviceName, hooks, cfg)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(cfg.ServiceAddress, opts...)
}
