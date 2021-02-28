package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	pingpong "rest_custom_middleware/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"github.com/go-chi/chi"
)

type AppConfig struct{}

func GetPingList(ctx context.Context, req *pingpong.GetPingListRequest) (*pingpong.Pong, error) {
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

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
			return &pingpong.ServiceInterface{
					GetPingList: GetPingList,
				}, &core.Hooks{
					AddHTTPMiddleware: func(ctx context.Context, r chi.Router) {
						r.Use(withValue("fruit", "rambutan"))
						r.Use(withValue("preamble", "once upon a time there was a"))
					},
				},
				nil
		},
	)
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
