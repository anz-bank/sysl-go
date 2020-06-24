# Simple Server
SIMPLEGRPC_IN=simplegrpc
SIMPLEGRPC_OUT=$(TEST_OUT_DIR)/simplegrpc

SIMPLEGRPC_SYSL_FILES=$(addprefix $(SIMPLEGRPC_OUT)/, $(GRPC_SERVER_FILES))
SIMPLEGRPC_PROTO_FILES=$(TEST_OUT_DIR)/simplepb/simplegrpc.pb.go
SIMPLE_APP=$(SIMPLEGRPC_OUT)/app.go

.PHONY: simplegrpc-gen
simplegrpc-gen: APP=SimpleGrpc
simplegrpc-gen: MODEL=$(SIMPLEGRPC_IN)/simplegrpc.sysl
simplegrpc-gen: OUT=$(SIMPLEGRPC_OUT)
simplegrpc-gen: PROTO_IN=$(TEST_IN_DIR)/simplegrpc
simplegrpc-gen: PROTO_OUT=$(TEST_OUT_DIR)/simplepb

simplegrpc-gen: $(SIMPLEGRPC_SYSL_FILES) $(SIMPLEGRPC_PROTO_FILES) $(SIMPLE_APP)

.PHONY: simplegrpc-clean
simplegrpc-clean:
	rm -f $(SIMPLEGRPC_SYSL_FILES) $(SIMPLEGRPC_PROTO_FILES)

clean: simplegrpc-clean
gen: simplegrpc-gen

# SySL build rule and file list
$(SIMPLEGRPC_OUT)/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

$(SIMPLE_APP): $(TRANSFORMS)/svc_app.sysl $(MODEL)
	$(run-sysl)	

codegen/tests/simplepb/%.pb.go : $(TEST_IN_DIR)/$(SIMPLEGRPC_IN)/%.proto
	$(run-protoc)
