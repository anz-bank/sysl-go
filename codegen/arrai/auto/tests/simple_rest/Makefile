SYSLGO_SYSL=specs/pingpong.sysl
SYSLGO_PACKAGES=pingpong
SYSLGO_APP.pingpong = PingPong
SYMLINK=1
PKGPATH=simple_rest
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

# This rule is wonky as make does not understand there is
# a dependency between the specs and the *.go files *inside*
# gen/pkg/servers/PingPong. But, if we add those detailed rules,
# it is not compatible with how codegen.mk is structured.
test:	cmd/PingPong/main_test.go cmd/PingPong/main.go gen/pkg/servers/PingPong
	go test -v ./...
PHONY: .test

include codegen.mk

clean:
	rm -rf gen/pkg/servers/PingPong/
	rm -rf pingpong.testresults
.PHONY: clean
