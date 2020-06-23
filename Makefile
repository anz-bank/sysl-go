all: test check-coverage arrai-nodiff lint tidy ## Tests, lints and checks coverage

clean:  ## Remove generated files

.PHONY: all clean

# -- Lint ----------------------------------------------------------------------
lint: ## Lint Go Source Code
	golangci-lint run

tidy: ## Run go mod tidy
	go mod tidy

check-tidy: ## Check go.mod and go.sum is tidy
	go mod tidy && test -z "$$(git status --porcelain)"

.PHONY: lint tidy check-tidy

# -- Test ----------------------------------------------------------------------
COVERFILE=coverage.out
COVERAGE = 50

test: ## Run all tests
	go test -coverprofile=$(COVERFILE) -tags codeanalysis ./...

check-coverage: test  ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

coverage: test  ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

clean:
	rm -f $(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ && $$3 < $(COVERAGE) {exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: check-coverage coverage test

# --- Utilities ---------------------------------------------------------------
COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)
BOLD          = $(shell tput bold 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9_-]+$$/ { printf "$(BOLD)$(COLOUR_WHITE)%-20s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: help

# -- Codegen ----------------------------------------------------------------------
# Transform settings - common across all code generation
TRANSFORMS=codegen/transforms

# Input models for code generation
TEST_IN_DIR=codegen/testdata
# Base directory for code generation output
TEST_OUT_DIR=codegen/tests

define run-sysl
sysl codegen \
	--dep-path github.com/anz-bank/sysl-go/$(TEST_OUT_DIR)  \
	--root $(TEST_IN_DIR) \
	--root-transform . \
	--transform $< \
	--grammar codegen/grammars/go.gen.g \
	--start goFile \
	--outdir $(OUT) \
	--basepath github.com/anz-bank/sysl-go/$(TEST_OUT_DIR) \
	--app-name $(APP) \
	$(MODEL)
goimports -w $@
endef

# PROTO_IN and PROTO_OUT are defined in Make modules
define run-protoc
protoc --proto_path=$(PROTO_IN) --go_out=plugins=grpc:$(PROTO_OUT) $^
goimports -w $@
endef

# Output files generated for gRPC servers
GRPC_SERVER_FILES=grpc_interface.go grpc_handler.go

gen: ## Run sysl codegen and proto codegen

.PHONY: gen

include $(TEST_IN_DIR)/*/Module.mk


# Arr.ai codegen

ARRAI_TRANSFORMS=codegen/arrai

targets = \
	dbendpoints \
	deps \
	downstream \
	simple \
	simplegrpc

deps.app = Deps
deps.groups = rest-service

dbendpoints.app = DbEndpoints
dbendpoints.groups = rest-service

downstream.app = Downstream
downstream.groups = rest-service

simple.app = Simple
simple.groups = rest-app

simplegrpc.app = SimpleGrpc
simplegrpc.groups = grpc-app

.PRECIOUS: $(patsubst %,codegen/testdata/%/sysl.json,$(targets))

codegen/testdata/%/sysl.json: codegen/testdata/%/*.sysl
	sysl pb --mode=json --root $(TEST_IN_DIR) $*/$*.sysl > $@ || rm -f $@

ARRAI_OUT=codegen/arrai/tests

$(ARRAI_OUT)/% : codegen/testdata/%/sysl.json \
		$(patsubst %,$(ARRAI_TRANSFORMS)/%.arrai,\
			grpc_handler \
			grpc_interface \
			service \
			svc_app \
			svc_service \
			svc_error_types \
			svc_types \
			svc_interface \
			svc_handler \
			svc_router \
			go \
			sysl \
		)
	mkdir -p $@
	$(ARRAI_TRANSFORMS)/service.arrai github.com/anz-bank/sysl-go/codegen/tests $< $($*.app) "$($*.groups)" | tar xf - -C $@
	goimports -w $@ || :
	touch $@

arrai: $(patsubst %,codegen/arrai/tests/%,$(targets))

.PHONY: arrai-nodiff
arrai-nodiff: $(patsubst %,codegen/arrai/tests/%.nodiff,$(targets))

codegen/arrai/tests/%.nodiff: $(ARRAI_OUT)/% $(TIDY)
	diff -rwuBq $(TEST_OUT_DIR)/$* $</ | awk 'BEGIN { err = 0 } END { exit err } /^Files .* differ$$/ { print; err = 1 }' && touch $@

include $(ARRAI_TRANSFORMS)/Module.mk
