#!/bin/sh
# Substitute CORE_SERVICE_URL env var into nginx config at runtime.
# Required for Cloud Run where the core service URL is injected as an env var.
set -e

: "${CORE_SERVICE_URL:?CORE_SERVICE_URL env var is required}"

envsubst '${CORE_SERVICE_URL}' \
    < /etc/nginx/templates/nginx-cloudrun.conf \
    > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
