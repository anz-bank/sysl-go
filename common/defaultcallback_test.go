package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapError_NilMapErrorFunc_DefaultMapping(t *testing.T) {
	// Arrange
	cb := Callback{MapErrorFunc: nil}

	// Act
	httpErr := cb.MapError(context.Background(), fmt.Errorf("fmt.Errorf go brrr"))

	// Assert
	require.Equal(t, 500, httpErr.HTTPCode)
}

func TestMapError_NilMapErrorFuncWithDownstreamError_DefaultMapping(t *testing.T) {
	// Arrange
	cb := Callback{MapErrorFunc: nil}

	// Act
	httpErr := cb.MapError(context.Background(), &DownstreamError{Kind: DownstreamUnavailableError})

	// Assert
	require.Equal(t, 503, httpErr.HTTPCode)
}

func TestMapError_CustomMapErrorFunc_IsCalled(t *testing.T) {
	// Arrange
	wasCalled := false
	cb := Callback{MapErrorFunc: func(ctx context.Context, err error) *HTTPError {
		wasCalled = true
		return &HTTPError{HTTPCode: 418, Description: "I'm a teapot"}
	}}

	// Act
	httpErr := cb.MapError(context.Background(), fmt.Errorf("need coffee"))

	// Assert
	require.Equal(t, 418, httpErr.HTTPCode)
	require.True(t, wasCalled)
}
