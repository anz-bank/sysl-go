SYSLGO_SYSL=specs/specs.sysl
SYSLGO_PACKAGES=temporal_client
SYSLGO_APP.temporal_client = Frontdoor

include ../common.mk
include codegen.mk

test: gen-all-servers
	go run ./cmd/temporal_client/ --help
