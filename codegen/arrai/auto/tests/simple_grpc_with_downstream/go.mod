module simple_grpc_with_downstream

go 1.14

replace github.com/anz-bank/sysl-go => ../../../../..

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	github.com/anz-bank/pkg v0.0.33
	github.com/anz-bank/sysl-go v0.0.0-20200325045908-46c4ce0a2736
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/sethvargo/go-retry v0.1.0
	github.com/spf13/afero v1.4.0
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)
