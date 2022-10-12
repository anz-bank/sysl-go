SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway
SYSLGO_APP.gateway = Gateway
PKGPATH = grpc_jwt_authorization

include ../common.mk

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# internal/gen/pkg/servers/gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/gateway/main.go cmd/gateway/main_test.go internal/gen/pkg/servers/gateway
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk
