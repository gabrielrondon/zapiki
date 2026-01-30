#!/bin/bash

# Script to retrieve the test API key from the database

set -e

echo "Retrieving API key from database..."

# Check if docker-compose is running
if ! docker ps | grep -q zapiki-postgres; then
    echo "Error: PostgreSQL container is not running"
    echo "Start it with: make docker-up"
    exit 1
fi

# Get API key from database
API_KEY=$(docker exec zapiki-postgres psql -U zapiki -d zapiki -t -c "SELECT key FROM api_keys WHERE name = 'Test API Key' LIMIT 1;" | xargs)

if [ -z "$API_KEY" ]; then
    echo "Error: No API key found in database"
    echo "The database may not be initialized yet"
    exit 1
fi

echo ""
echo "Your test API key is:"
echo ""
echo "  $API_KEY"
echo ""
echo "Use it in requests like this:"
echo ""
echo "  curl -H \"X-API-Key: $API_KEY\" http://localhost:8080/api/v1/systems"
echo ""
