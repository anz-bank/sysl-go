package config

import (
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
	ServiceAddress string     `yaml:"serviceAddress" mapstructure:"serviceAddress"`
	TLS            *TLSConfig `yaml:"tls" mapstructure:"tls"`
	WithBlock      bool       `yaml:"withBlock" mapstructure:"withBlock"`
}

func NewDefaultCommonGRPCDownstreamData() *CommonGRPCDownstreamData {
	return &CommonGRPCDownstreamData{}
}

// DefaultGRPDialOptions creates []grpc.DialOption from the given config.
// If cfg is nil then NewDefaultCommonGRPCDownstreamData will be used to define
// the dial options.
func DefaultGrpcDialOptions(cfg *CommonGRPCDownstreamData) ([]grpc.DialOption, error) {
	if cfg == nil {
		cfg = NewDefaultCommonGRPCDownstreamData()
	}
	var opts []grpc.DialOption
	if cfg.TLS != nil {
		tlsConfig, err := MakeTLSConfig(cfg.TLS)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if cfg.WithBlock {
		opts = append(opts, grpc.WithBlock())
	}
	return opts, nil
}
