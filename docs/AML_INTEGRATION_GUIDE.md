# 🏦 AML/KYC Integration Guide

## Overview

Zapiki's AML/KYC templates allow you to implement privacy-preserving compliance checks using zero-knowledge proofs. Users prove compliance requirements (age, residency, income, sanctions) **without revealing sensitive personal data**.

**Benefits**:
- ✅ GDPR/CCPA compliant (data minimization)
- ✅ Reduce false positives
- ✅ Faster customer onboarding
- ✅ Lower compliance costs
- ✅ Increase customer trust

---

## Quick Start

### Prerequisites

1. **Get API Key**: [docs/FRONTEND_API_KEY.md](./FRONTEND_API_KEY.md)
2. **Base URL**: `https://zapiki-production.up.railway.app`

### Available Templates

| Template | Use Case | Public Input | Private Input |
|----------|----------|--------------|---------------|
| Age Verification | Prove age ≥ threshold | minimum_age, current_year | birth_year |
| Sanctions Check | Prove NOT on sanctions list | sanctions_list_root | user_id |
| Residency Proof | Prove jurisdiction | allowed_country_code | user_country_code, address_hash |
| Income Verification | Prove income ≥ threshold | minimum_income | actual_income, income_source_hash |

---

## 1. Age Verification

**Use Case**: Prove user is 18+ without revealing birthdate

### Request

```bash
curl -X POST https://zapiki-production.up.railway.app/api/v1/aml/age-verification \
  -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  -H "Content-Type: application/json" \
  -d '{
    "minimum_age": 18,
    "current_year": 2026,
    "birth_year": 1990,
    "nonce": "random_unique_string"
  }'
```

### Response

```json
{
  "proof_id": "proof_abc123",
  "status": "pending",
  "message": "Age verification proof generation started. Poll /api/v1/proofs/proof_abc123 for status."
}
```

### JavaScript Example

```javascript
const zapiki = {
  apiKey: 'test_zapiki_key_1230ab3c044056686e2552fb5a2648cd',
  baseUrl: 'https://zapiki-production.up.railway.app'
};

// Generate age proof
async function proveAge(birthYear) {
  const response = await fetch(`${zapiki.baseUrl}/api/v1/aml/age-verification`, {
    method: 'POST',
    headers: {
      'X-API-Key': zapiki.apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      minimum_age: 18,
      current_year: new Date().getFullYear(),
      birth_year: birthYear,
      nonce: crypto.randomUUID()
    })
  });

  const result = await response.json();
  return result.proof_id;
}

// Poll for proof completion
async function waitForProof(proofId) {
  while (true) {
    const response = await fetch(`${zapiki.baseUrl}/api/v1/proofs/${proofId}`, {
      headers: { 'X-API-Key': zapiki.apiKey }
    });

    const proof = await response.json();

    if (proof.status === 'completed') {
      return proof;
    } else if (proof.status === 'failed') {
      throw new Error('Proof generation failed');
    }

    // Wait 2 seconds before polling again
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
}

// Usage
async function verifyUserAge() {
  try {
    const proofId = await proveAge(1990);
    console.log('Proof generation started:', proofId);

    const proof = await waitForProof(proofId);
    console.log('Proof completed:', proof);

    // Send proof to verifier (bank, exchange, etc.)
    return proof;
  } catch (error) {
    console.error('Age verification failed:', error);
  }
}
```

### TypeScript Example

```typescript
interface AgeVerificationRequest {
  minimum_age: number;
  current_year: number;
  birth_year: number;
  nonce?: string;
}

interface ProofResponse {
  proof_id: string;
  status: 'pending' | 'completed' | 'failed';
  message: string;
  proof?: any;
}

class ZapikiAML {
  private apiKey: string;
  private baseUrl: string;

  constructor(apiKey: string, baseUrl: string = 'https://zapiki-production.up.railway.app') {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async verifyAge(birthYear: number, minimumAge: number = 18): Promise<ProofResponse> {
    const response = await fetch(`${this.baseUrl}/api/v1/aml/age-verification`, {
      method: 'POST',
      headers: {
        'X-API-Key': this.apiKey,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        minimum_age: minimumAge,
        current_year: new Date().getFullYear(),
        birth_year: birthYear,
        nonce: crypto.randomUUID()
      })
    });

    return await response.json();
  }

  async waitForProof(proofId: string, maxWaitMs: number = 60000): Promise<any> {
    const startTime = Date.now();

    while (Date.now() - startTime < maxWaitMs) {
      const response = await fetch(`${this.baseUrl}/api/v1/proofs/${proofId}`, {
        headers: { 'X-API-Key': this.apiKey }
      });

      const proof = await response.json();

      if (proof.status === 'completed') return proof;
      if (proof.status === 'failed') throw new Error('Proof failed');

      await new Promise(resolve => setTimeout(resolve, 2000));
    }

    throw new Error('Proof timeout');
  }
}

// Usage
const zapiki = new ZapikiAML('test_zapiki_key_1230ab3c044056686e2552fb5a2648cd');

async function main() {
  const result = await zapiki.verifyAge(1990, 18);
  const proof = await zapiki.waitForProof(result.proof_id);
  console.log('Proof:', proof);
}
```

---

## 2. Sanctions Check

**Use Case**: Prove user is NOT on OFAC/UN sanctions list

### Request

```bash
curl -X POST https://zapiki-production.up.railway.app/api/v1/aml/sanctions-check \
  -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  -H "Content-Type: application/json" \
  -d '{
    "sanctions_list_root": "0x1234abcd...",
    "current_timestamp": 1704067200,
    "user_id": "hashed_user_identifier"
  }'
```

