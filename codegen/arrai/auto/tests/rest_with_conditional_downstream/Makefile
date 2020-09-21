SYSLGO_SYSL=specs/gateway.sysl
SYSLGO_PACKAGES=Gateway Backend
PKGPATH=rest_with_conditional_downstream
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../


# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# gen/pkg/servers/Gateway. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	main_test.go main.go gen/pkg/servers/Gateway
	go test -v ./...
PHONY: .test

include codegen.mk

clean:
	rm -rf gen/pkg/servers/Gateway
.PHONY: clean
