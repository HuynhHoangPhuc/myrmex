#!/bin/bash
# Create super_admin account directly in the production database.
# Password must be a bcrypt hash — generate with: htpasswd -bnBC 10 "" "$PASSWORD" | tr -d ':\n'
# Usage: ADMIN_PASSWORD=xxx ./deploy/migration/bootstrap-admin.sh <API_URL> <DATABASE_URL>
set -euo pipefail

API_URL="${1:?Usage: bootstrap-admin.sh <API_URL> <DATABASE_URL>}"
DB_URL="${2:?Usage: bootstrap-admin.sh <API_URL> <DATABASE_URL>}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@hcmus.edu.vn}"
ADMIN_NAME="${ADMIN_NAME:-System Admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:?Set ADMIN_PASSWORD env var}"

echo "=== Bootstrapping super_admin: $ADMIN_EMAIL ==="

# Generate bcrypt hash locally (requires apache2-utils / httpd-tools)
if command -v htpasswd &>/dev/null; then
  HASH=$(htpasswd -bnBC 10 "" "$ADMIN_PASSWORD" | tr -d ':\n' | sed 's/$2y/$2a/')
else
  echo "ERROR: htpasswd not found. Install apache2-utils and re-run."
  exit 1
fi

psql "$DB_URL" <<SQL
INSERT INTO core.users (id, email, password_hash, full_name, role, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  '$ADMIN_EMAIL',
  '$HASH',
  '$ADMIN_NAME',
  'super_admin',
  now(),
  now()
)
ON CONFLICT (email) DO UPDATE SET
  role         = 'super_admin',
  password_hash = EXCLUDED.password_hash;
SQL

echo "Admin user upserted. Verifying login..."
TOKEN=$(curl -sf "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

if [ -z "$TOKEN" ]; then
  echo "ERROR: Login test failed — check password and API URL."
  exit 1
fi

echo "✅ Bootstrap complete. Admin token acquired."
echo "   Export for import scripts: export ADMIN_TOKEN='$TOKEN'"
