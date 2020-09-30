package authexpr

import (
	"strings"
)

func MakeStandardJWTHasScope(claims map[string]interface{}) func(scope string) (bool, error) {
	return func(queryScope string) (bool, error) {
		// Ref: https://tools.ietf.org/html/rfc8693#section-4.2
		if scope, ok := claims["scope"]; ok {
			if scopeString, ok := scope.(string); ok {
				// The syntax of the value stored under the "scope" key
				// is defined in rfc6749#section-3.3 as:
				//
				// scope       = scope-token *( SP scope-token )
				// scope-token = 1*( %x21 / %x23-5B / %x5D-7E )
				//
				// i.e. the string contains 1 or more space-separated
				// scope-token values, where each scope-token is a
				// word consisting of one or more bytes matching the
				// above ranges. For simplicity we don't bother
				// validating the scope-token bytes and accept anything.
				//
				// Ref: https://tools.ietf.org/html/rfc6749#section-3.3
				// Ref: https://tools.ietf.org/html/rfc5234
				scopes := strings.Split(scopeString, " ")
				for _, s := range scopes {
					if s == queryScope {
						return true, nil
					}
				}
			}
		}
		return false, nil
	}
}
