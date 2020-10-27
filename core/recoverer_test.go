package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestRecoverer(t *testing.T) {
	mware, hook := testutil.LoggerHookContextMiddleware()

	ts := httptest.NewServer(mware(Recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("Test")
	}))))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/", ts.URL))
	require.NoError(t, err)
	if res != nil {
		defer res.Body.Close()
	}
	require.NotEmpty(t, hook.Entries)
}
