all: test check-coverage lint tidy ## Tests, lints and checks coverage

clean:  ## Remove generated files

.PHONY: all clean

# -- Lint ----------------------------------------------------------------------
lint: ## Lint Go Source Code
	golangci-lint run

tidy: ## Run go mod tidy
	go mod tidy

.PHONY: lint tidy

# -- Test ----------------------------------------------------------------------
COVERFILE=coverage.out
COVERAGE = 50

test: ## Run all tests
	go test -count=1 -cover -coverprofile=$(COVERFILE) -tags codeanalysis ./...

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
GRAMMAR=codegen/grammars/go.gen.g
START=goFile
TEST_DIR=codegen/tests
BATH_PATH=github.com/anz-bank/sysl-go/codegen/tests/

define run-sysl
sysl codegen --dep-path github.com/anz-bank/sysl-go/$(TEST_DIR)/$(EXT_LIB_DIR)  --root . --root-transform . --transform $< --grammar $(GRAMMAR) --start $(START) --outdir $(OUT) --basepath $(BATH_PATH) --app-name $(APP) $(MODEL)
goimports -w $@
endef

run-protoc=protoc --proto_path=$(PROTO_IN) --go_out=plugins=grpc:$(PROTO_OUT) $^

GRPC_SERVER_FILES=grpc_interface.go grpc_handler.go

gen: ## Run sysl codegen and proto codegen

.PHONY: gen

include codegen/testdata/*/Module.mk
