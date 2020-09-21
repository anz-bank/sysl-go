purpose:

test of codegen for arrai/auto approach, focusing on REST

Tests scenario where we specify:
* a trivial downstream backend service with three different REST endpoints
* a second "gateway" backend service, also with a single REST endpoint, that sits between the client and the downstream service
* the gateway calls different endpoints on the backend service
* the endpoint calls to the backend service are expressed using sysl's conditional logic

This test assumes that the codegen has no support for interpreting the conditional logic expressed in the sysl file. We can
leave it up to the application programmers to fill that in. The code generator does need to be able to discover the calls
to the different endpoints and provide the application programmer with a Client that is able to call anything that might
need to be called.
