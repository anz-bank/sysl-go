package jwtauth

import (
	"context"
	"encoding/json"
)

// Claims is weakly typed so it can hold any conceivable JSON claims value.
type Claims = map[string]interface{}

func clone(c Claims) Claims {
	// hack
	data, err := json.Marshal(&c)
	if err != nil {
		panic("cannot copy Claims")
	}
	var c2 Claims
	err = json.Unmarshal(data, &c2)
	if err != nil {
		panic("cannot copy Claims")
	}
	return c2
}

type claimsKeyStruct struct{}

var claimsKey = &claimsKeyStruct{}

// AddClaimsToContext adds claims to the context.
func AddClaimsToContext(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, clone(c))
}

// GetClaimsFromContext retrieves claims from the context.
//
// Returned claims is a safe copy of the context claims, so the context
// cannot be modified. To add new claims, you must re-add them to the context
// with AddClaimsToContext, and get a new context with the new claims added.
func GetClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}, false
	}
	return clone(claims), true
}
