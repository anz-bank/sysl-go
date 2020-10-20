ARG REGISTRY

FROM ${REGISTRY}${REGISTRY:+/}golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -mod=vendor -o template ./cmd/template

FROM ${REGISTRY}${REGISTRY:+/}alpine
WORKDIR /app
COPY --from=builder /app/template /bin/
CMD template
