# ─── Build Stage ─────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /agent cmd/agent/main.go

# ─── Runtime Stage (Distroless) ──────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /agent /agent

USER nonroot:nonroot

ENTRYPOINT ["/agent"]
