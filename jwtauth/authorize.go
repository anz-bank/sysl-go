package jwtauth

// Authoriser is an interface that can authorize claims.
type Authoriser interface {
	Authorise(Claims) error
}

// AuthoriseFunc is a function type that implements Authorizor.
type AuthoriseFunc func(Claims) error

// Authorise implements Authorizor for the AuthorizeFunc type.
func (a AuthoriseFunc) Authorise(c Claims) error {
	return a(c)
}
