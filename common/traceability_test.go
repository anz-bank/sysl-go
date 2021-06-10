package common

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

//nolint:maligned // This is just a test and it's easier to read in this order
type testData struct {
	incomingHeaderID             string
	insertDifferentHeaderIDToCfg bool
	reqid                        *string
	expectWarning                bool
}

var standardHeaderID = defaultIncomingHeaderForID
var differentHeaderID = "TraceID"

var tests = []testData{
	{standardHeaderID, false, NewString(""), true},
	{standardHeaderID, false, NewString("652817bc-ee0c-40e3-936c-fa74aea0ad49"), false},
	{standardHeaderID, true, NewString("652817bc-ee0c-40e3-936c-fa74aea0ad49"), true},
	{differentHeaderID, true, NewString("652817bc-ee0c-40e3-936c-fa74aea0ad49"), false},
	{standardHeaderID, false, NewString("652817bc-AB0C-40E3-936C-fA74AAA0AA49"), false},
	{standardHeaderID, false, nil, true},
}

func TestTraceabilityMiddleware(t *testing.T) {
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("TestTraceabilityMiddleware#%d", i), func(t *testing.T) {
			ctx, logger := testutil.NewTestContextWithLogger()

			if tt.insertDifferentHeaderIDToCfg {
				cfg := config.DefaultConfig{}
				cfg.Library.Trace.IncomingHeaderForID = differentHeaderID
				ctx = config.PutDefaultConfig(ctx, &cfg)
			}

			mware := TraceabilityMiddleware
			body := bytes.NewBufferString("test")
			req, err := http.NewRequest("GET", "localhost/", body)
			require.Nil(t, err)
			req = req.WithContext(ctx)
			if tt.reqid != nil {
				req.Header.Add(tt.incomingHeaderID, *tt.reqid)
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
