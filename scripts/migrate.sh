#!/usr/bin/env bash
set -e
DB_URL=${DB_URL:-"postgres://user:pass@localhost:5432/gamedb?sslmode=disable"}
migrate -path ./scripts/migrations -database "$DB_URL" $1