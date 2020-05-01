package handlerinitialiser

import (
	"context"

	"github.com/go-chi/chi"
	"google.golang.org/grpc"
)

type HandlerInitialiser interface {
	WireRoutes(ctx context.Context, r chi.Router)
	Name() string        // Human-friendly name of the service
	Config() interface{} // Reference to config for this service.
}

type GrpcHandlerInitialiser interface {
	RegisterServer(ctx context.Context, server *grpc.Server)
}
