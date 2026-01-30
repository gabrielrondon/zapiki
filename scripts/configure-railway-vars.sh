#!/bin/bash

# Railway Environment Variables Configuration Script
# This script configures all necessary environment variables for Zapiki services on Railway

set -e

echo "🔧 Configuring Railway Environment Variables"
echo "=============================================="
echo ""

# Check if logged in to Railway
if ! railway whoami &> /dev/null; then
    echo "❌ Not logged in to Railway. Please run: railway login"
    exit 1
fi

# Link to project if not already linked
if [ ! -f ".railway" ]; then
    echo "🔗 Linking to Railway project..."
    railway link --project a372d700-a757-465a-8564-a393e1cd3cff
fi

echo "📦 Configuring API Service (zapiki)..."
echo "-------------------------------------"

# Configure API service environment variables
railway variables --service zapiki set \
  API_PORT=8080 \
  ENV=production \
  POSTGRES_HOST=postgres.railway.internal \
  POSTGRES_PORT=5432 \
  POSTGRES_USER=postgres \
  POSTGRES_PASSWORD=lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO \
  POSTGRES_DB=railway \
  POSTGRES_SSLMODE=require \
  POSTGRES_MAX_CONNS=25 \
  POSTGRES_MIN_CONNS=5 \
  REDIS_HOST=redis.railway.internal \
  REDIS_PORT=6379 \
  REDIS_PASSWORD=JwFAzrMmFheGecokcsozvLzDqFZhOLmh \
  REDIS_DB=0 \
  ENABLE_COMMITMENT=true \
  ENABLE_GROTH16=true \
  ENABLE_PLONK=true \
  ENABLE_STARK=false \
  RATE_LIMIT_FREE_TIER=100 \
  RATE_LIMIT_PRO_TIER=10000 \
  SERVER_READ_TIMEOUT=30 \
  SERVER_WRITE_TIMEOUT=30 \
  SERVER_IDLE_TIMEOUT=120

echo "✅ API service configured"
echo ""

echo "👷 Configuring Worker Service (zakipi-worker)..."
echo "------------------------------------------------"

# Configure Worker service environment variables (same as API)
railway variables --service zakipi-worker set \
  ENV=production \
  POSTGRES_HOST=postgres.railway.internal \
  POSTGRES_PORT=5432 \
  POSTGRES_USER=postgres \
  POSTGRES_PASSWORD=lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO \
  POSTGRES_DB=railway \
  POSTGRES_SSLMODE=require \
  POSTGRES_MAX_CONNS=25 \
  POSTGRES_MIN_CONNS=5 \
  REDIS_HOST=redis.railway.internal \
  REDIS_PORT=6379 \
  REDIS_PASSWORD=JwFAzrMmFheGecokcsozvLzDqFZhOLmh \
  REDIS_DB=0 \
  ENABLE_COMMITMENT=true \
  ENABLE_GROTH16=true \
  ENABLE_PLONK=true \
  ENABLE_STARK=false \
  WORKER_CONCURRENCY=10

echo "✅ Worker service configured"
echo ""

echo "🎯 Setting Service Start Commands..."
echo "------------------------------------"

# Note: Railway deployment settings need to be set via dashboard or railway.toml
echo "ℹ️  Make sure start commands are set in Railway dashboard:"
echo "   - zapiki service: ./zapiki-api"
echo "   - zakipi-worker service: ./zapiki-worker"
echo ""

echo "✅ All environment variables configured!"
echo ""
echo "Next steps:"
echo "1. Verify settings in Railway dashboard"
echo "2. Trigger a new deployment if needed"
echo "3. Run database migration: ./scripts/railway-migrate.sh"
echo ""
