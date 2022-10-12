SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway
SYSLGO_APP.gateway = Gateway
PKGPATH = rest_post_urlencoded_form

include ../common.mk

test: gen-all-servers
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk
