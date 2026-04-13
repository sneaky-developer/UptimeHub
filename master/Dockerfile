# ─── Build Stage ─────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server cmd/server/main.go

# ─── Runtime Stage ───────────────────────────────────────────────────
FROM debian:bookworm-slim

# Skipping apt-get update to avoid network hangups
# Most Go binaries don't need additional runtime packages on debian-slim
WORKDIR /app

# Copy binary from builder
COPY --from=builder /server /usr/local/bin/server

# Expose port
EXPOSE 8080

# Run
CMD ["server"]
