#!/bin/bash
# Daily PostgreSQL backup — uploads to GCS bucket
# Runs via cron at 02:00 UTC (see startup script)
set -euo pipefail

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="/tmp/myrmex-staging-${TIMESTAMP}.sql.gz"
BUCKET="${GCS_BACKUP_BUCKET:-}"
COMPOSE_FILE="/opt/myrmex/deploy/docker/compose.yml"
STAGING_FILE="/opt/myrmex/deploy/staging/compose.staging.yml"
ENV_FILE="/opt/myrmex/deploy/staging/.env"

# Source env for GCS_BACKUP_BUCKET
if [ -f "$ENV_FILE" ]; then
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  BUCKET="${GCS_BACKUP_BUCKET:-}"
fi

# pg_dump via docker compose exec
docker compose -f "$COMPOSE_FILE" -f "$STAGING_FILE" \
  exec -T postgres pg_dump -U myrmex --clean --if-exists myrmex \
  | gzip > "$BACKUP_FILE"

echo "Backup created: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# Upload to GCS or keep locally
if [ -n "$BUCKET" ]; then
  gsutil cp "$BACKUP_FILE" "gs://${BUCKET}/backups/"
  echo "Uploaded to gs://${BUCKET}/backups/"
  rm -f "$BACKUP_FILE"
else
  LOCAL_DIR="/opt/myrmex/backups"
  mkdir -p "$LOCAL_DIR"
  mv "$BACKUP_FILE" "$LOCAL_DIR/"
  echo "Warning: GCS_BACKUP_BUCKET not set, backup saved to ${LOCAL_DIR}/"
  # Prune local backups older than 7 days
  find "$LOCAL_DIR" -name "myrmex-staging-*.sql.gz" -mtime +7 -delete
fi
echo "Backup complete: ${TIMESTAMP}"
