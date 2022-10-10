SYSLGO_SYSL=specs/specs.sysl
SYSLGO_PACKAGES=rest grpc
SYSLGO_APP.rest = REST_Subdir
SYSLGO_SUBDIR.rest = rest
SYSLGO_APP.grpc = GRPC_Subdir
SYSLGO_SUBDIR.grpc = grpc

include ../common.mk
include codegen.mk

subdir/proto.pb.go: specs/proto.proto
	(cd specs && protoc --go_out=../subdir --go-grpc_out=../subdir --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative proto.proto)

test: gen-all-servers subdir/proto.pb.go
	go run ./rest/cmd/rest/main.go --help 
	go run ./grpc/cmd/grpc/main.go --help
