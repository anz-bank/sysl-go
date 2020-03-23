package common

import (
	"context"
)

const (
	missingParam          = "Missing one or more of the required parameters"
	internalServerError   = "Internal Server Error"
	unauthorizedError     = "Unauthorized error"
	downstreamUnavailable = "Downstream system is unavailable"
	timeoutDownstream     = "Time out from down stream services"
	unknownError          = "Unknown Error"
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
