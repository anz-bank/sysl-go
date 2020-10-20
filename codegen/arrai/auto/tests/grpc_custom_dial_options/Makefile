SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway encoderbackend
SYSLGO_APP.gateway = Gateway
SYSLGO_APP.encoderbackend = EncoderBackend
PKGPATH = grpc_custom_dial_options

PROTOS = encoder_backend gateway

include ../common.mk

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# internal/gen/pkg/servers/gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test: cmd/gateway/main.go cmd/gateway/main_test.go internal/gen/pkg/servers/gateway
	go test -v ./...
PHONY: .test

# n.b. commented out these deps as the CI build doesnt have protoc installed yet
# instead the generated *.pb.go files are checked in to version control.
# test:	internal/gen/pb/encoder_backend/encoder_backend.pb.go internal/gen/pb/gateway/gateway.pb.go

internal/gen/pb/encoder_backend/encoder_backend.pb.go: specs/encoder_backend.proto
	$(PROTOC_GRPC_PB_GO)

internal/gen/pb/gateway/gateway.pb.go: specs/gateway.proto
	$(PROTOC_GRPC_PB_GO)

include codegen.mk
