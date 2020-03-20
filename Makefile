COVERFILE=coverage.out

# Transform settings - common across all code generation
TRANSFORMS=codegen/transforms
GRAMMAR=codegen/grammars/go.gen.g
START=goFile
TEST_DIR=codegen/tests

define run-sysl
sysl codegen --dep-path github.com/anz-bank/sysl-go/$(TEST_DIR)/$(EXT_LIB_DIR)  --root . --root-transform . --transform $< --grammar $(GRAMMAR) --start $(START) --outdir $(OUT) --app-name $(APP) $(MODEL)
goimports -w $@
endef

run-protoc=protoc --proto_path=$(PROTO_IN) --go_out=plugins=grpc:$(PROTO_OUT) $^

GRPC_SERVER_FILES=grpc_interface.go grpc_handler.go

all: clean gen test lint

tidy:
	go mod tidy

gen:

test: gen
	go test -count=1 -cover -tags codeanalysis ./...

test-coverage:
	go test -coverprofile=$(COVERFILE) ./...

lint: gen
	golangci-lint run

cover: test-coverage
	go tool cover -html=$(COVERFILE)

cleanall: clean tidy

clean:
	go clean ./...

.PHONY: gen test test-coverage lint cover clean cleanall tidy

include codegen/testdata/*/Module.mk
