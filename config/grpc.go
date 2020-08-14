package config

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ExtractGrpcServerOptions(cfg *CommonServerConfig) ([]grpc.ServerOption, error) {
	if cfg == nil || cfg.TLS == nil {
		return []grpc.ServerOption{}, nil
	}

	tlsConfig, err := MakeTLSConfig(cfg.TLS)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(tlsConfig)

	return []grpc.ServerOption{grpc.Creds(creds)}, nil
}

// CommonGRPCDownstreamData collects all the client gRPC configuration.
type CommonGRPCDownstreamData struct {
	ServiceAddress string     `yaml:"serviceAddress"`
	TLS            *TLSConfig `yaml:"tls"`
	WithBlock      bool       `yaml:"withBlock"`
}

func NewDefaultCommonGRPCDownstreamData() *CommonGRPCDownstreamData {
	return &CommonGRPCDownstreamData{}
}

// DefaultGRPCClient returns a new *grpc.ClientConn with sensible defaults, in
// particular it has a timeout set!
func DefaultGRPCClient(cfg *CommonGRPCDownstreamData) (*grpc.ClientConn, error) {
	if cfg == nil {
		cfg = NewDefaultCommonGRPCDownstreamData()
	}

	var opts []grpc.DialOption
	if cfg.TLS != nil {
		tlsConfig, err := makeSelfSignedTLSConfig(cfg.TLS)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewTLS(tlsConfig)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if cfg.WithBlock {
		opts = append(opts, grpc.WithBlock())
	}
	return grpc.Dial(cfg.ServiceAddress, opts...)
}
