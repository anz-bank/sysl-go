package common

import (
	"context"
	"fmt"
	"net/http"
)

type Kind int

const (
	UnknownError Kind = iota
	BadRequestError
	InternalError
	UnauthorizedError
	DownstreamUnavailableError
	DownstreamTimeoutError
	DownstreamUnauthorizedError       // 401 from downstream
	DownstreamUnexpectedResponseError // unexpected response from downstream
	DownstreamResponseError           // application-leve error response from downstream
)

const downstreamResponseSnippetMaxLength = 128

func (k Kind) String() string {
	switch k {
	case BadRequestError:
		return "Missing one or more of the required parameters"
	case InternalError:
		return "Internal Server Error"
	case UnauthorizedError:
		return "Unauthorized error"
	case DownstreamUnavailableError:
		return "Downstream system is unavailable"
	case DownstreamTimeoutError:
		return "Time out from down stream services"
	case DownstreamUnauthorizedError:
		return "Unauthorized error from downstream services"
	case DownstreamUnexpectedResponseError:
		return "Unexpected response from downstream services"
	case DownstreamResponseError:
		return "Error response from downstream services"
	default:
		return "Internal Server Error"
	}
}

type ErrorKinder interface {
	ErrorKind() Kind
}

// If the error returned is an ErrorWriter the error handling will call the writeError method before any of the
// regular error handling (no mapping).
//
// If the call returns true it means it wrote the error and will not do any more handling.
// If it returns false it will go through the normal error writing path (via both the MapError and WriteError callbacks).
type ErrorWriter interface {
	WriteError(ctx context.Context, w http.ResponseWriter) bool
}

type ServerError struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("ServerError(Kind=%s, Message=%s, Cause=%s)", e.Kind, e.Message, e.Cause)
}

func (e *ServerError) ErrorKind() Kind {
	return e.Kind
}

func (e *ServerError) Unwrap() error {
	return e.Cause
}

func CreateError(ctx context.Context, kind Kind, message string, cause error) error {
	// we may push the error to NR here
	if err := CheckContextTimeout(ctx, message, cause); err != nil {
		return err
	}

	switch cause.(type) {
	case ErrorKinder, CustomError, wrappedError:
		return cause
	default:
		return &ServerError{Kind: kind, Message: message, Cause: cause}
	}
}

func CheckContextTimeout(ctx context.Context, message string, cause error) error {
	if ctx.Err() == context.DeadlineExceeded {
		return &ServerError{Kind: DownstreamTimeoutError, Message: message, Cause: cause}
	}
	return nil
}

type DownstreamError struct {
	Kind     Kind
	Response *http.Response
	Body     []byte
	Cause    error
}

func (e *DownstreamError) ErrorKind() Kind {
	return e.Kind
}

func (e *DownstreamError) Error() string {
	return fmt.Sprintf("DownstreamError(Kind=%s, Method=%s, URL=%s, StatusCode=%d, ContentType=%s, ContentLength=%d, Snippet=%s, Cause=%s)",
		e.Kind.String(),
		e.Response.Request.Method,
		e.Response.Request.URL.String(),
		e.Response.StatusCode,
		e.Response.Header.Get("Content-Type"),
		e.Response.ContentLength,
		string(e.Body),
		e.Cause)
}

func (e *DownstreamError) Unwrap() error {
	return e.Cause
}

func CreateDownstreamError(ctx context.Context, kind Kind, response *http.Response, body []byte, cause error) error {
	// we may push the error to NR here

	// add the request method and url as message, make the troubleshooting easier
	if err := CheckContextTimeout(ctx, fmt.Sprintf("%s %s", response.Request.Method, response.Request.URL.String()), cause); err != nil {
		return err
	}

	err := &DownstreamError{
		Kind:     kind,
		Response: response,
		Cause:    cause,
	}

	bodyLength := len(body)
	switch {
	case bodyLength == 0:
	case bodyLength > downstreamResponseSnippetMaxLength:
		err.Body = body[:downstreamResponseSnippetMaxLength]
	default:
		err.Body = body
	}

	return err
}

type ZeroHeaderLengthError struct {
	paramCanonical string
}

func NewZeroHeaderLengthError(param string) error {
	return &ZeroHeaderLengthError{paramCanonical: http.CanonicalHeaderKey(param)}
}

func (e *ZeroHeaderLengthError) Error() string {
	return fmt.Sprintf("%s header length is zero", e.paramCanonical)
}

func (e *ZeroHeaderLengthError) CausedByParam(param string) bool {
	return e.paramCanonical == http.CanonicalHeaderKey(param)
}

type InvalidHeaderError struct {
	paramCanonical string
	cause          error
}

func NewInvalidHeaderError(param string, cause error) error {
	return &InvalidHeaderError{paramCanonical: http.CanonicalHeaderKey(param), cause: cause}
}

func (e *InvalidHeaderError) Error() string {
	return fmt.Sprintf("%s header is invalid according to the spec", e.paramCanonical)
}

func (e *InvalidHeaderError) CausedByParam(param string) bool {
	return e.paramCanonical == http.CanonicalHeaderKey(param)
}

// This may be nil (for example if the cause was a regular expression mismatch) or may be InvalidValidationError for
// bad values passed in and nil or ValidationErrors as error otherwise.
// You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors
// (there may be more than one).
func (e *InvalidHeaderError) GetCause() error {
	return e.cause
}
