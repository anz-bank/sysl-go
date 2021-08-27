SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway gatewayWithBff
SYSLGO_APP.gateway = Gateway
SYSLGO_APP.gatewayWithBff = GatewayWithBff
PKGPATH = rest_miscellaneous

include ../common.mk

test:gen-all-servers
	go test -v ./...
PHONY: .test

include codegen.mk