### JavaScript Example

```javascript
async function checkSanctions(userId) {
  const response = await fetch(`${zapiki.baseUrl}/api/v1/aml/sanctions-check`, {
    method: 'POST',
    headers: {
      'X-API-Key': zapiki.apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      sanctions_list_root: await getSanctionsListRoot(), // From your backend
      current_timestamp: Math.floor(Date.now() / 1000),
      user_id: hashUserId(userId) // SHA-256 hash
    })
  });

  return await response.json();
}
```

---

## 3. Residency Proof

**Use Case**: Prove user resides in allowed jurisdiction without revealing address

### Request

```bash
curl -X POST https://zapiki-production.up.railway.app/api/v1/aml/residency-proof \
  -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  -H "Content-Type: application/json" \
  -d '{
    "allowed_country_code": 1,
    "current_timestamp": 1704067200,
    "user_country_code": 1,
    "address_hash": "sha256_of_full_address"
  }'
```

### JavaScript Example

```javascript
async function proveResidency(userCountry, userAddress) {
  const addressHash = await sha256(userAddress);

  const response = await fetch(`${zapiki.baseUrl}/api/v1/aml/residency-proof`, {
    method: 'POST',
    headers: {
      'X-API-Key': zapiki.apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      allowed_country_code: 1, // USA
      current_timestamp: Math.floor(Date.now() / 1000),
      user_country_code: userCountry,
      address_hash: addressHash
    })
  });

  return await response.json();
}

// SHA-256 helper
async function sha256(message) {
  const msgBuffer = new TextEncoder().encode(message);
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}
```

---

## 4. Income Verification

**Use Case**: Prove income ≥ threshold for lending/credit without revealing exact salary

### Request

```bash
curl -X POST https://zapiki-production.up.railway.app/api/v1/aml/income-verification \
  -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  -H "Content-Type: application/json" \
  -d '{
    "minimum_income": 50000,
    "current_timestamp": 1704067200,
    "actual_income": 75000,
    "income_source_hash": "sha256_of_w2_or_tax_return"
  }'
```

### JavaScript Example

```javascript
async function proveIncome(actualIncome, incomeDocument) {
  const sourceHash = await sha256(incomeDocument);

  const response = await fetch(`${zapiki.baseUrl}/api/v1/aml/income-verification`, {
    method: 'POST',
    headers: {
      'X-API-Key': zapiki.apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      minimum_income: 50000,
      current_timestamp: Math.floor(Date.now() / 1000),
      actual_income: actualIncome,
      income_source_hash: sourceHash
    })
  });

  return await response.json();
}
```

---

## Verification

After generating a proof, anyone can verify it without contacting the prover:

```javascript
async function verifyProof(proof, publicInputs) {
  const response = await fetch(`${zapiki.baseUrl}/api/v1/verify`, {
    method: 'POST',
    headers: {
      'X-API-Key': zapiki.apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      proof_system: 'groth16',
      proof: proof.proof,
      verification_key: proof.verification_key,
      public_inputs: publicInputs
    })
  });

  const result = await response.json();
  return result.valid; // true or false
}
```

---

## React Integration

```tsx
import { useState } from 'react';

function AgeVerificationForm() {
  const [birthYear, setBirthYear] = useState('');
  const [loading, setLoading] = useState(false);
  const [proof, setProof] = useState(null);

  const handleVerify = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const zapiki = new ZapikiAML(process.env.REACT_APP_ZAPIKI_API_KEY);
      const result = await zapiki.verifyAge(parseInt(birthYear), 18);
      const finalProof = await zapiki.waitForProof(result.proof_id);

      setProof(finalProof);
      alert('Age verified! Your birthdate was NOT revealed.');
    } catch (error) {
      alert('Verification failed: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleVerify}>
      <input
        type="number"
        placeholder="Birth Year (e.g., 1990)"
        value={birthYear}
        onChange={(e) => setBirthYear(e.target.value)}
        required
      />
      <button type="submit" disabled={loading}>
        {loading ? 'Generating Proof...' : 'Verify Age (18+)'}
      </button>
      {proof && <div>✅ Proof Generated: {proof.id}</div>}
    </form>
  );
}
```

---

## Security Best Practices

### 1. Never Expose Private Inputs
- **DO NOT** log `birth_year`, `actual_income`, `user_country_code`, etc.
- **DO NOT** store private inputs in your database
- **ONLY** send private inputs to Zapiki API (encrypted in transit via HTTPS)

### 2. Use Nonces
- Always generate unique `nonce` for each proof
- Prevents proof replay attacks
- Use `crypto.randomUUID()` or similar

### 3. Validate Public Inputs
- Verifier should check `minimum_age`, `current_year`, etc. match expected values
- Don't trust proofs with unexpected public inputs

### 4. Hash Sensitive Commitments
- Use SHA-256 to hash `address`, `income_source`, etc.
- Never send plaintext documents to the API

---

## Pricing

- **Free Tier**: 100 proofs/month
- **SMB**: $0.10 per proof
- **Enterprise**: Contact for volume discounts

---

## Support

- **Issues**: [GitHub Issues](https://github.com/gabrielrondon/zapiki/issues)
- **Email**: support@zapiki.io
- **Docs**: [Full Documentation](https://zapiki-docs.com)

---

## Next Steps

1. Get API key: [docs/FRONTEND_API_KEY.md](./FRONTEND_API_KEY.md)
2. Test in sandbox with example data
3. Integrate into your KYC/onboarding flow
4. Go to production

**Ready to reduce compliance costs by 28%? Start building with Zapiki today.**
