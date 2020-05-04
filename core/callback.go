package core

import (
	"context"

	"github.com/anz-bank/sysl-go/common"
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
	Config() interface{}
	// MapError maps an error to an HTTPError in instances where custom error mapping is required. Return nil to perform default error mapping; defined as:
	// 1. CustomError.HTTPError if the original error is a CustomError, otherwise
	// 2. common.MapError
	MapError(ctx context.Context, err error) *common.HTTPError
	// DownstreamTimeoutContext add the desired timeout duration to the context for downstreams
	// A separate service timeout (usually greater than the downstream) should also be in
	// place to automatically respond to callers
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}

// GrpcGenCallback is currently a subset of RestGenCallback so is defined separately for convenience
type GrpcGenCallback interface {
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}
