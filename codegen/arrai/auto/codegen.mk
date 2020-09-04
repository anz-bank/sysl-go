# Master: github.com/anz-bank/sysl-go/codegen/arrai/auto/codegen.mk
#
# 1. copy this file to your repo and include it in your Makefile
# 2. To use local tools, set environment variable NO_DOCKER=1 when running make.
# 3. If NO_DOCKER=1, set SYSL_GO_ROOT to the local path of the sysl-go repo.

ifndef SYSLFILE
$(error Set SYSLFILE to the path of the Sysl file you want to codegen for.)
endif

ifndef APPS
$(error Set APPS to a list of the apps you want to codegen. e.g.: )
endif

ifndef PKGPATH
PKGPATH = $(shell awk '/^module/{print$$2}' go.mod)
$(warning PKGPATH not set. Inferred from go.mod as $(PKGPATH))
endif

SERVERS_ROOT = gen/pkg/servers
DOCKER = docker
AUTO = arrai --out=dir:$(1) $(SYSL_GO_ROOT)/codegen/arrai/auto/auto.arrai

ifdef NO_DOCKER

PROTOC  = protoc
SYSL    = sysl
AUTOGEN = $(AUTO)
ifndef SYSL_GO_ROOT
$(error Set SYSL_GO_ROOT is required for NO_DOCKER. Set it to the local path of the sysl-go repo.)
endif

else

SYSL_GO_ROOT = /sysl-go
DOCKER_RUN = $(DOCKER) run --rm -v $$(pwd):/work -w /work
PROTOC  = $(DOCKER_RUN) anzbank/protoc-gen-sysl:v0.0.24
SYSL    = $(DOCKER_RUN) anzbank/sysl:v0.185.0
AUTOGEN = $(DOCKER_RUN) sysl-go $(AUTO)

endif

.PHONY: all
all: $(foreach app,$(APPS),$(SERVERS_ROOT)/$(app))

.INTERMEDIATE: model.json
model.json: $(SYSLFILE)
	$(SYSL) pb --mode json $< > $@ || (rm $@ && false)

$(SERVERS_ROOT)/%: model.json
	$(call AUTOGEN,$@) $(PKGPATH)/$@ $< $* =
	find $@ -type d | xargs goimports -w
	touch $@

.PHONY: docker.%
docker.%:
	go mod tidy
	$(DOCKER) build -t $* .
	$(DOCKER) run -p 5751:5751 -v $$(pwd)/:/work $* /work/config.yaml

ifdef NO_DOCKER
# Auto-update the copied version of this file.
codegen = $(filter %codegen.mk,$(MAKEFILE_LIST))
$(codegen): $(SYSL_GO_ROOT)/codegen/arrai/auto/codegen.mk
	@echo Updating codegen.mk
	cp $< $@
endif
