# Zapiki Usage Examples

This document provides practical examples of using the Zapiki API.

## Table of Contents
- [Basic Examples](#basic-examples)
- [Use Cases](#use-cases)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Basic Examples

### Example 1: Simple Data Commitment

Commit to a piece of data without revealing it:

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "string",
      "value": "I predict the stock will go up"
    }
  }'
```

**Response**:
```json
{
  "proof_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "completed",
  "proof": {
    "commitment": "7f3a8b9c2d1e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a",
    "nonce": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
    "signature": "9f8e7d6c5b4a39281726354a4b3c2d1e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6",
    "timestamp": "2024-01-30T10:00:00Z",
    "public_key": "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"
  },
  "verification_key": {
    "public_key": "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"
  },
  "generation_time_ms": 12
}
```

Later, you can reveal the original data and prove you committed to it at that time.

### Example 2: JSON Data Commitment

Commit to structured data:

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "json",
      "value": {
        "prediction": "Bitcoin will reach $50k",
        "date": "2024-12-31",
        "confidence": 0.85
      }
    }
  }'
```

### Example 3: Binary Data Commitment

Commit to binary data (hex-encoded):

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "bytes",
      "value": "48656c6c6f20576f726c64"
    }
  }'
```

### Example 4: Verifying a Proof

```bash
curl -X POST http://localhost:8080/api/v1/verify \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "proof": {
      "commitment": "7f3a8b9c2d1e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a",
      "nonce": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
      "signature": "9f8e7d6c5b4a39281726354a4b3c2d1e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6",
      "timestamp": "2024-01-30T10:00:00Z",
      "public_key": "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"
    },
    "verification_key": {
      "public_key": "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"
    }
  }'
```

**Response**:
```json
{
  "valid": true,
  "verified_at": "2024-01-30T10:05:00Z"
}
```

## Use Cases

### Use Case 1: Prediction Market Commitment

**Scenario**: You want to make a prediction and prove later that you made it before an event occurred.

**Step 1 - Make prediction (before event)**:
```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "json",
      "value": {
        "prediction": "Team A will win the championship",
        "event_date": "2024-06-15",
        "predictor": "alice@example.com"
      }
    }
  }' > prediction_proof.json
```

Save the proof and verification key.

**Step 2 - After event, prove your prediction**:

Show the original data and the proof. Anyone can verify:
```bash
cat prediction_proof.json | jq '.proof, .verification_key' > proof_to_share.json

# Others can verify
curl -X POST http://localhost:8080/api/v1/verify \
  -H "X-API-Key: THEIR_API_KEY" \
  -d @proof_to_share.json
```

### Use Case 2: Secure Timestamping

**Scenario**: Prove a document existed at a specific time.

```bash
# Hash your document first
DOC_HASH=$(shasum -a 256 important_document.pdf | cut -d' ' -f1)

# Create commitment
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"proof_system\": \"commitment\",
    \"data\": {
      \"type\": \"string\",
      \"value\": \"$DOC_HASH\"
    }
  }" > timestamp_proof.json
```

The proof includes a timestamp showing when the commitment was created.

### Use Case 3: Sealed Bid Auction

**Scenario**: Submit a sealed bid that can be revealed later.

**Bidder submits bid**:
```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: BIDDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "json",
      "value": {
        "auction_id": "auction_123",
        "bidder_id": "bidder_456",
        "bid_amount": 10000,
        "currency": "USD"
      }
    }
  }' > sealed_bid.json
```

**After bidding closes, reveal bids**:

Bidders reveal their original bid data. The system verifies:
1. The bid was committed before the deadline (check timestamp)
2. The revealed data matches the commitment (check signature)

### Use Case 4: Data Provenance

**Scenario**: Prove data came from a specific source at a specific time.

```bash
# Data provider commits to data
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: PROVIDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "json",
      "value": {
        "data_source": "weather_station_42",
        "temperature": 72.5,
        "humidity": 65,
        "timestamp": "2024-01-30T10:00:00Z",
        "location": "37.7749,-122.4194"
      }
    }
  }' > data_provenance.json
```

Anyone can verify the data came from that source at that time.

## Error Handling

### Invalid API Key

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: invalid_key" \
  -d '{...}'
```

**Response** (401):
```json
{
  "error": "Invalid API key"
}
```

### Missing Required Field

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "proof_system": "commitment"
  }'
```

**Response** (400):
```json
{
  "error": "data is required"
}
```

### Rate Limit Exceeded

```bash
# After making too many requests
curl -H "X-API-Key: YOUR_API_KEY" \
     http://localhost:8080/api/v1/systems
```

**Response** (429):
```json
{
  "error": "Rate limit exceeded"
}
```

**Headers**:
```
X-RateLimit-Limit: 10
X-RateLimit-Window: 1m0s
```

### Unsupported Proof System

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "proof_system": "nonexistent",
    "data": {...}
  }'
```

**Response** (500):
```json
{
  "error": "unsupported proof system: proof system nonexistent not found"
}
```

## Best Practices

### 1. Store Proofs Safely

Always save the complete proof response:

```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{...}' > proof_$(date +%s).json
```

### 2. Keep Original Data

Store the original data separately so you can reveal it later:

