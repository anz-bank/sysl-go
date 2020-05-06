package simple

import (
	"context"
	"time"

	"github.com/anz-bank/sysl-go/common"

	"github.com/go-chi/chi"
)

type Callback struct {
	returnedMapError *common.HTTPError
}
type CallbackWithMapError struct{}
type Config struct{}

func (c Config) Validate() error {
	return nil
}

func (c Callback) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 1*time.Second)
}

func (c Callback) Config() interface{} {
	return Config{}
}

func (c Callback) MapError(ctx context.Context, cause error) *common.HTTPError {
	return c.returnedMapError
}

func (c Callback) AddMiddleware(ctx context.Context, r chi.Router) {
}

func (c Callback) BasePath() string {
	return "/"
}
