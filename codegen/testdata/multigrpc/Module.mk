# Test suite input and output locations
MULTIGRPC_IN=codegen/testdata/multigrpc
MULTIGRPC_OUT=$(TEST_DIR)/multigrpc

CARDS_FILES=$(addprefix $(MULTIGRPC_OUT)/cards/, $(GRPC_SERVER_FILES))
WALLET_FILES=$(addprefix $(MULTIGRPC_OUT)/wallet/, $(GRPC_SERVER_FILES))
MULTIGRPC_PROTO_FILES=$(addprefix $(MULTIGRPC_OUT)/cardspb/, cards.pb.go wallet_api.pb.go cards_api.pb.go)

# Cards build
.PHONY: cards-gen
cards-gen: APP=Cards
cards-gen: MODEL=$(MULTIGRPC_IN)/cards_api.sysl
cards-gen: OUT=$(MULTIGRPC_OUT)/cards
cards-gen: EXT_LIB_DIR=multigrpc

cards-gen: $(CARDS_FILES)

# Wallet build
.PHONY: wallet-gen
wallet-gen: APP=Wallet
wallet-gen: MODEL=$(MULTIGRPC_IN)/wallet_api.sysl
wallet-gen: OUT=$(MULTIGRPC_OUT)/wallet
wallet-gen: EXT_LIB_DIR=multigrpc

wallet-gen: $(WALLET_FILES)

# Generate all files for this test suite
.PHONY: multigrpc-gen
multigrpc-gen: PROTO_IN=$(MULTIGRPC_IN)
multigrpc-gen: PROTO_OUT=$(MULTIGRPC_OUT)/cardspb

multigrpc-gen: cards-gen wallet-gen $(MULTIGRPC_PROTO_FILES)

# Clean all the generated files
.PHONY: multigrpc-clean
multigrpc-clean:
	rm $(CARDS_FILES) $(WALLET_FILES) $(MULTIGRPC_PROTO_FILES)

# Add to top level gen and clean
gen: multigrpc-gen
clean: multigrpc-clean

# Cards build rule and file liet
$(MULTIGRPC_OUT)/cards/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

# Wallet build rule and file list
$(MULTIGRPC_OUT)/wallet/%.go: $(TRANSFORMS)/%.sysl
	$(run-sysl)

$(MULTIGRPC_OUT)/cardspb/%.pb.go : $(MULTIGRPC_IN)/%.proto
	$(run-protoc)
