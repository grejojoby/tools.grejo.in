# ────────────────────────────────────────────
# Stage 1 — build a fully static binary
# ────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

ARG BUILD_NUMBER=0

WORKDIR /app

COPY go.mod ./
COPY main.go ./
COPY handler/ handler/

RUN echo "{\"build\":$BUILD_NUMBER}" > build.json

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ────────────────────────────────────────────
# Stage 2 — minimal scratch image
# No OS, no shell, no package manager
# ────────────────────────────────────────────
FROM scratch

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/build.json .
COPY static/ static/

ENV PORT=8922
EXPOSE 8922

ENTRYPOINT ["/app/server"]
