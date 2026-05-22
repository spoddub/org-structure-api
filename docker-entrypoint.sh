#!/bin/sh
set -e

/app/bin/goose -dir /app/migrations postgres "$DATABASE_URL" up

exec /app/bin/org-structure-api