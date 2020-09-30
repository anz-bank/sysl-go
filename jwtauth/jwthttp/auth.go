package jwthttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/anz-bank/sysl-go/jwtauth"
)

// Config defines authentication config for an http middleware.
type Config struct {
	jwtauth.Config `mapstructure:",squash"`
	Headers        []string `json:"headers" yaml:"headers" mapstructure:"headers"`
}

// AuthFromConfig creates an auth middleware from config.
func AuthFromConfig(ctx context.Context, c *Config, client func(string) *http.Client) (*Auth, error) {
	authenticator, err := jwtauth.AuthFromConfig(ctx, &c.Config, client)
	if err != nil {
		return nil, err
	}
	return &Auth{
		Headers:       c.Headers,
		Authenticator: authenticator,
		UnauthHandler: DefaultUnauthHandler,
	}, nil
}

// Auth can authenticate and authorize requests.
type Auth struct {
	// Headers to search for a bearer token
	// Leaving empty will cause the authenticator to authenticate against an empty string
	// This is useful for use in local without jwts
	Headers []string

	// Authenticator is the authenticator used to verify jwts
	jwtauth.Authenticator

	// UnauthHandler handles the response in the case of a unauthenticated request
	// Leaving nil will default to a bare 401 response
	UnauthHandler func(http.ResponseWriter, *http.Request, error)

	// Authorisers store the authorisers that will get applied by this auth struct.
	Authorisers []jwtauth.Authoriser
}

// WithUnauthHandler creates a new auth object with the unauth handler set
//
// This allows defining custom response behaviour on a per-route basis.
func (a Auth) WithUnauthHandler(handler func(http.ResponseWriter, *http.Request, error)) *Auth {
	a.UnauthHandler = handler
	return &a
}

func (a Auth) WithAuthorisers(authorisers ...jwtauth.Authoriser) *Auth {
	a.Authorisers = append(a.Authorisers, authorisers...)
	return &a
}

// Authorise implements jwtauth.Authoriser
//
// Applies each stored outhoriser in order that it was added to the auth struct.
func (a Auth) Authorise(claims jwtauth.Claims) error {
	for _, authoriser := range a.Authorisers {
		if err := authoriser.Authorise(claims); err != nil {
			return err
		}
	}
	return nil
}

// Auth is a middleware function. It takes a handler and produces a new handler that authenticates and authorises
// requests before passing them to the given handler.
func (a Auth) Auth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if a.UnauthHandler == nil {
			a.UnauthHandler = DefaultUnauthHandler
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := jwtauth.GetClaimsFromContext(r.Context())
			if !ok {
				var err error
				claims, err = a.AuthenticateRequest(r)
				if err != nil {
					a.UnauthHandler(w, r, err)
					return
				}
				r = r.WithContext(jwtauth.AddClaimsToContext(r.Context(), claims))
			}
			if err := a.Authorise(claims); err != nil {
				a.UnauthHandler(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthAllowAnon is a middleware function. It takes a handler and produces a new handler that authenticates and
// authorises requests before passing them to the given handler.
//
// If an authorization header is present, the contained JWT is validated.  Otherwise, middleware processing continues.
// AllowAnon is useful where claims (if present) are required by a middleware stack, but not all endpoints in a mux require
// a jwt to be present.
//
// In this situation, authenticated endpoints each require an additional Auth middleware (which will reuse the authenticated
// claims).
func (a Auth) AuthAllowAnon() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if a.UnauthHandler == nil {
			a.UnauthHandler = DefaultUnauthHandler
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := jwtauth.GetClaimsFromContext(r.Context())
			if !ok {
				raw := a.getBearer(r.Header)
				if raw == "" {
					next.ServeHTTP(w, r)
					return
				}
				var err error
				claims, err := a.Authenticate(r.Context(), raw)
				if err != nil {
					a.UnauthHandler(w, r, err)
					return
				}
				r = r.WithContext(jwtauth.AddClaimsToContext(r.Context(), claims))
			}
			if err := a.Authorise(claims); err != nil {
				a.UnauthHandler(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthenticateRequest authenticates the request
//
// Returns the claims contained in the jwt, or an error if unable to
// authenticate.
func (a *Auth) AuthenticateRequest(req *http.Request) (jwtauth.Claims, error) {
	// Find the token in the request and authenticate it
	// It is the job of AuthN to accept or reject requests with no token
	raw := a.getBearer(req.Header)
	claims, err := a.Authenticate(req.Context(), raw)
	if err != nil {
		return jwtauth.Claims{}, err
	}
	return claims, nil
}

func (a *Auth) getBearer(headers http.Header) string {
	for _, header := range a.Headers {
		val := headers.Get(header)
		if len(val) > 8 && strings.ToLower(val[:7]) == "bearer " {
			return val[7:]
		}
	}
	return ""
}
