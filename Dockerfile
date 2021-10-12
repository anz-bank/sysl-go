FROM golang:1.16.3-buster AS stage

# requires git make curl
# but this base image has all of those tools already

ENV SYSL_VERSION=0.459.0
ENV ARRAI_VERSION=0.277.0

ENV PROTOC_VERSION=3.17.3
ENV PROTOC_GEN_GO_VERSION=1.27.1
ENV PROTOC_GEN_GO_GRPC_VERSION=1.1

# install sysl. sysl's build process added a dependency on docker, which
# is an obstacle to building from source, so instead install the binary
WORKDIR /temp-deps/sysl
RUN curl -LJO https://github.com/anz-bank/sysl/releases/download/v"$SYSL_VERSION"/sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && tar -xvf sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && mv sysl /bin/sysl
RUN chown root:root /bin/sysl

# install arrai
RUN curl -LJO https://github.com/arr-ai/arrai/releases/download/v"$ARRAI_VERSION"/arrai_v"$ARRAI_VERSION"_linux-amd64.tar.gz && tar -xvf arrai_v"$ARRAI_VERSION"_linux-amd64.tar.gz && mv arrai /bin/arrai
RUN chown root:root /bin/arrai

# install goimports
RUN go get golang.org/x/tools/cmd/goimports

#install unzip
RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      unzip

#install protoc compiler and plugins
RUN curl -LJO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v$PROTOC_GEN_GO_VERSION
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$PROTOC_GEN_GO_GRPC_VERSION

FROM golang:1.16.3-buster
COPY --from=stage /bin/arrai /bin
COPY --from=stage /bin/sysl /bin
COPY --from=stage /go/bin/goimports /bin
COPY --from=stage /bin/protoc /bin
COPY --from=stage /go/bin/protoc-gen-go /bin
COPY --from=stage /go/bin/protoc-gen-go-grpc /bin

# copy sysl-go to /sysl-go
COPY ./codegen/arrai/auto /sysl-go/codegen/arrai/auto

WORKDIR /work
ARG SYSLGO_VERSION=latest
ENV SYSLGO_VERSION=$SYSLGO_VERSION
ENTRYPOINT ["/sysl-go/codegen/arrai/auto/scripts/bootstrapper.sh"]
