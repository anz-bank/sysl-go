module simple_rest

go 1.14

require (
	github.com/anz-bank/pkg v0.0.11
	github.com/anz-bank/sysl-go v0.84.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/rickb777/date v1.14.0
	github.com/sethvargo/go-retry v0.1.0
	github.com/spf13/afero v1.3.4
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.29.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/anz-bank/sysl-go => ../../../../..
