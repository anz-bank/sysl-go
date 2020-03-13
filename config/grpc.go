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
