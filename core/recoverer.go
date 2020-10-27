package core

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/anz-bank/pkg/log"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
	})
}
