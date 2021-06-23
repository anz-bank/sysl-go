package core

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/config"
)

func BuildDownstreamHTTPClient(ctx context.Context, serviceName string, hooks *Hooks, cfg *config.CommonDownstreamData) (client *http.Client, serviceURL string, err error) {
	if hooks != nil && hooks.HTTPClientBuilder != nil {
		client, serviceURL, err = hooks.HTTPClientBuilder(serviceName)
	} else {
		client, err = config.DefaultHTTPClient(ctx, cfg)
		if cfg != nil {
			serviceURL = cfg.ServiceURL
		}
	}

	if err != nil {
		return nil, "", err
	}

	client.Transport = common.NewLoggingRoundTripper(serviceName, client.Transport)
	if hooks != nil && hooks.DownstreamRoundTripper != nil {
		client.Transport = hooks.DownstreamRoundTripper(serviceName, serviceURL, client.Transport)
	}

	return
}

// BuildDownstreamGRPCClient creates a grpc client connection to the target indicated by cfg.ServiceAddress.
// The dial options can be customised by cfg or by hooks, see ResolveGrpcDialOptions for details. The
// serviceName is the name of the target service. This function is intended to be called from generated code.
func BuildDownstreamGRPCClient(ctx context.Context, serviceName string, hooks *Hooks, cfg *config.CommonGRPCDownstreamData) (*grpc.ClientConn, error) {
	opts, err := ResolveGrpcDialOptions(ctx, serviceName, hooks, cfg)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(cfg.ServiceAddress, opts...)
}
