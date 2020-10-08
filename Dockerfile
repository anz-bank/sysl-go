FROM golang:1.14.9-alpine3.12

# install additional utilities
RUN apk add --no-cache bash git make

# install sysl and arrai
WORKDIR /temp
RUN git clone --depth 1 --branch v0.207.0 https://github.com/anz-bank/sysl.git && make -C sysl install
RUN git clone --depth 1 --branch v0.186.0 https://github.com/arr-ai/arrai.git && make -C arrai install

# install goimports
RUN go get golang.org/x/tools/cmd/goimports

# copy sysl-go to /sysl-go
COPY . /sysl-go

# set the entrypoint
# note: this entrypoint is deprecated and will be removed in
# the future in favour of integration with codegen/arrai/auto
ENTRYPOINT [ "/sysl-go/scripts/arrai-docker.sh" ]
