#!/bin/sh
# Applies every schema folder's migrations in order. Each folder keeps its
# own version table, so the app and market schemas migrate independently.
set -eu
: "${DATABASE_URL:?DATABASE_URL is required}"
for schema in market app; do
  echo "migrating $schema"
  goose -dir "/db/$schema" -table "goose_${schema}_version" postgres "$DATABASE_URL" up
done
