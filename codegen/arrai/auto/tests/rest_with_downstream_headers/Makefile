SYSLGO_SYSL=specs/gateway.sysl
SYSLGO_PACKAGES=gateway encoderbackend
SYSLGO_APP.gateway = Gateway
SYSLGO_APP.encoderbackend = EncoderBackend
SYMLINK=1
PKGPATH=rest_with_downstream_headers
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../


# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# gen/pkg/servers/Gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/Gateway/main_test.go cmd/Gateway/main.go gen/pkg/servers/Gateway
	go test -v ./...
PHONY: .test

include codegen.mk

clean:
	rm -rf gen/pkg/servers/Gateway
.PHONY: clean
