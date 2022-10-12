SYSLGO_SYSL = specs/app.sysl
SYSLGO_PACKAGES = app
SYSLGO_APP.app = App
PKGPATH = rest_admin_server

include ../common.mk

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# internal/gen/pkg/servers/app. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/app/main_test.go cmd/app/main.go internal/gen/pkg/servers/app
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk
