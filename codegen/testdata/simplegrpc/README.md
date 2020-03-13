A simple gRPC server specification in SySL and Protobuf format. Usually the `.proto` would be generated from the `.sysl` but at the time of writing that transformation was not available.

The `.proto` is used to generate the necessary Go code (via `protoc`) to bind the gRPC protocol to Go. The `.sysl` file is used to generate the necessary Go code that binds the `protoc`-generated Go code to sysl-go-comms.

How this is done can be seen in the `Module.mk`.

Go tests can be found in `../tests/simplegrpc`, the target directory for the code generation steps.
