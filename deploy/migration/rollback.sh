#!/bin/bash
# Wipe all production schemas. Use ONLY as last resort on go-live day.
# Usage: ./deploy/migration/rollback.sh <DATABASE_URL>
set -euo pipefail

DB_URL="${1:?Usage: rollback.sh <DATABASE_URL>}"

echo "⚠️  WARNING: This will DROP ALL DATA from production."
echo "   Notify HCMUS before proceeding: 'System temporarily offline for maintenance'"
read -rp "Type 'ROLLBACK' to confirm: " confirm
if [ "$confirm" != "ROLLBACK" ]; then
  echo "Aborted."
  exit 0
fi

echo "=== Dropping all schemas ==="
psql "$DB_URL" <<'SQL'
DROP SCHEMA IF EXISTS notification CASCADE;
DROP SCHEMA IF EXISTS analytics   CASCADE;
DROP SCHEMA IF EXISTS student     CASCADE;
DROP SCHEMA IF EXISTS timetable   CASCADE;
DROP SCHEMA IF EXISTS subject     CASCADE;
DROP SCHEMA IF EXISTS hr          CASCADE;
DROP SCHEMA IF EXISTS core        CASCADE;
SQL

echo "✅ Schemas dropped."
echo "   Next steps:"
echo "   1. Re-deploy previous known-good image revision via Cloud Run console"
echo "   2. Re-run migrations: gcloud run jobs execute myrmex-migrate --wait"
echo "   3. Investigate root cause before re-importing"