```bash
# Save original data
echo '{"secret": "data"}' > original_data.json

# Create commitment
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -d "{\"proof_system\":\"commitment\",\"data\":{\"type\":\"json\",\"value\":$(cat original_data.json)}}" \
  > proof.json
```

### 3. Use Environment Variables for API Keys

```bash
export ZAPIKI_API_KEY="your_api_key_here"

curl -H "X-API-Key: $ZAPIKI_API_KEY" \
     http://localhost:8080/api/v1/systems
```

### 4. Check Proof Status for Async Operations

For future SNARK/STARK proofs that take longer:

```bash
# Submit proof
PROOF_ID=$(curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: $ZAPIKI_API_KEY" \
  -d '{...}' | jq -r '.proof_id')

# Poll for completion
while true; do
  STATUS=$(curl -s -H "X-API-Key: $ZAPIKI_API_KEY" \
    http://localhost:8080/api/v1/proofs/$PROOF_ID | jq -r '.status')

  if [ "$STATUS" = "completed" ]; then
    echo "Proof ready!"
    break
  fi

  echo "Status: $STATUS, waiting..."
  sleep 2
done
```

### 5. Batch Requests with jq

Generate multiple proofs:

```bash
cat data_items.json | jq -c '.[]' | while read item; do
  curl -X POST http://localhost:8080/api/v1/proofs \
    -H "X-API-Key: $ZAPIKI_API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"proof_system\":\"commitment\",\"data\":{\"type\":\"json\",\"value\":$item}}"
  sleep 0.1  # Respect rate limits
done
```

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

async function generateProof(data) {
  const response = await axios.post('http://localhost:8080/api/v1/proofs', {
    proof_system: 'commitment',
    data: {
      type: 'json',
      value: data
    }
  }, {
    headers: {
      'X-API-Key': process.env.ZAPIKI_API_KEY,
      'Content-Type': 'application/json'
    }
  });

  return response.data;
}

async function verifyProof(proof, verificationKey) {
  const response = await axios.post('http://localhost:8080/api/v1/verify', {
    proof_system: 'commitment',
    proof: proof,
    verification_key: verificationKey
  }, {
    headers: {
      'X-API-Key': process.env.ZAPIKI_API_KEY,
      'Content-Type': 'application/json'
    }
  });

  return response.data.valid;
}

// Usage
(async () => {
  const result = await generateProof({ message: 'Hello ZK!' });
  console.log('Proof ID:', result.proof_id);

  const isValid = await verifyProof(result.proof, result.verification_key);
  console.log('Verified:', isValid);
})();
```

### Python

```python
import requests
import os

ZAPIKI_API_KEY = os.getenv('ZAPIKI_API_KEY')
BASE_URL = 'http://localhost:8080/api/v1'

def generate_proof(data):
    response = requests.post(f'{BASE_URL}/proofs',
        headers={
            'X-API-Key': ZAPIKI_API_KEY,
            'Content-Type': 'application/json'
        },
        json={
            'proof_system': 'commitment',
            'data': {
                'type': 'json',
                'value': data
            }
        })
    return response.json()

def verify_proof(proof, verification_key):
    response = requests.post(f'{BASE_URL}/verify',
        headers={
            'X-API-Key': ZAPIKI_API_KEY,
            'Content-Type': 'application/json'
        },
        json={
            'proof_system': 'commitment',
            'proof': proof,
            'verification_key': verification_key
        })
    return response.json()['valid']

# Usage
result = generate_proof({'message': 'Hello ZK!'})
print(f"Proof ID: {result['proof_id']}")

is_valid = verify_proof(result['proof'], result['verification_key'])
print(f"Verified: {is_valid}")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type ProofRequest struct {
    ProofSystem string    `json:"proof_system"`
    Data        InputData `json:"data"`
}

type InputData struct {
    Type  string      `json:"type"`
    Value interface{} `json:"value"`
}

func generateProof(data interface{}) (map[string]interface{}, error) {
    req := ProofRequest{
        ProofSystem: "commitment",
        Data: InputData{
            Type:  "json",
            Value: data,
        },
    }

    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/proofs", bytes.NewBuffer(body))
    httpReq.Header.Set("X-API-Key", os.Getenv("ZAPIKI_API_KEY"))
    httpReq.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}

func main() {
    result, _ := generateProof(map[string]string{"message": "Hello ZK!"})
    fmt.Printf("Proof ID: %s\n", result["proof_id"])
}
```

## Advanced Examples

### Chaining Commitments

Create a chain of commitments where each references the previous:

```bash
# First commitment
PROOF1=$(curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: $ZAPIKI_API_KEY" \
  -d '{"proof_system":"commitment","data":{"type":"string","value":"Block 1"}}')

PROOF1_ID=$(echo $PROOF1 | jq -r '.proof_id')

# Second commitment references first
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: $ZAPIKI_API_KEY" \
  -d "{\"proof_system\":\"commitment\",\"data\":{\"type\":\"json\",\"value\":{\"block\":2,\"prev\":\"$PROOF1_ID\"}}}"
```

This creates a simple blockchain-like structure.

---

For more examples and use cases, see the [API Documentation](API.md).
