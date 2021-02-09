package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestRecoverer(t *testing.T) {
	mware, logger := loggerHookContextMiddleware()

	ts := httptest.NewServer(mware(Recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("Test")
	}))))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/", ts.URL))
	require.NoError(t, err)
	if res != nil {
		defer res.Body.Close()
	}
	require.NotZero(t, logger.EntryCount())
}

func loggerHookContextMiddleware() (func(next http.Handler) http.Handler, *testutil.TestLogger) {
	logger := testutil.NewTestLogger()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(log.PutLogger(r.Context(), logger))
			next.ServeHTTP(w, r)
		})
	}, logger
}
