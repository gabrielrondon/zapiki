# 🔑 API Key para Frontend - Zapiki

## ✅ API Key de Produção

Use a chave de teste por enquanto (já está ativa no banco):

```
test_zapiki_key_1230ab3c044056686e2552fb5a2648cd
```

**Rate Limit**: 100 requisições/minuto

## 🔄 Para criar uma chave dedicada ao frontend

Execute o script em `scripts/create_frontend_api_key.go`:

```bash
go run scripts/create_frontend_api_key.go
```

Isso criará uma chave com:
- Rate limit: 1000 req/min (10x maior)
- Usuário: Frontend Application
- Nome: Frontend Application Key

## 🧪 Testar a Chave

```bash
# Teste 1: Health Check (sem auth)
curl https://zapiki-production.up.railway.app/health

# Teste 2: Listar sistemas (com auth)
curl -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  https://zapiki-production.up.railway.app/api/v1/systems

# Teste 3: Gerar prova
curl -X POST https://zapiki-production.up.railway.app/api/v1/proofs \
  -H "X-API-Key: test_zapiki_key_1230ab3c044056686e2552fb5a2648cd" \
  -H "Content-Type: application/json" \
  -d '{"proof_system":"commitment","data":{"type":"string","value":"Frontend Test"}}'
```

## ⚙️ Uso no Frontend

### JavaScript/TypeScript
```javascript
const ZAPIKI_API_KEY = import.meta.env.VITE_ZAPIKI_API_KEY;
const ZAPIKI_BASE_URL = 'https://zapiki-production.up.railway.app';

// Exemplo de request
async function generateProof(data) {
  const response = await fetch(`${ZAPIKI_BASE_URL}/api/v1/proofs`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': ZAPIKI_API_KEY
    },
    body: JSON.stringify({
      proof_system: 'commitment',
      data: {
        type: 'string',
        value: data
      }
    })
  });

  return response.json();
}

// Listar sistemas disponíveis
async function listSystems() {
  const response = await fetch(`${ZAPIKI_BASE_URL}/api/v1/systems`, {
    headers: {
      'X-API-Key': ZAPIKI_API_KEY
    }
  });

  return response.json();
}

// Verificar prova
async function verifyProof(proofSystem, proof, verificationKey) {
  const response = await fetch(`${ZAPIKI_BASE_URL}/api/v1/verify`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': ZAPIKI_API_KEY
    },
    body: JSON.stringify({
      proof_system: proofSystem,
      proof: proof,
      verification_key: verificationKey
    })
  });

  return response.json();
}
```

### React (.env)
```env
VITE_ZAPIKI_API_KEY=test_zapiki_key_1230ab3c044056686e2552fb5a2648cd
VITE_ZAPIKI_BASE_URL=https://zapiki-production.up.railway.app
```

### Next.js (.env.local)
```env
NEXT_PUBLIC_ZAPIKI_API_KEY=test_zapiki_key_1230ab3c044056686e2552fb5a2648cd
NEXT_PUBLIC_ZAPIKI_BASE_URL=https://zapiki-production.up.railway.app
```

## ⚠️ Segurança

1. **Nunca commitar a API key no Git**
   - Usar variáveis de ambiente
   - Adicionar `.env` no `.gitignore`

2. **Frontend vs Backend**
   - Esta chave pode ser exposta no frontend (client-side)
   - Para operações sensíveis, considere criar um backend intermediário

3. **Rate Limiting**
   - Chave de teste: 100 req/min
   - Chave dedicada: 1000 req/min
   - Trate erros 429 (Too Many Requests)

## 📚 Documentação da API

Consulte o OpenAPI spec completo:
- Arquivo: `openapi.yaml` (v1.2.0)
- Online: https://zapiki-production.up.railway.app/docs (em breve)

## 🔗 Endpoints Principais

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/health` | GET | Health check (sem auth) |
| `/api/v1/systems` | GET | Listar sistemas de prova |
| `/api/v1/templates` | GET | Listar templates |
| `/api/v1/proofs` | POST | Gerar prova |
| `/api/v1/proofs/{id}` | GET | Buscar prova por ID |
| `/api/v1/verify` | POST | Verificar prova |
| `/api/v1/proofs/batch` | POST | Gerar múltiplas provas |

## 💡 Exemplos Práticos

### 1. Commitment Proof (Rápido - ~50ms)
```javascript
const proof = await generateProof({
  proof_system: 'commitment',
  data: {
    type: 'string',
    value: 'Meu documento secreto'
  }
});
// Retorna: { proof_id, status: 'completed', proof, verification_key }
```

### 2. Groth16 SNARK (Assíncrono - ~30s)
```javascript
const proof = await generateProof({
  proof_system: 'groth16',
  data: {
    type: 'json',
    value: { a: 5, b: 6, c: 30 }
  }
});
// Retorna: { proof_id, status: 'pending', job_id }
// Poll GET /api/v1/proofs/{proof_id} até status = 'completed'
```

### 3. Template (Pré-configurado)
```javascript
const templates = await fetch('/api/v1/templates', {
  headers: { 'X-API-Key': ZAPIKI_API_KEY }
});

const ageTemplate = templates.find(t => t.name.includes('Age'));

const proof = await fetch(`/api/v1/templates/${ageTemplate.id}/generate`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': ZAPIKI_API_KEY
  },
  body: JSON.stringify({
    inputs: { age: 25, threshold: 18, over_threshold: 1 }
  })
});
```

## 🎯 Rate Limits

Quando atingir o rate limit, você receberá:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

Status code: `429 Too Many Requests`

Implemente retry com backoff exponencial no frontend.
