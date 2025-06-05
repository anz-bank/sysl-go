SYSLGO_SYSL = specs/template.sysl
SYSLGO_PACKAGES = template
SYSLGO_APP.template = Template
TEMPLATE = template.arraiz "custom_argument"
PKGPATH = simple_rest

# note: we don't measure anything about code generated
# files, otherwise this test will "fail" any time the rules
# for the codegen are slightly tweaked. instead this test
# only checks to see if the custom expected files are present
EXPECTED_OUTPUT_DIR = golden
EXPECTED_FILES = $(shell find $(EXPECTED_OUTPUT_DIR) -type f)
ACTUAL_FILES = $(patsubst $(EXPECTED_OUTPUT_DIR)/%,%,$(EXPECTED_FILES))

CLEAN = go.mod go.sum template.arraiz $(ACTUAL_FILES)
include ../common.mk

test: template.arraiz go.mod internal/gen/pkg/servers/template
	goimports -w internal/gen/pkg/servers/template
	for f in $(ACTUAL_FILES); do echo $$f; diff -w $$f $(EXPECTED_OUTPUT_DIR)/$$f || FAILED=1; done; [ "$$FAILED" != 1 ]
	$(MAKE) clean

.PHONY: test clean

include codegen.mk

$(ACTUAL_FILES) &: internal/gen/pkg/servers/template
	tree

template.arraiz: template.arrai
	arrai bundle -o template.arraiz template.arrai

go.mod: template.arraiz
	go mod init custom_template

internal/gen/pkg/servers/template: template.arraiz
