purpose:

Test some support for REST error handling.

This includes the pathological scenario where a downstream wants to return an error response JSON object
with a property named "Error" that collides with the go "func Error() string" method of go's error interface.