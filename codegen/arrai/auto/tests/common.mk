SYMLINK=1
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

PB_GO_TARGETS = $(foreach proto,$(PROTOS),internal/gen/pb/$(proto)/$(proto).pb.go)

PROTOC_GRPC_PB_GO = mkdir -p $(dir $@) && protoc --proto_path=specs --go_out=$(dir $@) --go-grpc_out=$(dir $@) $^

default: test
.PHONY: default

protos: $(PB_GO_TARGETS)
.PHONY: protos

clean:
	rm -rf \
		.gitattributes \
		.github \
		Dockerfile \
		internal/gen/pkg \
		$(CLEAN)
.PHONY: clean
