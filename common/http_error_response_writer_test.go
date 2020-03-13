package common

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHttpError_WriteError(t *testing.T) {
	tests := []struct {
		name       string
		err        HTTPError
		body       string
		statusCode int
	}{
		{
			name: "with description and code",
			err: HTTPError{
				HTTPCode:    400,
				Code:        "1234",
				Description: "Missing one or more of the required parameters",
			},
			body:       `{"status":{"code":"1234","description":"Missing one or more of the required parameters"}}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "without description",
			err: HTTPError{
				HTTPCode:    400,
				Code:        "1234",
				Description: "",
			},
			body:       `{"status":{"code":"1234"}}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "without code",
			err: HTTPError{
				HTTPCode:    400,
				Description: "Missing one or more of the required parameters",
			},
			body:       `{"status":{"description":"Missing one or more of the required parameters"}}`,
			statusCode: http.StatusBadRequest,
		},
	}

	logger, _ := test.NewNullLogger()
	ctx := LoggerToContext(context.Background(), logger, logger.WithField("test", "test"))

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.err.WriteError(ctx, w)
			resp := w.Result()
			defer resp.Body.Close()
			b, _ := ioutil.ReadAll(resp.Body)

			require.Equal(t, tt.statusCode, resp.StatusCode)
			require.Equal(t, tt.body, string(b))
			require.Equal(t, "application/json;charset=UTF-8", resp.Header.Get("Content-Type"))
		})
	}
}
