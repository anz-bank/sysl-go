SYSLGO_SYSL=specs/template.sysl
SYSLGO_PACKAGES=template
SYSLGO_APP.template = Template
TEMPLATE=template.arraiz "custom_argument"
SYMLINK=1
PKGPATH=simple_rest
NO_DOCKER=1
SYSL_GO_ROOT=../../../../../

EXPECTED_OUTPUT_DIR = golden
EXPECTED_FILES = $(shell find $(EXPECTED_OUTPUT_DIR) -type f)
ACTUAL_FILES = $(foreach F,$(EXPECTED_FILES), $(F:$(EXPECTED_OUTPUT_DIR)/%=%))

test: template.arraiz go.mod gen/pkg/servers/Template
	for f in $(ACTUAL_FILES); do diff $$f $(EXPECTED_OUTPUT_DIR)/$$f; done
	$(MAKE) clean

.PHONY: test clean

include codegen.mk

$(ACTUAL_FILES) &: gen/pkg/servers/Template

template.arraiz: template.arrai
	arrai bundle -o template.arraiz template.arrai

go.mod: template.arraiz
	go mod init custom_template

# remove go.mod to not hinder with arrai bundling, imports cannot be outside a go.mod
clean:
	rm -rf gen/ go.mod go.sum template.arraiz $(ACTUAL_FILES)
