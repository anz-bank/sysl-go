SYSLGO_SYSL = specs/gateway.sysl
SYSLGO_PACKAGES = gateway gatewayWithBff
SYSLGO_APP.gateway = Gateway
SYSLGO_APP.gatewayWithBff = GatewayWithBff
PKGPATH = rest_miscellaneous

include ../common.mk

test: gen-all-servers
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk

# copy external_types to gen folder
$(SERVERS_ROOT)/gateway/external_types.go: internal/external_types/gateway/external_types.go
	cp -f $< $@

$(SERVERS_ROOT)/gatewayWithBff/gateway/external_types.go: internal/external_types/gateway/external_types.go
	cp -f $< $@

gen-all-servers: $(SERVERS_ROOT)/gateway/external_types.go $(SERVERS_ROOT)/gatewayWithBff/gateway/external_types.go
