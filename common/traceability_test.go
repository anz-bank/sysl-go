package common

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

type testData struct {
	reqid         *string
	expectWarning bool
}

var tests = []testData{
	{NewString(""), true},
	{NewString("652817bc-ee0c-40e3-936c-fa74aea0ad49"), false},
	{NewString("652817bc-AB0C-40E3-936C-fA74AAA0AA49"), false},
	{nil, true},
}

func TestTraceabilityMiddleware(t *testing.T) {
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("TestTraceabilityMiddleware#%d", i), func(t *testing.T) {
			ctx, logger := testutil.NewTestContextWithLogger()

			mware := TraceabilityMiddleware
			body := bytes.NewBufferString("test")
			req, err := http.NewRequest("GET", "localhost/", body)
			require.Nil(t, err)
			req = req.WithContext(ctx)
			if tt.reqid != nil {
				req.Header.Add("RequestID", *tt.reqid)
			}

			fn := mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				if tt.expectWarning {
					require.NotZero(t, logger.EntryCount())
				} else {
					require.Zero(t, logger.EntryCount())
					require.Equal(t, strings.ToLower(*tt.reqid), strings.ToLower(GetTraceIDFromContext(r.Context()).String()))
				}
			}))
			fn.ServeHTTP(nil, req)
		})
	}
}
