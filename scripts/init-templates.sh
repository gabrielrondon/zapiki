#!/bin/bash

# Script to initialize templates in the database

set -e

echo "Initializing Zapiki templates..."

# Check if docker-compose is running
if ! docker ps | grep -q zapiki-postgres; then
    echo "Error: PostgreSQL container is not running"
    echo "Start it with: make docker-up"
    exit 1
fi

echo "Step 1: Creating base circuits for templates..."

# Create circuits for templates
docker exec zapiki-postgres psql -U zapiki -d zapiki <<EOF
-- Create Simple Circuit for multiplication proofs
INSERT INTO circuits (
    id,
    user_id,
    name,
    description,
    proof_system,
    circuit_definition,
    proving_key_url,
    verification_key_url,
    is_public,
    created_at,
    updated_at
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'test@zapiki.io' LIMIT 1),
    'Simple Circuit',
    'Basic multiplication circuit: x * y = z',
    'groth16',
    '{"circuit_type":"simple"}',
    'db:simple:pk',
    'db:simple:vk',
    true,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- Create Age Verification Circuit
INSERT INTO circuits (
    id,
    user_id,
    name,
    description,
    proof_system,
    circuit_definition,
    proving_key_url,
    verification_key_url,
    is_public,
    created_at,
    updated_at
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'test@zapiki.io' LIMIT 1),
    'Age Verification Circuit',
    'Prove age >= threshold',
    'groth16',
    '{"circuit_type":"age_verification"}',
    'db:age:pk',
    'db:age:vk',
    true,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- Create Range Proof Circuit
INSERT INTO circuits (
    id,
    user_id,
    name,
    description,
    proof_system,
    circuit_definition,
    proving_key_url,
    verification_key_url,
    is_public,
    created_at,
    updated_at
) VALUES (
    uuid_generate_v4(),
    (SELECT id FROM users WHERE email = 'test@zapiki.io' LIMIT 1),
    'Range Proof Circuit',
    'Prove value is within range',
    'groth16',
    '{"circuit_type":"range_proof"}',
    'db:range:pk',
    'db:range:vk',
    true,
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

SELECT 'Created ' || COUNT(*) || ' circuits' FROM circuits WHERE is_public = true;
EOF

echo "Step 2: Creating templates..."

# Run the seed templates script
docker exec -i zapiki-postgres psql -U zapiki -d zapiki < scripts/seed-templates.sql

echo ""
echo "✓ Templates initialized successfully!"
echo ""
echo "Available templates:"
docker exec zapiki-postgres psql -U zapiki -d zapiki -c "SELECT name, category FROM templates WHERE is_active = true ORDER BY category, name;"
echo ""
echo "Test with:"
echo "  curl -H \"X-API-Key: \$API_KEY\" http://localhost:8080/api/v1/templates"
echo ""
