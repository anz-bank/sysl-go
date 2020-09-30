BEWARE

This is sysl-go's vendored version of the jwtauth library.
Some of the below capabilities have been excised.


# JWTAuth2

This document details the architecture of the jwtauth2 module. This module is a
replacement for middleware/jwtauth.

## Summary

JSON web tokens are a widely used auth token specification and are part of the
larger JSON object signing and encryption (JOSE) standards.

The basic goal of jwt middleware is to verify a request is authentic by
verifying the request jwt comes from a trusted source. Once it does this, it
extracts the claims inside the jwt, and passes them downstream in the request
context.

The goal of this module is to make it easy to create and configure said
middleware, and provide auxiliary functions to make it easy to interact with
the claims inside the jwt if an application needs to do further validation.

## Scope

The scope of this module is as follows

1. jwtauth package provides foundational auth functionality
    * Authenticator interface responsible for authenticating jwts and
    extracting claims
    * Configurable authenticator struct that can authenticate against multiple
    issuers
    * Authorizor interface responsible for authorizing against claims
    * RESTRICTED LOGGING (only logs background processes, eg: refresh remote
    jwks cache)
2. middleware packages provide middleware
    * One package per protocol (currently intended to include http and grpc)
    * Auth struct responsible for constructing middleware.
    * Configuration for Auth struct
    * Middleware responsible for
        * Authentication
        * Authorization
        * Handling auth errors
        * Logging

## Architecture (Design)

This section documents the design of the API. For the API itself, you can
generate the godoc using the `godoc` tool

```bash
$ godoc -index -http=localhost:6000
```

And go to `localhost:6000` in your browser and search for `jwtauth/v2`

### Base auth package

This package provides an API for configuring an applications authenticator and
defining custom authorizors. As such, the bas package API consists of two
components

#### Authentication

The role of authentication is to verify that the requestor is indeed who they
claim to be. In the jwt world, this is done by verifying the signature of the
jwt. To do this, we need to use a public key corresponding to the private key
that signed the jwt. There are multiple factors that influence how we obtain
the public key and whether we accept a jwt.

1. JWTs can come from any one of multiple trusted sources. To identify the
source, jwts have an issuer (iss) claim.
2. Issuers can sign jwts with any one of multiple keys. To identify the key,
jwts have a key id (kid) in the jwt header
3. JWTs can be signed using one of many signing algorithms. To identify the
algorithm, jwts have an algorithm (alg) in the jwt header. This is usually
handled by jwt parsers automatically, but...
4. Some signing algorithms are inherently insecure, we need to be able to
configure which algorithms we accept.

All of this complexity relates to executing a single function from an
applications perspective, so we hide it behind an authenticator interface.

```go
type Authenticator interface {
    Authenticate(string) (*Claims, error)
}
```

The standard implementation of this interface allows registering of multiple
jwt issuers

```go
type StdAuthenticator struct {
    verifiers map[string]Verifier
}
```

The map consists of key-value pairs where the keys correspond to named issuers,
while the values are verifiers that can verify a jwt from that issuer.

Verifiers are themselves interfaced to reflect different methods of obtaining
public keys.

```go
type Verifier interface {
    Verify(token *jwt.JSONWebToken, claims interface{}) error
}
```

The job of this interface is to verify and extract the jwt claims. It is
important for the claims to be interfaced, as this allows us to change
the claims format later if standards change, without changing the api too much.

##### Verifier implementations

TODO(cantosd): Document when remote jwks cache is implemented

##### Configuration

TODO(cantosd): Document when configuration is done

#### Authorization

The role of authorization is to authorize a request, that is, verify the
request has sufficient permissions to be executed. In JWT world, the jwt has
a set of claims. What these claims can look like is unclear, as this part of
the jwt standard is extensible.

However, any claims value must be possible to encode as a JSON object,
therefore in go code we can represent it as

```go
map[string]interface{}
```

Authorization consists of checking whether those claims are enough to execute
the request. Defining what *sufficient* means is up to the application, so this
module provides an interface to do so

```go
type Authorizor interface {
    Authorize(claims *Claims) error
}
```

We return an error instead of a bool, as an error can contain more information,
such as a cause that can be logged, and an error code for the response.

#### Errors

This package defines a set of error codes that encapsulate the different errors
that can occur during authentication and authorization with jwts. The error
struct exposes status methods that can convert an internal error code to a
protocol error code

```go
type AuthError struct {
    Code: int
    Cause: error
}
func (a *AuthError) HTTPStatus() int {
    // translate internal code to http status
}
func (a *AuthError) GRPCCode() int {
    // translate code to grpc status code
}
```

Consumers can decide whether they send the cause with the response or not.

### Middleware

TODO(cantosd): document when implemented
