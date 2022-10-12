package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"rest_custom_middleware/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"github.com/go-chi/chi"
)

type AppConfig struct{}

func GetPingList(ctx context.Context, _ *pingpong.GetPingListRequest) (*pingpong.Pong, error) {
	preamble := ctx.Value("preamble").(string)
	fruit := ctx.Value("fruit").(string)
	return &pingpong.Pong{
		Data: fmt.Sprintf("%s %s", preamble, fruit),
	}, nil
}

// withValue(key, val) returns a middleware that injects a key, val item into the request context.
func withValue(key, val string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), key, val))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func GetWriteerrorcallbackList(ctx context.Context, _ *pingpong.GetWriteerrorcallbackListRequest) error {
	preamble := ctx.Value("preamble").(string)
	fruit := ctx.Value("fruit").(string)

	return &pingpong.ErrorResponse{
		Err: fmt.Sprintf("%s %s", preamble, fruit),
	}
}

type ErrorResponseWriter struct {
	err pingpong.ErrorResponse
}

// Error fulfills the error interface.
func (e ErrorResponseWriter) Error() string {
	return e.err.Error()
}

func (e ErrorResponseWriter) WriteError(ctx context.Context, w http.ResponseWriter) bool {
	b, err := json.Marshal(e.err)
	if err != nil {
		log.Error(ctx, err, "error marshalling error response")
		return false
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusPaymentRequired)

	// Ignore write error, if any, as it is probably a client issue.
	_, _ = w.Write(b)

	return true
}

func GetErrorwriterList(ctx context.Context, _ *pingpong.GetErrorwriterListRequest) error {
	preamble := ctx.Value("preamble").(string)
	fruit := ctx.Value("fruit").(string)

	return &ErrorResponseWriter{
		pingpong.ErrorResponse{
			Err: fmt.Sprintf("%s %s", preamble, fruit),
		},
	}
}

func createService(_ context.Context, _ AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
	return &pingpong.ServiceInterface{
			GetPingList:               GetPingList,
			GetWriteerrorcallbackList: GetWriteerrorcallbackList,
			GetErrorwriterList:        GetErrorwriterList,
		}, &core.Hooks{
			AddHTTPMiddleware: func(ctx context.Context, r chi.Router) {
				r.Use(withValue("fruit", "rambutan"))
				r.Use(withValue("preamble", "once upon a time there was a"))
			},

			MapError: func(ctx context.Context, err error) *common.HTTPError {
				var serverError *common.ServerError
				if errors.As(err, &serverError) {
					var errResp *pingpong.ErrorResponse
					if serverError.Cause != nil && errors.As(serverError.Cause, &errResp) {
						ret := &common.HTTPError{HTTPCode: http.StatusTeapot}
						ret.AddField("err", errResp)

						return ret
					}
				}

				return nil
			},

			WriteError: func(ctx context.Context, w http.ResponseWriter, httpError *common.HTTPError) {
				b, err := json.Marshal(httpError.GetField("err"))
				if err != nil {
					log.Error(ctx, err, "error marshalling error response")
					b = []byte(`{"status":{"code": "1234", "description": "Unknown Error"}}`)
					httpError.HTTPCode = http.StatusInternalServerError
				}

				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.WriteHeader(httpError.HTTPCode)

				// Ignore write error, if any, as it is probably a client issue.
				_, _ = w.Write(b)
			},
		},
		nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx, createService)
}

func main() {
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something goes wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
