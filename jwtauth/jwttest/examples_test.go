package jwttest_test

import (
	"context"
	"fmt"

	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/jwtauth/jwttest"
)

// Shows how to use the issuer to issue a token with the desired claims.
func ExampleIssuer_Issue() {
	ctx := context.Background()
	issuer, _ := jwttest.NewIssuer("test", 1024)
	token, _ := issuer.Issue(jwtauth.Claims{
		"sub":   "me",
		"aud":   []string{"target"},
		"scope": "MY.SCOPE ANOTHER.SCOPE",
	})

	claims, _ := issuer.Authenticate(ctx, token)
	fmt.Println("iss:", claims["iss"])
	fmt.Println("sub:", claims["sub"])
	fmt.Println("aud:", claims["aud"])
	fmt.Println("scope:", claims["scope"])

	// Output: iss: test
	// sub: me
	// aud: [target]
	// scope: MY.SCOPE ANOTHER.SCOPE
}
