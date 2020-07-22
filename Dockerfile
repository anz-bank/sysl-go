FROM arrai
ENV GOPATH /go
WORKDIR /usr
COPY . /sysl-go
CMD arrai run