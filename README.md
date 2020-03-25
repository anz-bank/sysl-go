# sysl-go

Communication library used by SYSL-generated code written in Go.

## 1.1. Getting Started

Go get the repository

    go get github.com/anz-bank/sysl-go

### 1.1.1. Local Development

### 1.1.2. Prerequisites

Ensure your environment provides:

- [go 1.13](https://golang.org/doc/install)
- [golangci-lint 1.23](https://github.com/golangci/golangci-lint)
- [protobuf 3.11.4](https://github.com/protocolbuffers/protobuf/)
- `make`
- `jq`
- proto3 and gRPC
  - https://github.com/protocolbuffers/protobuf/releases
  - https://github.com/golang/protobuf
  - https://github.com/grpc/grpc

On OSX, after installing [go 1.12.9](https://golang.org/doc/install) run

    brew install golangci/tap/golangci-lint make jq curl protoc-gen-go grpc

### 1.1.3 Development

Test and lint everything with

    make

View all relevant make targets with

    make help

View test coverage in the browser with

    make coverage
