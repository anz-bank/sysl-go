SYSLGO_SYSL=specs/specs.sysl
SYSLGO_PACKAGES=temporal_worker temporal_client downstream
SYSLGO_APP.temporal_client = Frontdoor
SYSLGO_APP.temporal_worker = TemporalWorker
SYSLGO_APP.downstream = SomeDownstream

include ../common.mk
include codegen.mk

test: gen-all-servers
	go test ./...
