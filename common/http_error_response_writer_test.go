package common

import (
	"context"
	"io"
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

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.err.WriteError(ctx, w)
			resp := w.Result()
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)

			require.Equal(t, tt.statusCode, resp.StatusCode)
			require.Equal(t, tt.body, string(b))
			require.Equal(t, "application/json;charset=UTF-8", resp.Header.Get("Content-Type"))
		})
	}
}

func TestHttpError_WriteErrorWithExtraFields(t *testing.T) {
	logger, _ := test.NewNullLogger()
	ctx := LoggerToContext(context.Background(), logger, logger.WithField("test", "test"))
	err := HTTPError{
		HTTPCode:    400,
		Code:        "1234",
		Description: "Missing one or more of the required parameters",
	}
	body := `{"status":{"aaa":123,"code":"1234","description":"Missing one or more of the required parameters","zzz":"hello"}}`
	statusCode := http.StatusBadRequest
	err.AddField("aaa", 123)
	err.AddField("zzz", "hello")

	w := httptest.NewRecorder()
	err.WriteError(ctx, w)
	resp := w.Result()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	require.Equal(t, statusCode, resp.StatusCode)
	require.Equal(t, body, string(b))
	require.Equal(t, "application/json;charset=UTF-8", resp.Header.Get("Content-Type"))
}

func TestHttpError_WriteErrorWithoutCodeAndWithExtraFields(t *testing.T) {
	logger, _ := test.NewNullLogger()
	ctx := LoggerToContext(context.Background(), logger, logger.WithField("test", "test"))
	err := HTTPError{
		HTTPCode:    400,
		Description: "Missing one or more of the required parameters",
	}
	body := `{"status":{"description":"Missing one or more of the required parameters","statusCode":123}}`
	statusCode := http.StatusBadRequest
	err.AddField("statusCode", 123)

	w := httptest.NewRecorder()
	err.WriteError(ctx, w)
	resp := w.Result()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	require.Equal(t, statusCode, resp.StatusCode)
	require.Equal(t, body, string(b))
	require.Equal(t, "application/json;charset=UTF-8", resp.Header.Get("Content-Type"))
}
