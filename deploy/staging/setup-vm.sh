#!/bin/bash
# One-time VM setup — run after Terraform creates the GCE instance
# Usage: ssh deploy@<VM_IP> 'bash -s' < deploy/staging/setup-vm.sh
set -euo pipefail

cd /opt/myrmex

# Clone repo (use HTTPS for public repo)
git clone https://github.com/phuchptty/myrmex.git . || echo "Repo already cloned"

# Create .env from template
if [ ! -f deploy/staging/.env ]; then
  cp deploy/staging/.env.example deploy/staging/.env
  echo "Created deploy/staging/.env — edit with actual values before deploying"
fi

# First deploy
docker compose -f deploy/docker/compose.yml \
  -f deploy/staging/compose.staging.yml \
  --env-file deploy/staging/.env \
  up -d

echo "Initial setup complete. Edit deploy/staging/.env then restart:"
echo "  docker compose -f deploy/docker/compose.yml -f deploy/staging/compose.staging.yml --env-file deploy/staging/.env up -d"
