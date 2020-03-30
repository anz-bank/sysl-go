package common

import (
	"context"
	"fmt"
	"strconv"
)

type CustomError map[string]string

func (e CustomError) Error() string {
	return fmt.Sprintf("%s(%v)", e["name"], e)
}

//nolint:funlen // TODO: Put http error code as a constant
func (e CustomError) HTTPError(ctx context.Context) HTTPError {
	httpStatusString := getOrDefault(e, "http_status", "500")
	httpStatus, err := strconv.Atoi(httpStatusString)
	if err != nil {
		logEntry := GetLogEntryFromContext(ctx)
		logEntry.Error(fmt.Sprintf("invalid http_status: %s for error: %s, returning 500", httpStatusString, e["name"]))
		httpStatus = 500
	}
	httpCode := getOrDefault(e, "http_code", "")
	httpMessage := getOrDefault(e, "http_message", "")
	return HTTPError{httpStatus, httpCode, httpMessage}
}

func getOrDefault(m map[string]string, key string, dflt string) string {
	value, ok := m[key]
	if ok {
		return value
	}
	return dflt
}
