package common

import (
	"context"
)

func HandleError(ctx context.Context, err error) HTTPError {
	logEntry := GetLogEntryFromContext(ctx)
	logEntry.Error(err)

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
			desc = "Missing one or more of the required parameters"
		case InternalError:
			httpCode = 500
			errorCode = "9998"
			desc = "Internal Server Error"
		case UnauthorizedError:
			httpCode = 401
			errorCode = "1003"
			desc = "Unauthorized error"
		case DownstreamUnavailableError:
			httpCode = 503
			errorCode = "1013"
			desc = "Downstream system is unavailable"
		case DownstreamTimeoutError:
			httpCode = 504
			errorCode = "1005"
			desc = "Time out from down stream services"
		default:
			httpCode = 500
			errorCode = "9999"
			desc = "Unknown Error"
		}
	default:
		if ctx.Err() == context.DeadlineExceeded {
			httpCode = 504
			errorCode = "1005"
			desc = "Time out from down stream services"
		} else {
			httpCode = 500
			errorCode = "9999"
			desc = "Unknown Error"
		}
	}

	return HTTPError{
		HTTPCode:    httpCode,
		Code:        errorCode,
		Description: desc,
	}
}
