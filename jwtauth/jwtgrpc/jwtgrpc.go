package jwtgrpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth errors that are safe to return over the wire.
var (
	// ErrNoAuthHeader is returned when the request did not contain an auth header.
	ErrNoAuthHeader = status.Errorf(codes.Unauthenticated, "no authorization header")

	// ErrParseJWT is returned when the header or JWT failed to parse correctly.
	ErrParseJWT = status.Errorf(codes.Unauthenticated, "invalid authorization header")

	// ErrAuthenticationFailed is returned when the jwt cannot be authenticated.
	ErrAuthenticationFailed = status.Errorf(codes.Unauthenticated, "authentication failed")

	// ErrClaimsValidationFailed is returned when the jwts claims are insufficient for the target method.
	ErrClaimsValidationFailed = status.Errorf(codes.PermissionDenied, "insufficient permissions")
)

// GetBearerFromIncomingContext extracts a Bearer auth header from incoming metadata.
func GetBearerFromIncomingContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrNoAuthHeader
	}
	authHdr := md.Get("Authorization")
	if len(authHdr) == 0 {
		return "", ErrNoAuthHeader
	}
	authVal := authHdr[0]
	// we don't use strings.Prefix methods because it makes it harder to ignore case
	if len(authVal) < 7 || strings.ToLower(authVal[:7]) != "bearer " {
		return "", ErrParseJWT
	}
	return authVal[7:], nil
}
