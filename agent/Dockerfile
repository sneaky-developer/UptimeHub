# ─── Build Stage ─────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /app

# git is already in bookworm
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /agent cmd/agent/main.go

# ─── Runtime Stage (Distroless) ──────────────────────────────────────
# Distroless static is still fine as it doesn't use package managers at runtime
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /agent /agent

USER nonroot:nonroot

ENTRYPOINT ["/agent"]
