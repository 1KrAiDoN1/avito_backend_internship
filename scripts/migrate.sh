#!/bin/bash
set -e

echo "Waiting for PostgreSQL to start..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  sleep 1
done

echo "PostgreSQL is ready!"

echo "Running migrations..."

for migration in /migrations/*.up.sql; do
  if [ -f "$migration" ]; then
    echo "Applying migration: $migration"
    if ! PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -f "$migration"; then
      echo "ERROR: Failed to apply migration: $migration"
      exit 1
    fi
    echo "✓ Successfully applied: $(basename $migration)"
  fi
done

echo "✓ All migrations completed successfully!"


