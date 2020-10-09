FROM golang:1.14.9-alpine3.12 AS stage

# install additional utilities
RUN apk add --no-cache git make

# install arrai
RUN git clone --depth 1 --branch v0.207.0 https://github.com/anz-bank/sysl.git && make -C sysl install
RUN git clone --depth 1 --branch v0.186.0 https://github.com/arr-ai/arrai.git && make -C arrai install

# install goimports
RUN go get golang.org/x/tools/cmd/goimports

FROM golang:1.14.9-alpine3.12
COPY --from=stage /go/bin/arrai /bin
COPY --from=stage /go/bin/sysl /bin
COPY --from=stage /go/bin/goimports /bin

# copy sysl-go to /sysl-go
COPY ./codegen/arrai/auto /sysl-go/codegen/arrai/auto

WORKDIR /work
ARG SYSLGO_VERSION=latest
ENV SYSLGO_VERSION=$SYSLGO_VERSION
ENTRYPOINT ["/sysl-go/codegen/arrai/auto/scripts/bootstrapper.sh"]
