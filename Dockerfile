FROM anzbank/arrai:v0.98.0
ENV GOPATH /go
WORKDIR /usr
COPY . /sysl-go
CMD arrai run