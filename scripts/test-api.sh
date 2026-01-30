#!/bin/bash

# Script to test the Zapiki API

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_URL=${API_URL:-http://localhost:8080}

echo -e "${YELLOW}Testing Zapiki API at $API_URL${NC}"
echo ""

# Get API key
echo "Getting API key..."
API_KEY=$(docker exec zapiki-postgres psql -U zapiki -d zapiki -t -c "SELECT key FROM api_keys WHERE name = 'Test API Key' LIMIT 1;" | xargs)

if [ -z "$API_KEY" ]; then
    echo -e "${RED}Error: Could not get API key${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Got API key${NC}"
echo ""

# Test 1: Health check
echo "Test 1: Health check (no auth required)"
RESPONSE=$(curl -s -w "\n%{http_code}" $API_URL/health)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Health check failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 2: List proof systems
echo "Test 2: List proof systems"
RESPONSE=$(curl -s -w "\n%{http_code}" -H "X-API-Key: $API_KEY" $API_URL/api/v1/systems)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Listed proof systems${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Failed to list proof systems (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi
echo ""

# Test 3: Generate commitment proof
echo "Test 3: Generate commitment proof"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $API_URL/api/v1/proofs \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "proof_system": "commitment",
        "data": {
            "type": "string",
            "value": "my secret data from test script"
        }
    }')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Generated proof${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"

    # Extract proof details for verification
    PROOF=$(echo "$BODY" | jq -r '.proof' 2>/dev/null)
    VK=$(echo "$BODY" | jq -r '.verification_key' 2>/dev/null)
    PROOF_ID=$(echo "$BODY" | jq -r '.proof_id' 2>/dev/null)
else
    echo -e "${RED}✗ Failed to generate proof (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
    PROOF=""
fi
echo ""

# Test 4: Verify proof
if [ ! -z "$PROOF" ] && [ "$PROOF" != "null" ]; then
    echo "Test 4: Verify proof"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $API_URL/api/v1/verify \
        -H "X-API-Key: $API_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"proof_system\": \"commitment\",
            \"proof\": $PROOF,
            \"verification_key\": $VK
        }")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" = "200" ]; then
        VALID=$(echo "$BODY" | jq -r '.valid' 2>/dev/null)
        if [ "$VALID" = "true" ]; then
            echo -e "${GREEN}✓ Proof verified successfully${NC}"
        else
            echo -e "${RED}✗ Proof verification failed${NC}"
        fi
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo -e "${RED}✗ Failed to verify proof (HTTP $HTTP_CODE)${NC}"
        echo "$BODY"
    fi
    echo ""
fi

# Test 5: Get proof by ID
if [ ! -z "$PROOF_ID" ] && [ "$PROOF_ID" != "null" ]; then
    echo "Test 5: Get proof by ID"
    RESPONSE=$(curl -s -w "\n%{http_code}" -H "X-API-Key: $API_KEY" $API_URL/api/v1/proofs/$PROOF_ID)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Retrieved proof${NC}"
        echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    else
        echo -e "${RED}✗ Failed to get proof (HTTP $HTTP_CODE)${NC}"
        echo "$BODY"
    fi
    echo ""
fi

echo -e "${YELLOW}Testing complete!${NC}"
