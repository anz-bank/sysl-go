ARG DOCKER_BASE=golang:1.24.2-bookworm
FROM ${DOCKER_BASE}

ENV SYSL_VERSION=0.790.0
ENV ARRAI_VERSION=0.321.0

WORKDIR /temp-deps/sysl
RUN curl -LJO https://github.com/anz-bank/sysl/releases/download/v"$SYSL_VERSION"/sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && tar -xvf sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && mv sysl /bin/sysl

WORKDIR /temp-deps/arrai
RUN curl -LJO https://github.com/arr-ai/arrai/releases/download/v"$ARRAI_VERSION"/arrai_v"$ARRAI_VERSION"_linux-amd64.tar.gz && tar -xvf arrai_v"$ARRAI_VERSION"_linux-amd64.tar.gz && mv arrai /bin/arrai

WORKDIR /temp-deps/golangci-lint
RUN curl -LJO https://github.com/golangci/golangci-lint/releases/download/v2.1.5/golangci-lint-2.1.5-linux-amd64.tar.gz && tar -xvf golangci-lint-2.1.5-linux-amd64.tar.gz && mv golangci-lint-2.1.5-linux-amd64/golangci-lint /bin/golangci-lint

RUN go install golang.org/x/tools/cmd/goimports@latest

# Need this line to fix git commands not working inside the docker image (it is run in actions/checkout but the git config is not passed in to the image, and even if it was, the full path doesn't match)
RUN git config --global --add safe.directory /work

ENTRYPOINT [ "/usr/bin/make" ]
