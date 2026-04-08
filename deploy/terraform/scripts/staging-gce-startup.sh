#!/bin/bash
# GCE staging VM startup script — installs Docker, Compose, and sets up backup cron
set -euo pipefail

# Skip if already provisioned
if [ -f /opt/myrmex/.provisioned ]; then
  echo "Already provisioned, skipping startup script"
  exit 0
fi

# Install Docker
curl -fsSL https://get.docker.com | sh

# Create deploy user (if not exists) and add to docker group
id -u deploy &>/dev/null || useradd -m -s /bin/bash deploy
usermod -aG docker deploy

# Install Docker Compose v2 plugin
apt-get update -qq && apt-get install -y -qq docker-compose-plugin

# Verify Compose version supports !override (v2.24.6+)
COMPOSE_VERSION=$(docker compose version --short)
echo "Docker Compose version: $COMPOSE_VERSION"

# Create app directory
mkdir -p /opt/myrmex
chown deploy:deploy /opt/myrmex

# Setup backup cron (daily at 02:00 UTC / 09:00 ICT)
cat > /etc/cron.d/myrmex-backup << 'CRON'
0 2 * * * deploy /opt/myrmex/deploy/staging/backup.sh >> /var/log/myrmex-backup.log 2>&1
CRON
chmod 644 /etc/cron.d/myrmex-backup

# Log rotation for backup cron
cat > /etc/logrotate.d/myrmex-backup << 'LOGROTATE'
/var/log/myrmex-backup.log {
    weekly
    rotate 4
    compress
    missingok
    notifempty
}
LOGROTATE

# Mark as provisioned
touch /opt/myrmex/.provisioned
echo "Startup script complete"
