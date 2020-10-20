FROM golang:1.14-stretch

WORKDIR /temp-deps/sysl
RUN curl -LJO https://github.com/anz-bank/sysl/releases/download/v0.207.0/sysl_0.207.0_linux-amd64.tar.gz && tar -xvf sysl_0.207.0_linux-amd64.tar.gz && mv sysl /bin/sysl

WORKDIR /temp-deps/arrai
RUN curl -LJO https://github.com/arr-ai/arrai/releases/download/v0.189.0/arrai_v0.189.0_linux-amd64.tar.gz && tar -xvf arrai_v0.189.0_linux-amd64.tar.gz && mv arrai /bin/arrai

WORKDIR /temp-deps/golangci-lint
RUN curl -LJO https://github.com/golangci/golangci-lint/releases/download/v1.29.0/golangci-lint-1.29.0-linux-amd64.tar.gz && tar -xvf golangci-lint-1.29.0-linux-amd64.tar.gz && mv golangci-lint-1.29.0-linux-amd64/golangci-lint /bin/golangci-lint

RUN go get golang.org/x/tools/cmd/goimports

ENTRYPOINT [ "/usr/bin/make" ]
