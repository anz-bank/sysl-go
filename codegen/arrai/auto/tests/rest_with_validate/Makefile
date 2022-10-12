SYSLGO_SYSL = specs/pingpong.sysl
SYSLGO_PACKAGES = pingpong pingpongwithvalidate
SYSLGO_APP.pingpong = PingPong
SYSLGO_APP.pingpongwithvalidate = PingPongWithValidate
PKGPATH = rest_with_validate

include ../common.mk

test:	gen-all-servers
	go test $(GO_TEST_FLAGS) ./...
PHONY: .test

include codegen.mk
