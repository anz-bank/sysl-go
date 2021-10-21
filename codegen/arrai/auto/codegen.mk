# Master: github.com/anz-bank/sysl-go/codegen/arrai/auto/codegen.mk
#
# 1. Copy this file to your repo and include it in your Makefile.
# 2. Set SYSLGO_SYSL to the .sysl file you want to codegen for.
# 3. Set SYSLGO_PACKAGES to a space-separated list of the Go packages you want
#    to codegen.
# 4. For each app you want to codegen, set SYSLGO_APP.<pkg> to the app name.
#    E.g.: SYSLGO_APP.myapp = MyApp.
# 5. To use local tools, set environment variable NO_DOCKER=1 when running make.
# 6. If NO_DOCKER=1, set SYSL_GO_ROOT to the local path of the sysl-go repo.
#
# A complete example:
#
#     # Makefile
#     SYSLGO_SYSL = specs/frontend/bff.sysl
#
#     SYSLGO_PACKAGES = myapp
#     SYSLGO_APP.myapp = MyApp
#
#     -include local.mk
#     include codegen.mk
#
#     # local.mk (optional)
#     NO_DOCKER = 1
#     SYSL_GO_ROOT = ../sysl-go
#
# 6. Run make
# 7. (optional) run cmd/myapp/main.go with the configuration to run the application.
#    run it with the -h flag to get the configuration documentation.

ifdef SYSLFILE
$(warning WARNING: Using deprecated SYSLFILE. Use SYSLGO_SYSL instead.)
SYSLGO_SYSL = $(SYSLFILE)
endif

ifdef APPS
$(warning WARNING: Using deprecated APPS. Change to SYSLGO_PACKAGES and SYSLGO_APP.<pkg>.)
SYSLGO_PACKAGES = $(APPS)
endif

ifndef SYSLGO_SYSL
$(error Set SYSLGO_SYSL to the path of the Sysl file you want to codegen for.)
endif

ifndef SYSLGO_PACKAGES
$(error Set SYSLGO_PACKAGES to a list of the Go packages you want to codegen. e.g.: )
endif

ifndef PKGPATH
PKGPATH = $(shell awk '/^module/{print$$2}' go.mod)
endif

ifndef TEMPLATE
TEMPLATE = $(SYSL_GO_ROOT)/codegen/arrai/auto/auto.arrai
endif

SERVERS_ROOT = internal/gen/pkg/servers
DOCKER = docker
AUTO = --out=dir:. $(TEMPLATE)

ifdef NO_DOCKER

PROTOC    = protoc
SYSL      = sysl
GOIMPORTS = goimports
AUTOGEN   = arrai $(AUTO)
ifndef SYSL_GO_ROOT
SYSL_GO_ROOT = $(shell go list -m -f '{{.Dir}}' github.com/anz-bank/sysl-go)
ifeq ($(SYSL_GO_ROOT),)
$(error Failed to determine the path of the sysl-go repo, set SYSL_GO_ROOT to the local path)
endif
endif

else

SYSL_GO_ROOT = /sysl-go
SYSL_GO_VERSION = $(shell go list -m -f '{{.Version}}' github.com/anz-bank/sysl-go)
ifeq ($(SYSL_GO_VERSION),)
$(error Failed to determine the version for github.com/anz-bank/sysl-go)
endif
SYSL_GO_IMAGE = anzbank/sysl-go:${SYSL_GO_VERSION}
DOCKER_RUN = $(DOCKER) run --rm -t -v $$(pwd):/work -w /work $(DOCKER_RUN_EXTRA_PARAMS)
PROTOC    = $(DOCKER_RUN) --entrypoint protoc $(SYSL_GO_IMAGE)
SYSL      = $(DOCKER_RUN) --entrypoint sysl $(SYSL_GO_IMAGE)
AUTOGEN   = $(DOCKER_RUN) --entrypoint arrai $(SYSL_GO_IMAGE) $(AUTO)
GOIMPORTS = $(DOCKER_RUN) --entrypoint goimports $(SYSL_GO_IMAGE)

endif

.PHONY: all
all: gen-all-servers

.PHONY: gen-all-servers
gen-all-servers: $(foreach app,$(SYSLGO_PACKAGES),$(SERVERS_ROOT)/$(app))

.INTERMEDIATE: model.json
model.json: $(SYSLGO_SYSL)
# note: if we use > $@ here then if we are running inside a
# terminal, then sysl pb detects that and logs warnings
# to stdout which pollute our JSON output. So use --output $@.
	$(SYSL) pb --mode json --root . $< --output $@ || (rm $@ && false)

$(SERVERS_ROOT)/%: model.json codegen.mk
	$(AUTOGEN) $* $(PKGPATH) $@ $< $(or $(SYSLGO_APP.$*),$*) =
	find $@ -type d | xargs $(GOIMPORTS) -w
	touch $@

.PHONY: docker.%
docker.%:
	go mod tidy
	$(DOCKER) build -t $* .
	$(DOCKER) run -p 5751:5751 -v $$(pwd)/:/work $* /work/config.yaml


ifndef SYMLINK
ifdef NO_DOCKER
# Auto-update the copied version of this file.
codegen = $(filter %codegen.mk,$(MAKEFILE_LIST))
$(codegen): $(SYSL_GO_ROOT)/codegen/arrai/auto/codegen.mk
	@echo Updating codegen.mk
	cp $< $@
endif
endif
