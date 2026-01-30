# Zapiki Quick Start Guide

This guide will get you up and running with Zapiki in 5 minutes.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- `curl` or similar HTTP client
- `jq` (optional, for pretty-printing JSON)

## Step 1: Start the Infrastructure

Start PostgreSQL and Redis using Docker Compose:

```bash
make docker-up
```

This will start:
- PostgreSQL on port 5432
- Redis on port 6379
- Minio (S3-compatible storage) on port 9000

Wait a few seconds for the services to fully start.

## Step 2: Start the API Server

In a new terminal, start the Zapiki API server:

```bash
make run
```

You should see:
```
Connected to PostgreSQL
Connected to Redis
Registered commitment proof system
Starting Zapiki API server on port 8080 (env: development)
```

## Step 3: Get Your API Key

In another terminal, retrieve your test API key:

```bash
./scripts/get-api-key.sh
```

This will output something like:
```
Your test API key is:

  test_zapiki_key_a1b2c3d4e5f6...
```

Copy this API key for use in the next steps.

## Step 4: Test the API

### Check System Health

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "services": {
    "api": "ok",
    "postgres": "ok",
    "redis": "ok"
  }
}
```

### List Available Proof Systems

```bash
export API_KEY="your_api_key_here"

curl -H "X-API-Key: $API_KEY" \
     http://localhost:8080/api/v1/systems | jq
```

Expected response:
```json
{
  "systems": [
    {
      "name": "commitment",
      "capabilities": {
        "supports_setup": false,
        "requires_trusted_setup": false,
        "supports_custom_circuits": false,
        "async_only": false,
        "typical_generation_time": 50,
        "max_proof_size": 512,
        "features": [
          "fast-generation",
          "simple-commitment",
          "digital-signature"
        ]
      }
    }
  ]
}
```

### Generate a Proof

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "string",
      "value": "my secret data"
    }
  }' | jq
```

Expected response:
```json
{
  "proof_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "proof": {
    "commitment": "a3f2...",
    "nonce": "b4e1...",
    "signature": "c5d3...",
    "timestamp": "2024-01-15T10:30:00Z",
    "public_key": "d6f4..."
  },
  "verification_key": {
    "public_key": "d6f4..."
  },
  "generation_time_ms": 12
}
```

Save the `proof` and `verification_key` for the next step.

### Verify a Proof

```bash
curl -X POST http://localhost:8080/api/v1/verify \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "proof": {
      "commitment": "a3f2...",
      "nonce": "b4e1...",
      "signature": "c5d3...",
      "timestamp": "2024-01-15T10:30:00Z",
      "public_key": "d6f4..."
    },
    "verification_key": {
      "public_key": "d6f4..."
    }
  }' | jq
```

Expected response:
```json
{
  "valid": true,
  "verified_at": "2024-01-15T10:31:00Z"
}
```

### Get Proof by ID

```bash
curl -H "X-API-Key: $API_KEY" \
     http://localhost:8080/api/v1/proofs/550e8400-e29b-41d4-a716-446655440000 | jq
```

### List All Your Proofs

```bash
curl -H "X-API-Key: $API_KEY" \
     http://localhost:8080/api/v1/proofs | jq
```

## Step 5: Run Automated Tests

Run the automated test script:

```bash
./scripts/test-api.sh
```

This will run through all the API endpoints and verify they work correctly.

## What's Next?

- Read the [API Documentation](API.md) for detailed endpoint information
- Explore the [Architecture Guide](ARCHITECTURE.md) to understand how Zapiki works
- Check the [Roadmap](../README.md#roadmap) for upcoming features

## Troubleshooting

### "Connection refused" errors

Make sure Docker services are running:
```bash
docker ps
```

You should see containers for `zapiki-postgres`, `zapiki-redis`, and `zapiki-minio`.

If not, restart them:
```bash
make docker-down
make docker-up
```

### "Invalid API key" errors

Make sure you're using the correct API key from `./scripts/get-api-key.sh`.

### Database connection errors

Reset the database:
```bash
make db-reset
```

Then restart the API server.

## Cleanup

To stop all services:

```bash
# Stop the API server (Ctrl+C in the terminal where it's running)

# Stop Docker services
make docker-down
```

To completely remove all data:

```bash
cd deployments/docker
docker-compose down -v
```
