SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway
SYSLGO_APP.gateway = Gateway
PKGPATH = rest_with_conditional_downstream

include ../common.mk

test: gen-all-servers
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk
