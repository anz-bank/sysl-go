SYSLGO_SYSL=specs/template.sysl
SYSLGO_PACKAGES=template
SYSLGO_APP.template = Template
SYMLINK=1
PKGPATH=simple_rest
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

EXPECTED_OUTPUT_DIR = golden
EXPECTED_FILES = $(shell find $(EXPECTED_OUTPUT_DIR) -type f)
ACTUAL_FILES = $(foreach F,$(EXPECTED_FILES), $(F:$(EXPECTED_OUTPUT_DIR)/%=%))

test: $(ACTUAL_FILES) $(EXPECTED_FILES)
	for f in $(ACTUAL_FILES); do diff $$f $(EXPECTED_OUTPUT_DIR)/$$f; done
	$(MAKE) clean

$(ACTUAL_FILES) &: gen/pkg/servers/Template

.PHONY: test clean

include codegen.mk

clean:
	rm -rf gen/ $(ACTUAL_FILES)
