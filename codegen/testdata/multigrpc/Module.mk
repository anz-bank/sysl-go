# Test suite input and output locations
MULTIGRPC_IN=multigrpc
MULTIGRPC_OUT=$(TEST_OUT_DIR)/multigrpc

CARDS_FILES=$(addprefix $(TEST_OUT_DIR)/cards/, $(GRPC_SERVER_FILES))
WALLET_FILES=$(addprefix $(TEST_OUT_DIR)/wallet/, $(GRPC_SERVER_FILES))

MULTIGRPC_PROTO_FILES=$(addprefix $(TEST_OUT_DIR)/cardspb/, cards.pb.go wallet_api.pb.go cards_api.pb.go)

# Cards build
.PHONY: cards-gen
cards-gen: APP=Cards
cards-gen: MODEL=$(MULTIGRPC_IN)/cards_api.sysl
cards-gen: OUT=$(TEST_OUT_DIR)/cards

cards-gen: $(CARDS_FILES)

# Wallet build
.PHONY: wallet-gen
wallet-gen: APP=Wallet
wallet-gen: MODEL=$(MULTIGRPC_IN)/wallet_api.sysl
wallet-gen: OUT=$(TEST_OUT_DIR)/wallet

wallet-gen: $(WALLET_FILES)

# Generate all files for this test suite
.PHONY: multigrpc-gen
multigrpc-gen: PROTO_IN=$(TEST_IN_DIR)/$(MULTIGRPC_IN)
multigrpc-gen: PROTO_OUT=$(TEST_OUT_DIR)/cardspb

multigrpc-gen: cards-gen wallet-gen $(MULTIGRPC_PROTO_FILES)

# Clean all the generated files
.PHONY: multigrpc-clean
multigrpc-clean:
	rm $(CARDS_FILES) $(WALLET_FILES) $(MULTIGRPC_PROTO_FILES)

# Add to top level gen and clean
gen: multigrpc-gen
clean: multigrpc-clean

# Cards build rule and file liet
codegen/tests/cards/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

# Wallet build rule and file list
codegen/tests/wallet/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

codegen/tests/cardspb/%.pb.go : $(TEST_IN_DIR)/$(MULTIGRPC_IN)/%.proto
	$(run-protoc)
