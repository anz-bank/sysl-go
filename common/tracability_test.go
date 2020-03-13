package common

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type testData struct {
	reqid         *string
	expectWarning bool
}

var tests = []testData{
	{NewString(""), true},
	{NewString("1234-5678910"), false},
	{nil, true},
}

func TestTraceabilityMiddleware(t *testing.T) {
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("TestTraceabilityMiddleware#%d", i), func(t *testing.T) {
			logger, loghook := test.NewNullLogger()

			mware := TraceabilityMiddleware(logger)
			body := bytes.NewBufferString("test")
			req, err := http.NewRequest("GET", "localhost/", body)
			require.Nil(t, err)
			req = req.WithContext(context.Background())
			if tt.reqid != nil {
				req.Header.Add("RequestID", *tt.reqid)
			}

			fn := mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				if tt.expectWarning {
					require.NotEmpty(t, loghook.Entries)
					require.NotEqual(t, "", loghook.LastEntry().Data["traceid"])
					require.Equal(t, loghook.LastEntry().Data["traceid"], GetTraceIDFromContext(r.Context()))
				} else {
					require.Empty(t, loghook.Entries)
					require.Equal(t, *tt.reqid, GetTraceIDFromContext(r.Context()))
				}
			}))

			fn.ServeHTTP(nil, req)
		})
	}
}
