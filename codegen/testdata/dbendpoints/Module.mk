# DB Server
DB_IN=dbendpoints
DB_OUT=codegen/tests/dbendpoints

DB_ALL_FILES=$(DB_TYPES) $(DB_INTERFACE) $(DB_HANDLER) $(DB_ROUTER) $(DB_CLIENT)

DB_TYPES=$(DB_OUT)/types.go
DB_INTERFACE=$(DB_OUT)/serviceinterface.go
DB_HANDLER=$(DB_OUT)/servicehandler.go
DB_ROUTER=$(DB_OUT)/requestrouter.go
DB_CLIENT=$(DB_OUT)/service.go

.PHONY: db-clean
db-clean:
	rm -f $(DB_ALL_FILES)

.PHONY: db-gen
db-gen: APP=DbEndpoints
db-gen: MODEL=$(TEST_IN_DIR)/$(DB_IN)/sysl.json

.PHONY: DbEndpoints
DBENDPOINTS=DbEndpoints
db-gen: $(DBENDPOINTS) $(MODEL)
	$(run-arrai)
	