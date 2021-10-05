module grpc_jwt_authorization

go 1.16

replace github.com/anz-bank/sysl-go => ../../../../..

require (
	github.com/anz-bank/sysl-go v0.189.0
	github.com/sethvargo/go-retry v0.1.0
	github.com/spf13/afero v1.6.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
