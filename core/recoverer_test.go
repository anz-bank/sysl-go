package core

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
)

func TestRecoverer(t *testing.T) {
	l := logrus.New()
	buffer := bytes.Buffer{}
	l.SetOutput(&buffer)

	ts := httptest.NewServer(Recoverer(l)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("Test")
	})))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/", ts.URL))
	require.NoError(t, err)
	if res != nil {
		defer res.Body.Close()
	}

	require.True(t, strings.Contains(buffer.String(), "Panic: Test"))
}
