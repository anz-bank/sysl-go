SYSLGO_SYSL=specs/template.sysl
SYSLGO_PACKAGES=template
SYSLGO_APP.template = Template
TEMPLATE_PATH = template.arraiz
SYMLINK=1
PKGPATH=simple_rest
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

EXPECTED_OUTPUT_DIR = golden
EXPECTED_FILES = $(shell find $(EXPECTED_OUTPUT_DIR) -type f)
ACTUAL_FILES = $(foreach F,$(EXPECTED_FILES), $(F:$(EXPECTED_OUTPUT_DIR)/%=%))

test: template.arraiz go.mod $(ACTUAL_FILES) $(EXPECTED_FILES)
	for f in $(ACTUAL_FILES); do diff $$f $(EXPECTED_OUTPUT_DIR)/$$f; done
	$(MAKE) clean

$(ACTUAL_FILES) &: gen/pkg/servers/Template

template.arraiz: template.arrai
	arrai bundle template.arrai > template.arraiz

go.mod: template.arraiz
	go mod init custom_template

.PHONY: test clean

include codegen.mk

# remove go.mod to not hinder with arrai bundling, imports cannot be outside a go.mod
clean:
	rm -rf gen/ go.mod go.sum template.arraiz $(ACTUAL_FILES)
