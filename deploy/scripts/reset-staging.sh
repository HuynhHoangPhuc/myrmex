#!/bin/bash
# Wipe and re-seed staging database. Runs migrations via Cloud Run Job then re-seeds.
# Usage: ./deploy/scripts/reset-staging.sh <STAGING_DATABASE_URL> [GCP_PROJECT] [GCP_REGION]
# Example: ./deploy/scripts/reset-staging.sh "postgresql://..." myrmex-proj asia-southeast1
set -euo pipefail

DB_URL="${1:?Usage: reset-staging.sh <STAGING_DATABASE_URL> [GCP_PROJECT] [GCP_REGION]}"
GCP_PROJECT="${2:-}"
GCP_REGION="${3:-asia-southeast1}"

echo "=== WARNING: This will wipe ALL staging data ==="
read -rp "Type 'yes' to continue: " confirm
if [ "$confirm" != "yes" ]; then
  echo "Aborted."
  exit 0
fi

echo "=== Dropping staging schemas ==="
psql "$DB_URL" <<'SQL'
DROP SCHEMA IF EXISTS notification CASCADE;
DROP SCHEMA IF EXISTS analytics   CASCADE;
DROP SCHEMA IF EXISTS student     CASCADE;
DROP SCHEMA IF EXISTS timetable   CASCADE;
DROP SCHEMA IF EXISTS subject     CASCADE;
DROP SCHEMA IF EXISTS hr          CASCADE;
DROP SCHEMA IF EXISTS core        CASCADE;
SQL

echo "=== Re-running migrations ==="
if [ -n "$GCP_PROJECT" ]; then
  gcloud run jobs execute myrmex-migrate-staging \
    --region="$GCP_REGION" \
    --project="$GCP_PROJECT" \
    --wait
  echo "Migrations complete via Cloud Run Job."
else
  echo "SKIP: GCP_PROJECT not provided — run migrations manually before seeding."
fi

echo "=== Re-seeding staging data ==="
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/seed-staging.sh" "$DB_URL"

echo "=== Staging reset complete ==="
