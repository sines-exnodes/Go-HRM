#!/bin/sh
set -e

# Build DATABASE_URL from DB_* if not provided — mirrors the Makefile +
# internal/config behaviour so the same .env works for both local and docker.
if [ -z "${DATABASE_URL}" ]; then
    : "${DB_HOST:=postgres}"
    : "${DB_PORT:=5432}"
    : "${DB_USER:=postgres}"
    : "${DB_PASSWORD:=postgres}"
    : "${DB_NAME:=exnodes_hrm}"
    : "${DB_SSLMODE:=disable}"
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
fi

echo "entrypoint: waiting for postgres to accept connections..."
# Belt-and-suspenders alongside compose `depends_on.condition: service_healthy`.
# `migrate version` is a cheap connection probe.
attempts=0
until migrate -path /app/migrations -database "${DATABASE_URL}" version >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ "${attempts}" -gt 60 ]; then
        echo "entrypoint: postgres unreachable after 60s, giving up" >&2
        exit 1
    fi
    sleep 1
done

echo "entrypoint: applying migrations..."
migrate -path /app/migrations -database "${DATABASE_URL}" up

echo "entrypoint: starting server..."
exec /app/server
