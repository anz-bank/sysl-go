package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/pkg/log"
)

const (
	missingParam          = "Missing one or more of the required parameters"
	internalServerError   = "Internal Server Error"
	unauthorizedError     = "Unauthorized error"
	downstreamUnavailable = "Downstream system is unavailable"
	timeoutDownstream     = "Time out from down stream services"
	unknownError          = "Unknown Error"
)

func HandleError(ctx context.Context, w http.ResponseWriter, kind Kind, message string, cause error, httpErrorMapper func(context.Context, error) *HTTPError) {
	err := CreateError(ctx, kind, message, cause)
	log.Error(ctx, err)

	var fields []KV
	if w, ok := cause.(wrappedError); ok {
		fields = w.fields
		err = CreateError(ctx, kind, message, w.e)
	}

	httpError := resolveErrorAsHTTPError(ctx, httpErrorMapper, err)

	for _, f := range fields {
		httpError.AddField(f.K, f.V)
	}

	httpError.WriteError(ctx, w)
}

func resolveErrorAsHTTPError(ctx context.Context, httpErrorMapper func(context.Context, error) *HTTPError, err error) *HTTPError {
	var httpError *HTTPError
	if httpErrorMapper != nil {
		httpError = httpErrorMapper(ctx, err)
	}
	if httpError == nil {
		switch t := err.(type) {
		case CustomError:
			httpError = t.HTTPError(ctx)
		default:
			e := MapError(ctx, err)
			httpError = &e
		}
	}
	return httpError
}

func MapError(ctx context.Context, err error) HTTPError {
	var (
		httpCode        int
		errorCode, desc string
	)

	switch e := err.(type) {
	case ErrorKinder:
		switch e.(ErrorKinder).ErrorKind() {
		case BadRequestError:
			httpCode = 400
			errorCode = "1001"
			desc = missingParam
		case InternalError:
			httpCode = 500
			errorCode = "9998"
			desc = internalServerError
		case UnauthorizedError:
			httpCode = 401
			errorCode = "1003"
			desc = unauthorizedError
		case DownstreamUnavailableError:
			httpCode = 503
			errorCode = "1013"
			desc = downstreamUnavailable
		case DownstreamTimeoutError:
			httpCode = 504
			errorCode = "1005"
			desc = timeoutDownstream
		default:
			httpCode = 500
			errorCode = "9999"
			desc = unknownError
		}
	default:
		if ctx.Err() == context.DeadlineExceeded {
			httpCode = 504
			errorCode = "1005"
			desc = timeoutDownstream
		} else {
			httpCode = 500
			errorCode = "9999"
			desc = unknownError
		}
	}

	return HTTPError{
		HTTPCode:    httpCode,
		Code:        errorCode,
		Description: desc,
	}
}
