# Requires protoc, protoc-gen-go and goimports.


# Hack: globally force any shell commands to run inside
# a bash shell with saner defaults set. If this is not done
# then multi-statement shell recipes such as run-arrai can
# fail in a way that is not noticed by make.
SHELL=/bin/bash -o pipefail -o errexit

all: test-arrai test check-coverage lint auto-test ## Tests, lints and checks coverage

.PHONY: all clean

# -- Lint ----------------------------------------------------------------------
lint: ## Lint Go Source Code
	golangci-lint run

check-tidy: ## Check go.mod and go.sum is tidy
	go mod tidy && git diff --exit-code HEAD -- ":(top)go.mod" ":(top)go.sum"

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

ALL_TESTS = $(sort $(dir $(wildcard codegen/arrai/auto/tests/*/Makefile)))

auto-test: $(patsubst %,auto-test-%.dummy,$(ALL_TESTS))
auto-test-%/.dummy:
	$(MAKE) -C $*

update-auto-test-go-mod: $(patsubst %,go-mod-%.dummy,$(ALL_TESTS)) ## Update go.mod and go.sum files within auto tests
go-mod-%/.dummy:
	cd $* && go mod download && go mod tidy

clean: $(patsubst %,clean-%.dummy,$(ALL_TESTS))
	rm -f $(COVERFILE)
clean-%/.dummy:
	$(MAKE) -C $* clean

ALL_GRPC_TESTS = $(sort $(dir $(wildcard codegen/arrai/auto/tests/*grpc*/Makefile)))

update-auto-test-proto-pb: $(patsubst %,proto-pb-%.dummy,$(ALL_GRPC_TESTS)) ## Update protos within auto tests
proto-pb-%/.dummy:
	$(MAKE) -C $* -B protos

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

protos: core/testdata/proto/test.pb.go
.PHONY: protos

%.pb.go: %.proto
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$^

go-work:
	if [ ! -f go.work ]; then go work init; fi
	find . -name go.mod -execdir go work use . \;
.PHONY: go-work
