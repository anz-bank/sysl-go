SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway encoderbackend
SYSLGO_APP.gateway = Gateway
SYSLGO_APP.encoderbackend = EncoderBackend
PKGPATH = rest_with_raw_response

include ../common.mk

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# internal/gen/pkg/servers/gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/gateway/main_test.go cmd/gateway/main.go internal/gen/pkg/servers/gateway
	go test -v ./...
PHONY: .test

include codegen.mk
