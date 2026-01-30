#!/bin/bash

# Railway Production Setup Script
# This script helps set up Zapiki on Railway

set -e

echo "🚀 Zapiki Railway Production Setup"
echo "===================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo -e "${RED}Railway CLI not found!${NC}"
    echo "Install it with: npm i -g @railway/cli"
    echo "Or visit: https://docs.railway.app/develop/cli"
    exit 1
fi

echo -e "${GREEN}✓ Railway CLI found${NC}"
echo ""

# Check if logged in
if ! railway whoami &> /dev/null; then
    echo -e "${YELLOW}Not logged in to Railway${NC}"
    echo "Logging in..."
    railway login
fi

echo -e "${GREEN}✓ Logged in to Railway${NC}"
echo ""

# Step 1: Create/Link Project
echo "Step 1: Project Setup"
echo "---------------------"
echo "Do you want to:"
echo "1) Create new Railway project"
echo "2) Link to existing project"
read -p "Enter choice (1 or 2): " project_choice

if [ "$project_choice" == "1" ]; then
    echo "Creating new Railway project..."
    railway init
elif [ "$project_choice" == "2" ]; then
    echo "Linking to existing project..."
    railway link
else
    echo -e "${RED}Invalid choice${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Project configured${NC}"
echo ""

# Step 2: Create services
echo "Step 2: Creating Services"
echo "-------------------------"
echo ""
echo -e "${YELLOW}Please create the following services in Railway dashboard:${NC}"
echo "1. PostgreSQL - Go to railway.app → Add Service → Database → PostgreSQL"
echo "2. Redis - Go to railway.app → Add Service → Database → Redis"
echo "3. API - Go to railway.app → Add Service → GitHub Repo (select zapiki)"
echo "4. Worker - Go to railway.app → Add Service → GitHub Repo (select zapiki)"
echo ""
read -p "Press Enter when you've created all services..."

echo -e "${GREEN}✓ Services created${NC}"
echo ""

# Step 3: Database Migration
echo "Step 3: Database Migration"
echo "--------------------------"
read -p "Do you want to run database migration now? (y/n): " run_migration

if [ "$run_migration" == "y" ]; then
    echo "Running database migration..."

    if railway run psql \$DATABASE_URL < deployments/docker/schema.sql; then
        echo -e "${GREEN}✓ Database migration completed${NC}"
    else
        echo -e "${RED}✗ Migration failed${NC}"
        echo "You can run it manually later with:"
        echo "  railway run psql \$DATABASE_URL < deployments/docker/schema.sql"
    fi
fi

echo ""

# Step 4: Initialize Templates
echo "Step 4: Initialize Templates"
echo "----------------------------"
read -p "Do you want to initialize templates? (y/n): " init_templates

if [ "$init_templates" == "y" ]; then
    echo "Initializing templates..."

    # First ensure circuits exist
    railway run psql \$DATABASE_URL <<EOF
-- Create test user if not exists
INSERT INTO users (email, name, tier)
VALUES ('admin@zapiki.io', 'Admin User', 'pro')
ON CONFLICT DO NOTHING;

-- Create circuits
INSERT INTO circuits (id, user_id, name, description, proof_system, circuit_definition, proving_key_url, verification_key_url, is_public, created_at, updated_at)
SELECT uuid_generate_v4(), u.id, 'Simple Circuit', 'Basic multiplication', 'groth16', '{"circuit_type":"simple"}', 'db:simple:pk', 'db:simple:vk', true, NOW(), NOW()
FROM users u WHERE u.email = 'admin@zapiki.io' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO circuits (id, user_id, name, description, proof_system, circuit_definition, proving_key_url, verification_key_url, is_public, created_at, updated_at)
SELECT uuid_generate_v4(), u.id, 'Age Verification Circuit', 'Prove age >= threshold', 'groth16', '{"circuit_type":"age_verification"}', 'db:age:pk', 'db:age:vk', true, NOW(), NOW()
FROM users u WHERE u.email = 'admin@zapiki.io' LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO circuits (id, user_id, name, description, proof_system, circuit_definition, proving_key_url, verification_key_url, is_public, created_at, updated_at)
SELECT uuid_generate_v4(), u.id, 'Range Proof Circuit', 'Prove value in range', 'groth16', '{"circuit_type":"range_proof"}', 'db:range:pk', 'db:range:vk', true, NOW(), NOW()
FROM users u WHERE u.email = 'admin@zapiki.io' LIMIT 1
ON CONFLICT DO NOTHING;
EOF

    # Then create templates
    if railway run psql \$DATABASE_URL < scripts/seed-templates.sql; then
        echo -e "${GREEN}✓ Templates initialized${NC}"
    else
        echo -e "${RED}✗ Template initialization failed${NC}"
    fi
fi

echo ""

# Step 5: Get API Key
echo "Step 5: API Key"
echo "---------------"
echo "Fetching API key..."

API_KEY=$(railway run psql \$DATABASE_URL -t -c "SELECT key FROM api_keys WHERE name = 'Test API Key' LIMIT 1;" | xargs)

if [ ! -z "$API_KEY" ]; then
    echo -e "${GREEN}✓ API Key retrieved${NC}"
    echo ""
    echo "Your API Key:"
    echo "  $API_KEY"
    echo ""
    echo "Save this key securely!"
else
    echo -e "${YELLOW}⚠ API key not found${NC}"
    echo "The migration may not have created it. Check the database."
fi

echo ""

# Step 6: Get Railway URL
echo "Step 6: Getting Service URL"
echo "---------------------------"
echo ""
echo -e "${YELLOW}Go to Railway dashboard and find your API service URL${NC}"
echo "It will look like: https://your-project.railway.app"
echo ""
read -p "Enter your Railway API URL: " RAILWAY_URL

echo ""
echo "Testing API..."
if curl -f -s "$RAILWAY_URL/health" > /dev/null; then
    echo -e "${GREEN}✓ API is healthy!${NC}"
    curl -s "$RAILWAY_URL/health" | python3 -m json.tool || curl -s "$RAILWAY_URL/health"
else
    echo -e "${RED}✗ API health check failed${NC}"
    echo "Check the logs with: railway logs"
fi

echo ""

# Summary
echo "======================================"
echo "✅ Setup Complete!"
echo "======================================"
echo ""
echo "Your Zapiki API is deployed at:"
echo "  $RAILWAY_URL"
echo ""
echo "Next steps:"
echo "1. Save your API key: $API_KEY"
echo "2. Test the API:"
echo "   curl $RAILWAY_URL/api/v1/systems -H \"X-API-Key: $API_KEY\""
echo ""
echo "3. Monitor services:"
echo "   railway logs         # View logs"
echo "   railway status       # Check status"
echo ""
echo "4. Configure custom domain (optional):"
echo "   Go to Railway dashboard → Settings → Domains"
echo ""
echo "Useful commands:"
echo "  railway logs --service api      # API logs"
echo "  railway logs --service worker   # Worker logs"
echo "  railway ps                      # List services"
echo "  railway open                    # Open dashboard"
echo ""
echo "Documentation: docs/PRODUCTION.md"
echo ""
echo "🚀 Your Zero-Knowledge Proof API is live!"
echo ""
