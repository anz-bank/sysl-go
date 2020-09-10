# sysl-go

Communication library used by SYSL-generated code written in Go.

## 1.1. Getting Started

#### Go get the repository

```
go get github.com/anz-bank/sysl-go
```
#### Use docker

```
docker pull anzbank/sysl-go:latest

docker run --rm -v $(pwd):/mount:ro anzbank/sysl-go /sysl-go/codegen/arrai/service.arrai github.com/anz-bank/sysl-template/gen /mount/api/project.json simple rest-app | tar xf - -C gen/simple
```
See [sysl-template](https://github.com/anz-bank/sysl-template) for more examples
### 1.1.1. Local Development

### 1.1.2. Prerequisites

Ensure your environment provides:

- [go 1.14](https://golang.org/doc/install)
- [golangci-lint 1.29.0](https://github.com/golangci/golangci-lint/releases/tag/v1.29.0)
- [protobuf 3.11.4](https://github.com/protocolbuffers/protobuf/)
- `make`
- proto3 and gRPC
  - https://github.com/protocolbuffers/protobuf/releases
  - https://github.com/golang/protobuf
  - https://github.com/grpc/grpc
- [`sysl`](https://sysl.io/docs/installation) tool available on PATH
- [`arrai`](https://github.com/arr-ai/arrai) tool available on PATH


On OSX, after installing [go 1.12.9](https://golang.org/doc/install) run

    brew install golangci/tap/golangci-lint make curl protoc-gen-go grpc

### 1.1.3 Development

Test and lint everything with

    make

View all relevant make targets with

    make help

View test coverage in the browser with

    make coverage
