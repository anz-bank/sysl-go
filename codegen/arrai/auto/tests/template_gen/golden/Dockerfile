ARG REGISTRY

FROM ${REGISTRY}${REGISTRY:+/}golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -mod=vendor -o Template ./cmd/Template

FROM ${REGISTRY}${REGISTRY:+/}alpine
WORKDIR /app
COPY --from=builder /app/Template /bin/
CMD Template
