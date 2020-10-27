package health

import (
	"context"

	health "github.com/anz-bank/pkg/health"
	"google.golang.org/grpc"
)

type Server struct {
	*health.Server
}

func (s *Server) RegisterServer(ctx context.Context, server *grpc.Server) {
	s.GRPC.RegisterWith(server)
}

func NewServer() (*Server, error) {
	s, err := health.NewServer()
	if err != nil {
		return nil, err
	}
	return &Server{s}, nil
}
