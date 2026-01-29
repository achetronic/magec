# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o magec .

# Download pretrained models
FROM golang:1.24-alpine AS models

WORKDIR /build
COPY scripts/download-model.go ./scripts/
RUN go run scripts/download-model.go

# Runtime stage
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /build/magec .
COPY gui/ ./gui/
COPY --from=models /build/gui/pretrained/ ./gui/pretrained/

EXPOSE 8080

ENTRYPOINT ["./magec"]
CMD ["--config", "/app/config.yaml"]
