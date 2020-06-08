package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/common"
	"github.com/stretchr/testify/require"
)

func TestRecoverer(t *testing.T) {
	ctx, _ := common.NewTestContextWithLoggerHook()

	ts := httptest.NewServer(Recoverer(ctx)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("Test")
	})))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/", ts.URL))
	require.NoError(t, err)
	if res != nil {
		defer res.Body.Close()
	}
	require.Panics(t, nil, "Panic: Test")
}
