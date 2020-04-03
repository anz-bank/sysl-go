package dbendpoints

import (
	"context"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anz-bank/sysl-go/common"

	"github.com/anz-bank/sysl-go/validator"

	"github.com/go-chi/chi"
)

type Callback struct{}
type Config struct{}

func (c Config) Validate() error {
	return nil
}

func (c Callback) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 1*time.Second)
}

func (c Callback) Config() validator.Validator {
	return Config{}
}

func (c Callback) MapError(ctx context.Context, cause error) *common.HTTPError {
	return nil
}

func (c Callback) AddMiddleware(ctx context.Context, r chi.Router) {
}

func (c Callback) BasePath() string {
	return "/"
}
