package authrules

import (
	"context"
	"fmt"
	"strings"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/authexpr"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/jwtauth/jwtgrpc"
)

// ClaimsBasedAuthorizationRule decides if access is approved or denied based on the given claims.
// Returning true, nil indicates access is approved.
// Returning false, nil indicates access is denied.
// Returning *, err endicates an error occurred when evaluating the rule.
type JWTClaimsBasedAuthorizationRule func(ctx context.Context, claims jwtauth.Claims) (bool, error)

func MakeDefaultJWTClaimsBasedAuthorizationRule(authorizationRuleExpression string) (JWTClaimsBasedAuthorizationRule, error) {
	// compile the rule expression early so we can detect misconfiguration and fail early.
	rootExpr, err := authexpr.CompileExpression(authorizationRuleExpression)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, claims jwtauth.Claims) (bool, error) {
		evalCtx := authexpr.EvaluationContext{
			JWTHasScope: authexpr.MakeStandardJWTHasScope(claims),
		}
		return rootExpr.Evaluate(evalCtx)
	}, nil
}

// MakeGRPCAuthorizationRule creates an authorization Rule from a claims-based authorization Rule
// and a jwtauth Authenticator.
func MakeGRPCJWTAuthorizationRule(authRule JWTClaimsBasedAuthorizationRule, authenticator jwtauth.Authenticator) (Rule, error) {
	return func(ctx context.Context) (context.Context, error) {
		rawToken, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		if err != nil {
			log.Debugf(ctx, "auth: error extracting jwt from context: %v", err)
			return nil, err
		}
		return authorize(ctx, rawToken, authRule, authenticator)
	}, nil
}

// MakeRESTJWTAuthorizationRule creates an authorization Rule from a claims-based authorization Rule
// and a jwtauth Authenticator.
func MakeRESTJWTAuthorizationRule(authRule JWTClaimsBasedAuthorizationRule, authenticator jwtauth.Authenticator) (Rule, error) {
	return func(ctx context.Context) (context.Context, error) {
		rawToken, err := getBearerTokenFromIncomingRESTContext(ctx)
		if err != nil {
			log.Debugf(ctx, "auth: error extracting jwt from context: %v", err)
			return nil, err
		}
		return authorize(ctx, rawToken, authRule, authenticator)
	}, nil
}

func authorize(ctx context.Context, rawToken string, authRule JWTClaimsBasedAuthorizationRule, authenticator jwtauth.Authenticator) (context.Context, error) {
	claims, err := authenticator.Authenticate(ctx, rawToken)
	if err != nil {
		log.Debugf(ctx, "auth: jwt authentication failed, access denied: %v", err)
		return ctx, err
	}
	isAuthorised, err := authRule(ctx, claims)
	if err != nil {
		log.Debugf(ctx, "auth: error evaluating authorization rule: %v", err)
		return ctx, err
	}
	if !isAuthorised {
		log.Debugf(ctx, "auth: request is not authorised, access denied")
		return ctx, jwtgrpc.ErrClaimsValidationFailed
	}

	log.Debugf(ctx, "auth: request authenticated and authorized successfully")
	ctx = jwtauth.AddClaimsToContext(ctx, claims)
	return ctx, nil
}

func getBearerTokenFromIncomingRESTContext(ctx context.Context) (string, error) {
	header := common.RequestHeaderFromContext(ctx)
	val := header.Get("Authorization")
	if len(val) > 8 && strings.ToLower(val[:7]) == "bearer " {
		return val[7:], nil
	}
	return "", &jwtauth.AuthError{
		Code:  jwtauth.AuthErrCodeInvalidJWT,
		Cause: fmt.Errorf("no Authorization header containing bearer token"),
	}
}

func InsecureAlwaysGrantAccess(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
