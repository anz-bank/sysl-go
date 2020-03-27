package handlerinitialiser

import (
	"context"

	"github.com/anz-bank/sysl-go/config"
	"github.com/go-chi/chi"
	"google.golang.org/grpc"
)

type HandlerInitialiser interface {
	WireRoutes(ctx context.Context, r chi.Router)
	Name() string                  // Human-friendly name of the service
	Config() config.UpstreamConfig // Reference to config for this service.
}

type GrpcHandlerInitialiser interface {
	RegisterServer(ctx context.Context, server *grpc.Server)
}
