package jwtauth

import (
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthError(t *testing.T) {
	err := &AuthError{
		Code:  1,
		Cause: errors.New("cause"),
	}
	errString := err.Error()
	assert.Contains(t, errString, strconv.Itoa(err.Code))
	assert.Contains(t, errString, err.Cause.Error())
}

func TestAuthErrorHTTPStatus(t *testing.T) {
	err := &AuthError{
		Code:  AuthErrCodeBadSignature,
		Cause: errors.New("cause"),
	}
	assert.Equal(t, http.StatusForbidden, err.HTTPStatus())
}

func TestAuthErrorHTTPStatusUnknownCode(t *testing.T) {
	err := &AuthError{
		Code:  10000,
		Cause: errors.New("cause"),
	}
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus())
}

func TestAuthErrorUnwrap(t *testing.T) {
	inner := errors.New("Error")
	err := &AuthError{
		Cause: inner,
	}
	assert.Equal(t, inner, err.Unwrap())
}
