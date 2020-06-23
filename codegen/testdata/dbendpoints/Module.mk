# DB Server
DB_IN=dbendpoints
DB_OUT=codegen/tests/dbendpoints

DB_ALL_FILES=$(DB_TYPES) $(DB_INTERFACE) $(DB_HANDLER) $(DB_ROUTER) $(DB_CLIENT)

DB_TYPES=$(DB_OUT)/types.go
DB_INTERFACE=$(DB_OUT)/serviceinterface.go
DB_HANDLER=$(DB_OUT)/servicehandler.go
DB_ROUTER=$(DB_OUT)/requestrouter.go
DB_CLIENT=$(DB_OUT)/service.go

.PHONY: db-gen
db-gen: APP=DbEndpoints
db-gen: MODEL=$(DB_IN)/dbendpoints.sysl
db-gen: OUT=$(DB_OUT)

db-gen: $(DB_ALL_FILES)

.PHONY: db-clean
db-clean:
	rm -f $(DB_ALL_FILES)

clean: db-clean
gen: db-gen

DB_SYSL=$(TEST_IN_DIR)/$(DB_IN)/dbendpoints.sysl

$(DB_TYPES): $(TRANSFORMS)/svc_types.sysl $(DB_SYSL)
	$(run-sysl)

$(DB_INTERFACE): $(TRANSFORMS)/svc_interface.sysl $(DB_SYSL)
	$(run-sysl)

$(DB_HANDLER): $(TRANSFORMS)/svc_handler.sysl $(DB_SYSL)
	$(run-sysl)

$(DB_ROUTER): $(TRANSFORMS)/svc_router.sysl $(DB_SYSL)
	$(run-sysl)

$(DB_CLIENT): $(TRANSFORMS)/svc_client.sysl $(DB_SYSL)
	$(run-sysl)
