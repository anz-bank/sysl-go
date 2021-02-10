purpose:

Demonstrate support for making HTTP POST request
where the request body is URL-encoded form data.

Tests a scenario where we specify:

1. an trivial external "banana stand" backend service
-   specified by an open API 3 spec
-   defining a POST endpoint that demands url encoded form data in the request body

2. a sysl-go generated REST service named "gateway" that is able to make a POST request with a URL-encoded body to the "banana stand" service.
