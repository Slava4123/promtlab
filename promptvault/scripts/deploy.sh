#!/bin/bash
set -euo pipefail

echo "=== PromptVault Deploy ==="

# Check .env.prod exists
if [ ! -f .env.prod ]; then
    echo "ERROR: .env.prod not found. Copy .env.example and fill in secrets."
    exit 1
fi

# Check required vars
for var in DOMAIN CERTBOT_EMAIL DATABASE_HOST DATABASE_USER DATABASE_PASSWORD DATABASE_NAME JWT_SECRET SERVER_FRONTEND_URL AI_OPENROUTER_API_KEY; do
    if ! grep -qE "^${var}=.+" .env.prod; then
        echo "ERROR: ${var} is not set in .env.prod"
        exit 1
    fi
done

# Check for placeholder values
if grep -qE 'CHANGE_ME|your-domain|your@email' .env.prod; then
    echo "ERROR: .env.prod contains placeholder values (CHANGE_ME, your-domain, etc.)"
    echo "Replace all placeholder values before deploying."
    exit 1
fi

# Export env vars so docker compose build args can read DOMAIN etc.
set +u  # allow unset optional vars during source
set -a
source .env.prod
set +a
set -u  # re-enable strict unset checking

echo "Building and starting services..."
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans

echo "Waiting for services to be healthy..."
sleep 10

echo "Service status:"
docker compose -f docker-compose.prod.yml ps

# Health check verification
echo ""
echo "Checking health endpoint..."
if curl -sf --max-time 10 "https://${DOMAIN}/api/health" > /dev/null 2>&1; then
    echo "Health check PASSED: https://${DOMAIN}/api/health"
else
    echo "WARNING: Health check failed. Services may still be starting."
    echo "Manual check: curl https://${DOMAIN}/api/health"
fi

echo ""
echo "=== Deploy complete ==="
