# Frontend build stage
FROM node:22-slim AS frontend

WORKDIR /build

COPY frontend/admin-ui/package.json frontend/admin-ui/package-lock.json* ./admin-ui/
RUN cd admin-ui && npm install --silent

COPY frontend/voice-ui/package.json frontend/voice-ui/package-lock.json* ./voice-ui/
RUN cd voice-ui && npm install --silent

COPY frontend/admin-ui/ ./admin-ui/
RUN cd admin-ui && npx vite build

COPY frontend/voice-ui/ ./voice-ui/
RUN cd voice-ui && npx vite build

# Download auxiliary models (VAD, mel-spectrogram, embeddings)
FROM golang:1.25-alpine AS models

WORKDIR /build
COPY scripts/download-model.go ./scripts/
RUN go run scripts/download-model.go

# Static ffmpeg binary (audio conversion: OGG→WAV, etc.)
FROM mwader/static-ffmpeg:7.1 AS ffmpeg

# Download ONNX Runtime
FROM debian:bookworm-slim AS onnx

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends wget ca-certificates && \
    ONNX_ARCH=$([ "$TARGETARCH" = "arm64" ] && echo "aarch64" || echo "x64") && \
    wget -q -O /tmp/onnxruntime.tgz \
      "https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-${ONNX_ARCH}-1.23.2.tgz" && \
    tar -xzf /tmp/onnxruntime.tgz -C /tmp && \
    mkdir -p /out && \
    cp /tmp/onnxruntime-linux-${ONNX_ARCH}-1.23.2/lib/libonnxruntime.so* /out/

# Go build stage (frontends + models embedded into binary)
FROM golang:1.25 AS builder

ARG TARGETARCH

WORKDIR /build

COPY server/go.mod server/go.sum ./
# Module download primes the layer cache. It cannot resolve a local replace
# directive (used while testing adk-utils-go changes before pushing them), so
# it is allowed to fail: the vendored build compiles from server/vendor and a
# genuinely missing dependency still fails the build step loudly.
RUN go mod download || echo "go mod download skipped (vendored build)"

COPY server/ ./
COPY --from=frontend /build/admin-ui/dist/ ./frontend/admin-ui/
COPY --from=frontend /build/voice-ui/dist/ ./frontend/voice-ui/
COPY models/wakeword/ ./models/wakeword/
COPY --from=models /build/models/auxiliary/ ./models/auxiliary/
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${TARGETARCH} go build -o magec .

# Runtime stage (distroless — no shell, minimal attack surface)
FROM gcr.io/distroless/cc-debian12

WORKDIR /app

COPY --from=onnx /out/libonnxruntime.so* /usr/lib/
COPY --from=ffmpeg /ffmpeg /usr/local/bin/ffmpeg
COPY --from=builder /build/magec .
COPY docker/compose/config.yaml ./config.yaml

EXPOSE 8080

ENTRYPOINT ["/app/magec"]
CMD ["--config", "/app/config.yaml"]
