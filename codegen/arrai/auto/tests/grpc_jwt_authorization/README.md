smoketest `grpc_jwt_authorization`
=================================

`grpc_jwt_authorization` is a smoke test for the
auto generated code to test support for JWT-based
authentication and authorization to control access
to generated gRPC methods.

JWT authentication and authorization dance:

```
    C : client
    B : token issuer
    R : resource server


    C --- i am C: my creds ->  B   OK, your creds look good, here's a JWT we have signed
    C <--- fresh JWT --------      certifying that you are C and you can "hello".

    .
    .
    .

    C ---- call hello + JWT -> R   OK, you have a JWT that claims you have permission to "hello"
                                   OK, the JWT says it was signed by B.  We trust B.
                                   OK, you are authorisated. "hello".
    C <--- hello response ------
```

In this smoke test we focus on code-generating `R`, the
resource server.

During the smoke test we:

* start with a sysl specification of the resource server that
  contains an `@authorization_rule` annotation describing
  a JWT-based authorization rule for one of its endpoints
* use sysl-go auto to generate the code for this resource server
* stand up a test token issuer and run the issuer as a http service
* configure our resource server to trust the test token issuer
* simulate a client (using the test harness) that obtains a token
  from the test issuer and then attempts to make a request to the
  resource server using the token
* try a few different variations with invalid tokens, corrupted
  tokens etc to demonstrate that the resource server detects this
  and denies access.

