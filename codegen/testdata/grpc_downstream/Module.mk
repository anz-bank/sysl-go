codegen/tests/grpc_downstream/grpc_downstream.pb.go: PROTO_IN=$(TEST_IN_DIR)/grpc_downstream
codegen/tests/grpc_downstream/grpc_downstream.pb.go: PROTO_OUT=$(TEST_OUT_DIR)/grpc_downstream

codegen/tests/grpc_downstream/grpc_downstream.pb.go : $(TEST_IN_DIR)/grpc_downstream/grpc_downstream.proto
	$(run-protoc)
