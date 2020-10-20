SYSLGO_SYSL=specs/template.sysl
SYSLGO_PACKAGES=template
SYSLGO_APP.template = Template
SYMLINK=1
PKGPATH=simple_rest
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

EXPECTED_OUTPUT_DIR = golden
EXPECTED_FILES = $(shell find $(EXPECTED_OUTPUT_DIR) -type f)
ACTUAL_FILES = $(patsubst $(EXPECTED_OUTPUT_DIR)/%,%,$(EXPECTED_FILES))

test: $(ACTUAL_FILES) $(EXPECTED_FILES)
	goimports -w cmd/template
	for f in $(ACTUAL_FILES); do echo $$f; diff -w $$f $(EXPECTED_OUTPUT_DIR)/$$f || FAILED=1; done; [ "$$FAILED" != 1 ]
	$(MAKE) clean

$(ACTUAL_FILES) &: internal/gen/pkg/servers/template

.PHONY: test clean

include codegen.mk

clean:
	rm -rf internal/gen/ $(ACTUAL_FILES)
