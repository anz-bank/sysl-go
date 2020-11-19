module rest_custom_middleware

go 1.14

require (
	github.com/anz-bank/pkg v0.0.27
	github.com/anz-bank/sysl-go v0.84.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/sethvargo/go-retry v0.1.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)

replace github.com/anz-bank/sysl-go => ../../../../..
