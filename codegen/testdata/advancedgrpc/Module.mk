# Simple Server
ADVANCEDGRPC_IN=advancedgrpc
ADVANCEDGRPC_OUT=$(TEST_OUT_DIR)/advancedgrpc

ADVANCEDGRPC_SYSL_FILES=$(addprefix $(ADVANCEDGRPC_OUT)/, $(GRPC_SERVER_FILES))
ADVANCEDGRPC_PROTO_FILES=$(TEST_OUT_DIR)/advancedpb/advancedgrpc.pb.go
ADVANCEDGRPC_APP=$(ADVANCEDGRPC_OUT)/app.go

.PHONY: advancedgrpc-gen
advancedgrpc-gen: APP=advancedgrpc
advancedgrpc-gen: MODEL=$(ADVANCEDGRPC_IN)/advancedgrpc.sysl
advancedgrpc-gen: OUT=$(ADVANCEDGRPC_OUT)
advancedgrpc-gen: PROTO_IN=$(TEST_IN_DIR)/advancedgrpc
advancedgrpc-gen: PROTO_OUT=$(TEST_OUT_DIR)/advancedpb

advancedgrpc-gen: $(ADVANCEDGRPC_PROTO_FILES)

.PHONY: advancedgrpc-clean
advancedgrpc-clean:
	rm -f $(ADVANCEDGRPC_SYSL_FILES) $(ADVANCEDGRPC_PROTO_FILES)

clean: advancedgrpc-clean
gen: advancedgrpc-gen


codegen/tests/advancedpb/%.pb.go : $(TEST_IN_DIR)/$(ADVANCEDGRPC_IN)/%.proto
	$(run-protoc)
