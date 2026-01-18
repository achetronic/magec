# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o magec .

# Runtime stage
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /build/magec .
COPY gui/ ./gui/

EXPOSE 8080

CMD ["./magec", "--config", "/app/config.yaml"]
