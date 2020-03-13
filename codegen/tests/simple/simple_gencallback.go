package simple

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go-comms/common"

	"github.com/anz-bank/sysl-go-comms/validator"

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

func (c Callback) HandleError(ctx context.Context, w http.ResponseWriter, kind common.Kind, message string, cause error) {
	se := common.CreateError(ctx, kind, message, cause)

	httpError := common.HandleError(ctx, se)

	httpError.WriteError(ctx, w)
}

func (c Callback) AddMiddleware(ctx context.Context, r chi.Router) {
}

func (c Callback) BasePath() string {
	return "/"
}
