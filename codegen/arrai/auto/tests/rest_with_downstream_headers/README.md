purpose:

test of codegen for arrai/auto approach, focusing on REST & HTTP header support

Tests a scenario where we specify:
* a trivial downstream backend service with a single REST endpoint
* a second "gateway" backend service, also with a single REST endpoint, that sits between the client and the downstream service
* some optional HTTP headers
* some required HTTP headers
