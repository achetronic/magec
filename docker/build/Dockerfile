# Build stage
FROM golang:1.25 AS builder

WORKDIR /build

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o magec .

# Download auxiliary models (VAD, mel-spectrogram, embeddings)
FROM golang:1.24-alpine AS models

WORKDIR /build
COPY scripts/download-model.go ./scripts/
RUN go run scripts/download-model.go

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install CA certificates for TLS and ffmpeg for audio conversion
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ffmpeg && rm -rf /var/lib/apt/lists/*

# Install ONNX Runtime
ADD https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz /tmp/onnxruntime.tgz
RUN tar -xzf /tmp/onnxruntime.tgz -C /tmp && \
    cp /tmp/onnxruntime-linux-x64-1.23.2/lib/libonnxruntime.so* /usr/lib/ && \
    rm -rf /tmp/onnxruntime* && \
    ldconfig

COPY --from=builder /build/magec .
COPY docker/build/entrypoint.sh .
COPY voice-ui/ ./voice-ui/
COPY admin-ui/dist/ ./admin-ui/dist/
COPY data/seeds/ ./data/seeds/

# Models (wake word + auxiliary for server-side detection)
COPY models/wakeword/ ./models/wakeword/
COPY --from=models /build/models/auxiliary/ ./models/auxiliary/

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
CMD ["--config", "/app/config.yaml"]
