/*
Package jwtauth manages request authentication with jwts

This module provides a simple API to manage authentication and authorization
with JWTs (JSON Web Token). Use this module when you have any one of the
following needs

1. You need to authenticate JWTs from various token issues

2. You need to authorize JWTs have sufficient claims to execute a request

3. You need fine-grained control over the endpoints that get auth'n/o, and what
gets checked

Authentication

Authentication in the JWT world refers to the action of verifying a jwt comes
from a trusted source. This occurs via the Authenticator interface...

    type Authenticator interface {
		Authenticate(string token) (*Claims, error)
	}

TODO(cantosd): document configuration when done

Authorization

Authorization in the JWT world refers to verifying the claims have sufficient
permissions to execute the request. This always happens after authentication,
and it is generally assumed that jwts coming from trusted sources have had the
claims verified by the source (issuer will not generate a jwt with admin
permissions for a customer).

Authorization is handled by the Authorizor interface...

    type Authorizor interface {
		Authorize(claims *Claims) error
	}

It is up to the application to implement this interface. To do this, first
define what valid permissions are for any given request, then implement this
interface and make sure the request passes through it.

TODO(cantosd): Add middleware when done
*/
package jwtauth
