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

.PHONY: simple-clean
simple-clean:
	rm -f $(SIMPLE_ALL_FILES)

.PHONY: simple-gen
simple-gen: APP=Simple
simple-gen: MODEL=$(TEST_IN_DIR)/$(SIMPLE_IN)/sysl.json

.PHONY: Simple
SIMPLE=Simple
simple-gen: $(SIMPLE) $(MODEL)
	$(run-arrai)

# Deps Server
DEPS_OUT=codegen/tests/deps
DEPS_IN=deps

DEPS_ALL_FILES=$(DEPS_ERRORS) $(DEPS_TYPES) $(DEPS_INTERFACE) $(DEPS_HANDLER) $(DEPS_ROUTER) $(DEPS_CLIENT)

DEPS_ERRORS=$(DEPS_OUT)/error_types.go
DEPS_TYPES=$(DEPS_OUT)/types.go
DEPS_INTERFACE=$(DEPS_OUT)/serviceinterface.go
DEPS_HANDLER=$(DEPS_OUT)/servicehandler.go
DEPS_ROUTER=$(DEPS_OUT)/requestrouter.go
DEPS_CLIENT=$(DEPS_OUT)/service.go

.PHONY: deps-clean
deps-clean:
	rm -f $(DEPS_ALL_FILES)

.PHONY: deps-gen
deps-gen: APP=Deps
deps-gen: MODEL=$(TEST_IN_DIR)/$(DEPS_IN)/sysl.json

.PHONY: Deps
DEPS=Deps
deps-gen: $(DEPS) $(MODEL)
	$(run-arrai)

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

.PHONY: downstream-clean
downstream-clean:
	rm -f $(DOWNSTREAM_ALL_FILES)

.PHONY: downstream-gen
downstream-gen: APP=Downstream
downstream-gen: MODEL=$(TEST_IN_DIR)/$(DOWNSTREAM_IN)/sysl.json

.PHONY: Downstream
DOWNSTREAM=Downstream
downstream-gen: $(DOWNSTREAM) $(MODEL)
	$(run-arrai)
	