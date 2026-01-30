# Zapiki API Documentation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints (except `/health`) require authentication via an API key.

Include the API key in the request header:

```
X-API-Key: your_api_key_here
```

Or as a Bearer token:

```
Authorization: Bearer your_api_key_here
```

## Rate Limiting

Rate limits are applied per API key:
- **Free tier**: 10 requests per minute
- **Pro tier**: 1000 requests per minute

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Window`: Time window for rate limiting

When rate limited, you'll receive a `429 Too Many Requests` response.

## Endpoints

### Health Check

**GET /health**

Check the health of the API and its dependencies.

**Authentication**: Not required

**Response**:
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

**Status Codes**:
- `200`: All services healthy
- `503`: One or more services degraded

---

### List Proof Systems

**GET /api/v1/systems**

List all available proof systems and their capabilities.

**Response**:
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
        "features": ["fast-generation", "simple-commitment", "digital-signature"]
      }
    }
  ]
}
```

---

### Generate Proof

**POST /api/v1/proofs**

Generate a zero-knowledge proof.

**Request Body**:
```json
{
  "proof_system": "commitment",
  "data": {
    "type": "string",
    "value": "my secret data"
  },
  "public_inputs": {},
  "options": {
    "async": false,
    "template_id": "uuid",
    "circuit_id": "uuid"
  }
}
```

**Parameters**:
- `proof_system` (required): Type of proof system ("commitment", "groth16", "plonk", "stark")
- `data` (required): Input data for proof generation
  - `type`: Data type ("string", "json", "bytes")
  - `value`: The actual data value
- `public_inputs` (optional): Public inputs for the proof
- `options` (optional):
  - `async`: Force async processing (default: auto-detect based on proof system)
  - `template_id`: Use a pre-built template
  - `circuit_id`: Use a specific circuit

**Synchronous Response** (commitment proofs):
```json
{
  "proof_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "proof": { ... },
  "verification_key": { ... },
  "generation_time_ms": 45
}
```

**Asynchronous Response** (SNARK/STARK proofs):
```json
{
  "proof_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "Proof generation started. Poll /api/v1/proofs/550e8400-... for status."
}
```

**Status Codes**:
- `200`: Proof generated successfully (sync) or job created (async)
- `400`: Invalid request (missing parameters, unsupported proof system)
- `401`: Missing or invalid API key
- `429`: Rate limit exceeded
- `500`: Internal server error

---

### Get Proof

**GET /api/v1/proofs/{id}**

Retrieve a proof by its ID.

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "proof_system": "commitment",
  "status": "completed",
  "proof_data": { ... },
  "verification_key": { ... },
  "public_inputs": {},
  "generation_time_ms": 45,
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:30:01Z"
}
```

**Status Values**:
- `pending`: Proof generation queued
- `processing`: Proof being generated
- `completed`: Proof ready
- `failed`: Proof generation failed

**Status Codes**:
- `200`: Proof found
- `404`: Proof not found or unauthorized

---

### List Proofs

**GET /api/v1/proofs**

List all proofs for the authenticated user.

**Query Parameters**:
- `limit` (optional): Number of results to return (default: 20, max: 100)
- `offset` (optional): Number of results to skip (default: 0)

**Response**:
```json
{
  "proofs": [ ... ],
  "limit": 20,
  "offset": 0
}
```

---

### Delete Proof

**DELETE /api/v1/proofs/{id}**

Delete a proof by its ID.

**Response**:
```json
{
  "message": "Proof deleted successfully"
}
```

**Status Codes**:
- `200`: Proof deleted
- `404`: Proof not found or unauthorized
- `500`: Error deleting proof

---

### Verify Proof

**POST /api/v1/verify**

Verify a zero-knowledge proof.

**Request Body**:
```json
{
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
  },
  "public_inputs": {}
}
```

**Parameters**:
- `proof_system` (required): Type of proof system
- `proof` (required): The proof to verify
- `verification_key` (required): Verification key for the proof
- `public_inputs` (optional): Public inputs used in the proof

**Response**:
```json
{
  "valid": true,
  "verified_at": "2024-01-15T10:31:00Z"
}
```

Or if invalid:
```json
{
  "valid": false,
  "error_message": "Invalid signature",
  "verified_at": "2024-01-15T10:31:00Z"
}
```

**Status Codes**:
- `200`: Verification completed (check `valid` field for result)
- `400`: Invalid request
- `500`: Verification error

---

## Data Types

### Input Data

The `data` field in proof generation requests supports three types:

**String**:
```json
{
  "type": "string",
  "value": "my secret data"
}
```

**JSON**:
```json
{
  "type": "json",
  "value": {
    "age": 25,
    "country": "US"
  }
}
```

**Bytes (hex-encoded)**:
```json
{
  "type": "bytes",
  "value": "48656c6c6f"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

Common error codes:
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid API key
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service degraded

## Proof Systems

### Commitment

Simple hash-based commitment with Ed25519 signature.

**Capabilities**:
- No setup required
- Synchronous generation (< 100ms)
- Small proof size (~512 bytes)
- Suitable for data commitments and timestamping

**Proof Format**:
```json
{
  "commitment": "hex-encoded SHA256 hash",
  "nonce": "hex-encoded random nonce",
  "signature": "hex-encoded Ed25519 signature",
  "timestamp": "ISO 8601 timestamp",
  "public_key": "hex-encoded Ed25519 public key"
}
```

### Groth16 (Coming Soon)

zk-SNARK with trusted setup.

### PLONK (Coming Soon)

zk-SNARK with universal setup.

### STARK (Coming Soon)

Transparent proof system (no trusted setup).
