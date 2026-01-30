#!/bin/bash

# Railway Database Migration Script
# Runs database schema migration and seeds templates

set -e

echo "🗄️  Railway Database Migration"
echo "=============================="
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

echo "📊 Running schema migration..."
echo "------------------------------"

# Run schema migration
if railway run --service Postgres psql postgresql://postgres:lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO@postgres.railway.internal:5432/railway < deployments/docker/schema.sql; then
    echo "✅ Schema migration completed"
else
    echo "❌ Schema migration failed"
    echo "You can run it manually with:"
    echo "  railway run --service Postgres psql <connection_url> < deployments/docker/schema.sql"
    exit 1
fi

echo ""
echo "🌱 Seeding templates..."
echo "----------------------"

# Seed templates
if railway run --service Postgres psql postgresql://postgres:lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO@postgres.railway.internal:5432/railway < scripts/seed-templates.sql; then
    echo "✅ Templates seeded successfully"
else
    echo "⚠️  Template seeding failed (may already exist)"
fi

echo ""
echo "🔑 Retrieving API Key..."
echo "-----------------------"

# Get API key
API_KEY=$(railway run --service Postgres psql postgresql://postgres:lFUKjRoZMsoZovfhlbmdBiCkMDsvkjZO@postgres.railway.internal:5432/railway -t -c "SELECT key FROM api_keys WHERE name = 'Test API Key' LIMIT 1;" | xargs 2>/dev/null || echo "")

if [ ! -z "$API_KEY" ]; then
    echo "✅ API Key retrieved"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Your API Key (save this securely!):"
    echo "  $API_KEY"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo "⚠️  API key not found"
    echo "You can create one manually in the database"
fi

echo ""
echo "✅ Migration completed!"
echo ""
