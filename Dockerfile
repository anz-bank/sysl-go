FROM golang:1.16.3-buster AS stage

# requires git make curl
# but this base image has all of those tools already

ENV SYSL_VERSION=0.258.0
ENV ARRAI_VERSION=0.194.0

# install sysl. sysl's build process added a dependency on docker, which
# is an obstacle to building from source, so instead install the binary
WORKDIR /temp-deps/sysl
RUN curl -LJO https://github.com/anz-bank/sysl/releases/download/v"$SYSL_VERSION"/sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && tar -xvf sysl_"$SYSL_VERSION"_linux-amd64.tar.gz && mv sysl /bin/sysl
RUN chown root:root /bin/sysl

# install arrai
RUN git clone --depth 1 --branch v"$ARRAI_VERSION" https://github.com/arr-ai/arrai.git && make -C arrai install

# install goimports
RUN go get golang.org/x/tools/cmd/goimports

FROM golang:1.16.3-buster
COPY --from=stage /go/bin/arrai /bin
COPY --from=stage /bin/sysl /bin
COPY --from=stage /go/bin/goimports /bin

# copy sysl-go to /sysl-go
COPY ./codegen/arrai/auto /sysl-go/codegen/arrai/auto

WORKDIR /work
ARG SYSLGO_VERSION=latest
ENV SYSLGO_VERSION=$SYSLGO_VERSION
ENTRYPOINT ["/sysl-go/codegen/arrai/auto/scripts/bootstrapper.sh"]
