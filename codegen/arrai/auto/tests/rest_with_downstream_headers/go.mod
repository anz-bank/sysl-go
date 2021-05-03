module rest_with_downstream_headers

go 1.16

require (
	github.com/anz-bank/sysl-go v0.189.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/sethvargo/go-retry v0.1.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/anz-bank/sysl-go => ../../../../..
