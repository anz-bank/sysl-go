package core

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/anz-bank/pkg/log"
)

func Recoverer(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				var err error
				if rvr := recover(); rvr != nil {
					switch x := rvr.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = errors.New("unknown panic")
					}
					log.Errorf(ctx, err, "Panic: %+v\n", rvr)
					log.Errorf(ctx, err, "%s", debug.Stack())

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
