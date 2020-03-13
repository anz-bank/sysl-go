# Simple Server
SIMPLE_IN=codegen/testdata/simple
SIMPLE_OUT=codegen/tests/simple

SIMPLE_ALL_FILES=$(SIMPLE_ERRORS) $(SIMPLE_TYPES) $(SIMPLE_INTERFACE) $(SIMPLE_HANDLER) $(SIMPLE_ROUTER) $(SIMPLE_CLIENT)

SIMPLE_ERRORS=$(SIMPLE_OUT)/error_types.go
SIMPLE_TYPES=$(SIMPLE_OUT)/types.go
SIMPLE_INTERFACE=$(SIMPLE_OUT)/serviceinterface.go
SIMPLE_HANDLER=$(SIMPLE_OUT)/servicehandler.go
SIMPLE_ROUTER=$(SIMPLE_OUT)/requestrouter.go
SIMPLE_CLIENT=$(SIMPLE_OUT)/service.go

.PHONY: simple-gen
simple-gen: APP=Simple
simple-gen: MODEL=$(SIMPLE_IN)/simple.sysl
simple-gen: OUT=$(SIMPLE_OUT)

simple-gen: $(SIMPLE_ALL_FILES)

.PHONY: simple-clean
simple-clean:
	rm $(SIMPLE_ALL_FILES)

clean: simple-clean
gen: simple-gen

$(SIMPLE_ERRORS): $(TRANSFORMS)/svc_error_types.sysl $(MODEL)
	$(run-sysl)

$(SIMPLE_TYPES): $(TRANSFORMS)/svc_types.sysl $(MODEL)
	$(run-sysl)

$(SIMPLE_INTERFACE): $(TRANSFORMS)/svc_interface.sysl $(MODEL)
	$(run-sysl)

$(SIMPLE_HANDLER): $(TRANSFORMS)/svc_handler.sysl $(MODEL)
	$(run-sysl)

$(SIMPLE_ROUTER): $(TRANSFORMS)/svc_router.sysl $(MODEL)
	$(run-sysl)

$(SIMPLE_CLIENT): $(TRANSFORMS)/svc_client.sysl $(MODEL)
	$(run-sysl)
