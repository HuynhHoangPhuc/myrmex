#!/bin/sh
set -e

# Wait for postgres to be ready
echo "Waiting for postgres..."
until pg_isready -h postgres -U myrmex; do
  sleep 2
done
echo "Postgres is ready."

# Run migrations for each service schema (separate version tables to avoid conflicts)
for svc in core module-hr module-subject module-timetable; do
  echo "Migrating $svc..."
  table="goose_db_version_$(echo $svc | tr '-' '_')"
  goose -dir /migrations/$svc -table "$table" postgres "$DATABASE_URL" up
done

# Seed demo data
echo "Seeding demo data..."
psql "$DATABASE_URL" -f /seed/seed.sql

echo "Migration and seed complete."
