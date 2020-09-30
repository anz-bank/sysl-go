package authrules

import (
	"context"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/authexpr"
	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/jwtauth/jwtgrpc"
)

// ClaimsBasedAuthorisationRule decides if access is approved or denied based on the given claims.
// Returning true, nil indicates access is approved.
// Returning false, nil indicates access is denied.
// Returning *, err endicates an error occurred when evaluating the rule.
type JWTClaimsBasedAuthorisationRule func(ctx context.Context, claims jwtauth.Claims) (bool, error)

func MakeDefaultJWTClaimsBasedAuthorisationRule(authorisationRuleExpression string) (JWTClaimsBasedAuthorisationRule, error) {
	// compile the rule expression early so we can detect misconfiguration and fail early.
	rootExpr, err := authexpr.CompileExpression(authorisationRuleExpression)
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

// MakeGRPCAuthorisationRule creates an authorisation Rule from a claims-based authorisation Rule
// and a jwtauth Authenticator.
func MakeGRPCJWTAuthorisationRule(authRule JWTClaimsBasedAuthorisationRule, authenticator jwtauth.Authenticator) (Rule, error) {
	return func(ctx context.Context) (context.Context, error) {
		rawToken, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		if err != nil {
			log.Debugf(ctx, "auth: error extracting jwt from context: %v", err)
			return nil, err
		}
		claims, err := authenticator.Authenticate(ctx, rawToken)
		if err != nil {
			log.Debugf(ctx, "auth: jwt authentication failed, access denied: %v", err)
			return nil, err
		}
		isAuthorised, err := authRule(ctx, claims)
		if err != nil {
			log.Debugf(ctx, "auth: error evaluating authorisation rule: %v", err)
			return nil, err
		}
		if !isAuthorised {
			log.Debugf(ctx, "auth: request is not authorised, access denied")
			return nil, jwtgrpc.ErrClaimsValidationFailed
		}

		log.Debugf(ctx, "auth: request authenticated and authorized successfully")
		ctx = jwtauth.AddClaimsToContext(ctx, claims)
		return ctx, nil
	}, nil
}

func InsecureAlwaysGrantAccess(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
