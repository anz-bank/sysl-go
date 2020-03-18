# sysl-go

Communication library used by SYSL-generated code written in Go.

## 1.1. Getting Started

Go get the repository

    go get github.com/anz-bank/sysl-go

### 1.1.1. Local Development

### 1.1.2. Prerequisites

Ensure your environment provides:

- [go 1.12.9](https://golang.org/)
- [golangci-lint 1.17.1](https://github.com/golangci/golangci-lint)
- some working method of obtaining dependencies listed in `go.mod` (working internet access, `GOPROXY`)
- env var `GOFLAGS="-mod=vendor"`
- correctly configured cntlm or alpaca proxy running on localhost at port 3128 (only for updating vendor dependencies)

### 1.1.3. Linting
    golangci-lint run ./...

### 1.1.4. Running the Tests
    go test -v -cover -count=1 `go list ./... | grep -v ./codegen`

To generate and view test coverage in a browser, use this

    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
