# ────────────────────────────────────────────
# Stage 1 — build a fully static binary
# ────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./
COPY handler/ handler/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ────────────────────────────────────────────
# Stage 2 — minimal scratch image
# No OS, no shell, no package manager
# ────────────────────────────────────────────
FROM scratch

WORKDIR /app

COPY --from=builder /app/server .
COPY static/ static/

ENV PORT=8922
EXPOSE 8922

ENTRYPOINT ["/app/server"]
