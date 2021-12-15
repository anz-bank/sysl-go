module simple_grpc_with_downstream

go 1.16

replace github.com/anz-bank/sysl-go => ../../../../..

require (
	github.com/anz-bank/sysl-go v0.189.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
)
