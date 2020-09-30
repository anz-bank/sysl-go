package jwtauth

import (
	"fmt"
	"net/http"
)

// AuthError is a protocol agnostic error.
//
// Codes can be translated into actual protocol error codes.
type AuthError struct {
	Code  int
	Cause error
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("jwtauth err %d: %v", e.Code, e.Cause)
}

// Unwrap implements the unwrap interface in errors.
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns an http status code corresponding to the AuthError code.
func (e *AuthError) HTTPStatus() int {
	code, ok := errHTTPCodeMap[e.Code]
	if !ok {
		return http.StatusInternalServerError
	}
	return code
}

// Authorization error codes.
const (
	AuthErrCodeUnknown = iota
	AuthErrCodeInvalidJWT
	AuthErrCodeUntrustedSource
	AuthErrCodeBadSignature
	AuthErrCodeInsufficientPermissions
)

var errHTTPCodeMap = map[int]int{
	AuthErrCodeUnknown: http.StatusInternalServerError,

	// Request has no jwt or jwt is invalid.
	AuthErrCodeInvalidJWT: http.StatusUnauthorized,

	// Request has credentials but we don't trust where request or got them from.
	AuthErrCodeUntrustedSource: http.StatusForbidden,

	// Request credentials cannot be verified, either public key is bad or jwt is not from the source it claims to be.
	AuthErrCodeBadSignature: http.StatusForbidden,

	// Request is authenticated but does not have sufficient permissions to execute.
	AuthErrCodeInsufficientPermissions: http.StatusForbidden,
}
