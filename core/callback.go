package core

import (
	"context"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/validator"
	"github.com/go-chi/chi"
)

// The generic callback
type GenCallback interface {
	// Config returns a structure representing the server config
	// This is returned from the status endpoint
	Config() validator.Validator
	// DownstreamTimeoutContext add the desired timeout duration to the context for downstreams
	// A separate service timeout (usually greater than the downstream) should also be in
	// place to automatically respond to callers
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}

// RestGenCallback extends GenCallback and is used by `sysl-go` to call hand-crafted code
type RestGenCallback interface {
	GenCallback
	// AddMiddleware allows hand-crafted code to add middleware to the router
	AddMiddleware(ctx context.Context, r chi.Router)
	// BasePath allows hand-crafted code to set the base path for the Router
	BasePath() string
	// MapError maps an error to an HTTPError in instances where custom error mapping is required. Return nil to perform default error mapping; defined as:
	// 1. CustomError.HTTPError if the original error is a CustomError, otherwise
	// 2. common.MapError
	MapError(ctx context.Context, err error) *common.HTTPError
}

// GrpcGenCallback extends the generic callback GenCallback
type GrpcGenCallback interface {
	GenCallback
}
