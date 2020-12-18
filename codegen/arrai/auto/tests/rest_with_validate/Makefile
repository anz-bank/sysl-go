SYSLGO_SYSL = specs/pingpong.sysl
SYSLGO_PACKAGES = pingpong
SYSLGO_APP.pingpong = PingPong
PKGPATH = rest_with_validate

include ../common.mk

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# internal/gen/pkg/servers/pingpong. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/pingpong/main_test.go cmd/pingpong/main.go internal/gen/pkg/servers/pingpong
	go test -v ./...
PHONY: .test

include codegen.mk
