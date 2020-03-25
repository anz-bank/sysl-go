# Simple Server
SIMPLESG=simplesg
SIMPLEGRPC_IN=codegen/testdata/simplegrpc
SIMPLEGRPC_OUT=codegen/tests/simplegrpc

SIMPLEGRPC_SYSL_FILES=$(addprefix $(SIMPLEGRPC_OUT)/$(SIMPLESG)/, $(GRPC_SERVER_FILES))
SIMPLEGRPC_PROTO_FILES=$(SIMPLEGRPC_OUT)/simplepb/simplegrpc.pb.go

.PHONY: simplegrpc-gen
simplegrpc-gen: APP=SimpleGrpc
simplegrpc-gen: MODEL=$(SIMPLEGRPC_IN)/simplegrpc.sysl
simplegrpc-gen: OUT=$(SIMPLEGRPC_OUT)/$(SIMPLESG)
simplegrpc-gen: EXT_LIB_DIR=simplegrpc
simplegrpc-gen: PROTO_IN=$(SIMPLEGRPC_IN)
simplegrpc-gen: PROTO_OUT=$(SIMPLEGRPC_OUT)/simplepb

simplegrpc-gen: $(SIMPLEGRPC_SYSL_FILES) $(SIMPLEGRPC_PROTO_FILES)

.PHONY: simplegrpc-clean
simplegrpc-clean:
	rm $(SIMPLEGRPC_SYSL_FILES) $(SIMPLEGRPC_PROTO_FILES)

clean: simplegrpc-clean
gen: simplegrpc-gen

# SySL build rule and file list
$(SIMPLEGRPC_OUT)/$(SIMPLESG)/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

$(SIMPLEGRPC_OUT)/simplepb/%.pb.go : $(SIMPLEGRPC_IN)/%.proto
	$(run-protoc)
