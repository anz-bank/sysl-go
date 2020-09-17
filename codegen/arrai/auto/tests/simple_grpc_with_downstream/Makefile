SYSLGO_SYSL=specs/gateway.sysl
SYSLGO_PACKAGES=Gateway EncoderBackend
PKGPATH=simple_grpc_with_downstream
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../


# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# gen/pkg/servers/Gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	gen/pkg/servers/Gateway gen/pb/encoder_backend/encoder_backend.pb.go gen/pb/gateway/gateway.pb.go
	go test -v ./...
PHONY: .test

gen/pb/encoder_backend/encoder_backend.pb.go: specs/encoder_backend.proto
	protoc --proto_path=specs --go_out=plugins=grpc:gen/pb/encoder_backend $^

gen/pb/gateway/gateway.pb.go: specs/gateway.proto
	protoc --proto_path=specs --go_out=plugins=grpc:gen/pb/gateway $^

include codegen.mk

clean:
	rm -rf gen/pkg/servers/Gateway gen/pb/gateway/gateway.pb.go gen/pb/encoder_backend/encoder_backend.pb.go
.PHONY: clean
