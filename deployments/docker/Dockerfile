# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/global-downloader \
    ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM python:3.12-slim AS runtime

# Install ffmpeg + yt-dlp runtime deps
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ffmpeg \
        curl \
        ca-certificates && \
    pip install --no-cache-dir yt-dlp && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy compiled Go binary from builder
COPY --from=builder /app/global-downloader .

# Create downloads directory
RUN mkdir -p /app/downloads

# Environment variable defaults (can be overridden at runtime)
ENV PORT=8080 \
    DOWNLOAD_DIR=/app/downloads \
    MAX_WORKERS=4 \
    MAX_QUEUE_SIZE=100 \
    YTDLP_PATH=yt-dlp \
    FFMPEG_PATH=ffmpeg \
    MAX_FILE_AGE_MINS=60

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./global-downloader"]
