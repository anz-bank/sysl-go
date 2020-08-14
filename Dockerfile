FROM anzbank/arrai:v0.129.0
ENV GOPATH /go
RUN apk add --no-cache bash
RUN go get golang.org/x/tools/cmd/goimports
WORKDIR /usr
COPY . /sysl-go
ENTRYPOINT [ ]
