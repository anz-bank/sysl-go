package core

import (
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func Recoverer(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					logger.Logf(logrus.ErrorLevel, "Panic: %+v\n", rvr)
					logger.Logf(logrus.ErrorLevel, "%s", debug.Stack())

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
