# syntax=docker/dockerfile:1.7

# ============================================================================
# builder — compiles the server binary + the golang-migrate CLI.
# ============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server

RUN go install -tags 'postgres' \
    github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# ============================================================================
# dev — hot-reload via Air. Mounts source code at runtime via docker-compose.
# ============================================================================
FROM golang:1.25-alpine AS dev

RUN apk add --no-cache git

RUN go install github.com/air-verse/air@latest && \
    go install -tags 'postgres' \
        github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

WORKDIR /app

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# ============================================================================
# prod — minimal Alpine runtime. Non-root user. ENTRYPOINT applies migrations
# before exec'ing the server.
# ============================================================================
FROM alpine:3.20 AS prod

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/server                 /app/server
COPY --from=builder /go/bin/migrate             /usr/local/bin/migrate
COPY --chown=app:app migrations                 /app/migrations
COPY --chown=app:app docker/entrypoint.sh       /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh && chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
