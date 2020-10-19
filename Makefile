# Requires protoc, protoc-gen-go and goimports.


# Hack: globally force any shell commands to run inside
# a bash shell with saner defaults set. If this is not done
# then multi-statement shell recipes such as run-arrai can
# fail in a way that is not noticed by make.
SHELL=/bin/bash -o pipefail -o errexit

all: test-arrai test check-coverage lint check-tidy auto-test ## Tests, lints and checks coverage

.PHONY: all clean

# -- Lint ----------------------------------------------------------------------
lint: ## Lint Go Source Code
	golangci-lint run

check-tidy: ## Check go.mod and go.sum is tidy
	go mod tidy && go mod tidy && git diff --exit-code HEAD -- ":(top)go.mod" ":(top)go.sum"

.PHONY: lint check-tidy

# -- Test (arrai) --------------------------------------------------------------
test-arrai:
	arrai test

.PHONY: test-arrai

# -- Test (go) --------------------------------------------------------------------
COVERFILE=coverage.out
COVERAGE = 50

test: ## Run all tests, apart from auto-test
	go test -coverprofile=$(COVERFILE) -tags codeanalysis ./...

check-coverage: test  ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

coverage: test  ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

auto-test:
	$(MAKE) -C codegen/arrai/auto/tests/simple_rest/
	$(MAKE) -C codegen/arrai/auto/tests/simple_rest_with_downstream/
	$(MAKE) -C codegen/arrai/auto/tests/simple_grpc_with_downstream/
	$(MAKE) -C codegen/arrai/auto/tests/rest_with_conditional_downstream/
	$(MAKE) -C codegen/arrai/auto/tests/rest_with_downstream_headers/
	$(MAKE) -C codegen/arrai/auto/tests/rest_error_downstream/
	$(MAKE) -C codegen/arrai/auto/tests/grpc_custom_server_options/
	$(MAKE) -C codegen/arrai/auto/tests/grpc_custom_dial_options/
	$(MAKE) -C codegen/arrai/auto/tests/template_gen
	$(MAKE) -C codegen/arrai/auto/tests/template_custom_gen
	$(MAKE) -C codegen/arrai/auto/tests/grpc_jwt_authorization/
	$(MAKE) -C codegen/arrai/auto/tests/rest_jwt_authorization/

clean:
	rm -f $(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ && $$3 < $(COVERAGE) {exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: check-coverage coverage test auto-test

# --- Utilities ---------------------------------------------------------------
COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)
BOLD          = $(shell tput bold 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9_-]+$$/ { printf "$(BOLD)$(COLOUR_WHITE)%-20s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
.PHONY: help

docker:
	docker build . -t sysl-go
.PHONY: docker