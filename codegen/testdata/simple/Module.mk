# Simple Server
SIMPLE_IN=simple
SIMPLE_OUT=codegen/tests/simple

SIMPLE_ALL_FILES=$(SIMPLE_ERRORS) $(SIMPLE_TYPES) $(SIMPLE_INTERFACE) $(SIMPLE_HANDLER) $(SIMPLE_ROUTER) $(SIMPLE_CLIENT) $(SIMPLE_APP)
SIMPLE_ERRORS=$(SIMPLE_OUT)/error_types.go
SIMPLE_TYPES=$(SIMPLE_OUT)/types.go
SIMPLE_INTERFACE=$(SIMPLE_OUT)/serviceinterface.go
SIMPLE_HANDLER=$(SIMPLE_OUT)/servicehandler.go
SIMPLE_ROUTER=$(SIMPLE_OUT)/requestrouter.go
SIMPLE_CLIENT=$(SIMPLE_OUT)/service.go
SIMPLE_APP=$(SIMPLE_OUT)/app.go

.PHONY: simple-gen
simple-gen: APP=Simple
simple-gen: MODEL=$(SIMPLE_IN)/simple.sysl
simple-gen: OUT=$(SIMPLE_OUT)

simple-gen: $(SIMPLE_ALL_FILES)

.PHONY: simple-clean
simple-clean:
	rm $(SIMPLE_ALL_FILES)

SIMPLE_SYSL=$(TEST_IN_DIR)/$(SIMPLE_IN)/simple.sysl

$(SIMPLE_ERRORS): $(TRANSFORMS)/svc_error_types.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_TYPES): $(TRANSFORMS)/svc_types.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_INTERFACE): $(TRANSFORMS)/svc_interface.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_HANDLER): $(TRANSFORMS)/svc_handler.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_ROUTER): $(TRANSFORMS)/svc_router.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_CLIENT): $(TRANSFORMS)/svc_client.sysl $(SIMPLE_SYSL)
	$(run-sysl)

$(SIMPLE_APP): $(TRANSFORMS)/svc_app.sysl $(MODEL)
	$(run-sysl)

# Deps Server
DEPS_OUT=codegen/tests/deps

DEPS_ALL_FILES=$(DEPS_ERRORS) $(DEPS_TYPES) $(DEPS_INTERFACE) $(DEPS_HANDLER) $(DEPS_ROUTER) $(DEPS_CLIENT)

DEPS_ERRORS=$(DEPS_OUT)/error_types.go
DEPS_TYPES=$(DEPS_OUT)/types.go
DEPS_INTERFACE=$(DEPS_OUT)/serviceinterface.go
DEPS_HANDLER=$(DEPS_OUT)/servicehandler.go
DEPS_ROUTER=$(DEPS_OUT)/requestrouter.go
DEPS_CLIENT=$(DEPS_OUT)/service.go

.PHONY: deps-gen
deps-gen: APP=Deps
deps-gen: MODEL=$(SIMPLE_IN)/deps.sysl
deps-gen: OUT=$(DEPS_OUT)

deps-gen: $(DEPS_ALL_FILES)

.PHONY: deps-clean
deps-clean:
	rm $(DEPS_ALL_FILES)

$(DEPS_ERRORS): $(TRANSFORMS)/svc_error_types.sysl $(MODEL)
	$(run-sysl)

$(DEPS_TYPES): $(TRANSFORMS)/svc_types.sysl $(MODEL)
	$(run-sysl)

$(DEPS_INTERFACE): $(TRANSFORMS)/svc_interface.sysl $(MODEL)
	$(run-sysl)

$(DEPS_HANDLER): $(TRANSFORMS)/svc_handler.sysl $(MODEL)
	$(run-sysl)

$(DEPS_ROUTER): $(TRANSFORMS)/svc_router.sysl $(MODEL)
	$(run-sysl)

$(DEPS_CLIENT): $(TRANSFORMS)/svc_client.sysl $(MODEL)
	$(run-sysl)

clean: simple-clean deps-clean downstream-clean
gen: deps-gen downstream-gen simple-gen

# Downstream System
DOWNSTREAM_OUT=codegen/tests/downstream

DOWNSTREAM_ALL_FILES=$(DOWNSTREAM_ERRORS) $(DOWNSTREAM_TYPES) $(DOWNSTREAM_INTERFACE) $(DOWNSTREAM_HANDLER) $(DOWNSTREAM_ROUTER) $(DOWNSTREAM_CLIENT)

DOWNSTREAM_ERRORS=$(DOWNSTREAM_OUT)/error_types.go
DOWNSTREAM_TYPES=$(DOWNSTREAM_OUT)/types.go
DOWNSTREAM_INTERFACE=$(DOWNSTREAM_OUT)/serviceinterface.go
DOWNSTREAM_HANDLER=$(DOWNSTREAM_OUT)/servicehandler.go
DOWNSTREAM_ROUTER=$(DOWNSTREAM_OUT)/requestrouter.go
DOWNSTREAM_CLIENT=$(DOWNSTREAM_OUT)/service.go

.PHONY: downstream-gen
downstream-gen: APP=Downstream
downstream-gen: MODEL=$(SIMPLE_IN)/downstream.sysl
downstream-gen: OUT=$(DOWNSTREAM_OUT)

downstream-gen: $(DOWNSTREAM_ALL_FILES)

.PHONY: downstream-clean
downstream-clean:
	rm $(DOWNSTREAM_ALL_FILES)

$(DOWNSTREAM_ERRORS): $(TRANSFORMS)/svc_error_types.sysl $(MODEL)
	$(run-sysl)

$(DOWNSTREAM_TYPES): $(TRANSFORMS)/svc_types.sysl $(MODEL)
	$(run-sysl)

$(DOWNSTREAM_INTERFACE): $(TRANSFORMS)/svc_interface.sysl $(MODEL)
	$(run-sysl)

$(DOWNSTREAM_HANDLER): $(TRANSFORMS)/svc_handler.sysl $(MODEL)
	$(run-sysl)

$(DOWNSTREAM_ROUTER): $(TRANSFORMS)/svc_router.sysl $(MODEL)
	$(run-sysl)

$(DOWNSTREAM_CLIENT): $(TRANSFORMS)/svc_client.sysl $(MODEL)
	$(run-sysl)
