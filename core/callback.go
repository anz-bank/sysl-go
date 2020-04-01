package core

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/validator"
	"github.com/go-chi/chi"
)

// RestGenCallback is used by `sysl-go` to call hand-crafted code
type RestGenCallback interface {
	// AddMiddleware allows hand-crafted code to add middleware to the router
	AddMiddleware(ctx context.Context, r chi.Router)
	// BasePath allows hand-crafted code to set the base path for the Router
	BasePath() string
	// Config returns a structure representing the server config
	// This is returned from the status endpoint
	Config() validator.Validator
	// HandleError allows custom HTTP errors to be added to `sys-go` errors
	HandleError(ctx context.Context, w http.ResponseWriter, kind common.Kind, message string, cause error)
	// DownstreamTimeoutContext add the desired timeout duration to the context for downstreams
	// A separate service timeout (usually greater than the downstream) should also be in
	// place to automatically respond to callers
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}

// GrpcGenCallback is currently a subset of RestGenCallback so is defined separately for convenience
type GrpcGenCallback interface {
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}
