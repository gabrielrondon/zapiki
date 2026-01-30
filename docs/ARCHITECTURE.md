# Zapiki Architecture

## Overview

Zapiki is a Zero-Knowledge Proof as a Service platform built with a layered architecture following clean architecture principles and dependency inversion.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Clients                               │
│  (Web Apps, Mobile Apps, Backend Services)              │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/REST
┌────────────────────▼────────────────────────────────────┐
│                 API Gateway                              │
│  - Authentication (API Keys)                             │
│  - Rate Limiting                                         │
│  - CORS                                                  │
│  - Logging                                               │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              HTTP Handlers                               │
│  - ProofHandler   - VerifyHandler                        │
│  - SystemHandler                                         │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              Service Layer                               │
│  - ProofService (orchestration)                          │
│  - VerifyService (verification)                          │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│        Proof System Abstraction                          │
│                                                          │
│  ┌────────────┐  ┌───────────┐  ┌──────────┐            │
│  │Commitment  │  │  Groth16  │  │  PLONK   │            │
│  │  Prover    │  │  Prover   │  │  Prover  │            │
│  └────────────┘  └───────────┘  └──────────┘            │
│                                                          │
│  All implement ProofSystem interface                     │
└────────────────────┬────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
┌────────▼─────────┐   ┌─────────▼────────┐
│  Storage Layer   │   │  Job Queue       │
│  - PostgreSQL    │   │  - Redis         │
│  - Redis Cache   │   │  - Workers       │
│  - S3/Minio      │   │                  │
└──────────────────┘   └──────────────────┘
```

## Layer Details

### 1. API Layer (`internal/api`)

**Responsibilities**:
- HTTP request handling
- Request validation
- Response formatting
- Middleware application

**Key Components**:

- **Handlers** (`handlers/`):
  - `proof_handler.go`: Proof CRUD operations
  - `verify_handler.go`: Proof verification
  - `system_handler.go`: System info and health

- **Middleware** (`middleware/`):
  - `auth.go`: API key authentication
  - `ratelimit.go`: Rate limiting via Redis
  - `logging.go`: Request/response logging
  - `cors.go`: CORS headers

- **Routes** (`routes/`):
  - `router.go`: Route definitions and middleware composition

### 2. Service Layer (`internal/service`)

**Responsibilities**:
- Business logic orchestration
- Proof system selection
- Sync vs async decision
- Database interaction coordination

**Key Components**:

- `proof_service.go`:
  - Determines if proof should be sync or async
  - Creates proof records
  - Delegates to appropriate proof system
  - Updates proof status

- `verify_service.go`:
  - Handles verification requests
  - Delegates to proof system verifiers

### 3. Proof System Layer (`internal/prover`)

**Responsibilities**:
- Proof generation algorithms
- Proof verification algorithms
- Cryptographic operations

**Key Components**:

- `interface.go`: Core `ProofSystem` interface
- `factory.go`: Proof system factory pattern
- `commitment/`: Commitment proof implementation
- `snark/gnark/`: Future SNARK implementations
- `stark/`: Future STARK implementations

**Interface Design**:
```go
type ProofSystem interface {
    Name() ProofSystemType
    Setup(ctx, circuit) (*SetupResult, error)
    Generate(ctx, *ProofRequest) (*ProofResponse, error)
    Verify(ctx, *VerifyRequest) (*VerifyResponse, error)
    Capabilities() Capabilities
}
```

### 4. Storage Layer (`internal/storage`)

**Responsibilities**:
- Data persistence
- Caching
- Queue management

**Key Components**:

- **PostgreSQL** (`postgres/`):
  - `postgres.go`: Connection pool management
  - `proof_repository.go`: Proof CRUD operations
  - `apikey_repository.go`: API key operations

- **Redis** (`redis/`):
  - `redis.go`: Redis client wrapper
  - Rate limiting implementation
  - Future: Job queue (Phase 2)

- **Object Storage** (`object/`):
  - Future: S3/Minio for large proofs and keys

### 5. Models (`internal/models`)

**Responsibilities**:
- Data structure definitions
- Type definitions
- Constants

**Key Models**:
- `User`: User account
- `APIKey`: API authentication
- `Proof`: Proof record
- `Circuit`: Circuit definition
- `Template`: Pre-built circuits
- `Job`: Async job tracking

## Data Flow

### Synchronous Proof Generation (Commitment)

```
1. Client sends POST /api/v1/proofs
2. AuthMiddleware validates API key
3. RateLimitMiddleware checks rate limit
4. ProofHandler parses request
5. ProofService:
   - Gets ProofSystem from Factory
   - Checks capabilities (async_only = false)
   - Creates Proof record (status = pending)
   - Calls ProofSystem.Generate()
   - Updates Proof record (status = completed)
6. Handler returns proof immediately
```

### Asynchronous Proof Generation (SNARK/STARK)

```
1. Client sends POST /api/v1/proofs
2. AuthMiddleware validates API key
3. RateLimitMiddleware checks rate limit
4. ProofHandler parses request
5. ProofService:
   - Gets ProofSystem from Factory
   - Checks capabilities (async_only = true)
   - Creates Proof record (status = pending)
   - Creates Job record
   - Queues job in Redis
6. Handler returns job_id immediately
7. Background Worker:
   - Picks up job from queue
   - Calls ProofSystem.Generate()
   - Updates Proof record with result
   - Updates Job status
8. Client polls GET /api/v1/proofs/{id} for status
```

## Database Schema

### Users
```sql
users (id, email, name, tier, created_at, updated_at)
```

### API Keys
```sql
api_keys (id, user_id, key, name, rate_limit, is_active,
          last_used_at, created_at, expires_at)
```

### Circuits
```sql
circuits (id, user_id, name, description, proof_system,
          circuit_definition, proving_key_url,
          verification_key_url, is_public, created_at, updated_at)
```

### Proofs
```sql
proofs (id, user_id, circuit_id, template_id, proof_system,
        status, input_data, proof_data, public_inputs,
        proof_url, error_message, generation_time_ms,
        created_at, completed_at)
```

### Jobs
```sql
jobs (id, user_id, proof_id, status, priority, retry_count,
      max_retries, error_message, created_at, started_at,
      completed_at)
```

### Templates
```sql
templates (id, name, description, category, proof_system,
           circuit_id, input_schema, example_inputs,
           documentation, is_active, created_at, updated_at)
```

## Design Patterns

### 1. Factory Pattern
The `prover.Factory` creates proof system instances:
```go
factory := prover.NewFactory()
factory.Register(commitmentProver)
system, _ := factory.Get("commitment")
```

### 2. Strategy Pattern
Each proof system implements the same `ProofSystem` interface, allowing runtime selection:
```go
type ProofSystem interface {
    Generate(...) (*ProofResponse, error)
    Verify(...) (*VerifyResponse, error)
}
```

### 3. Repository Pattern
Database access is abstracted behind repository interfaces:
```go
proofRepo.Create(ctx, proof)
proofRepo.GetByID(ctx, id)
```

### 4. Dependency Injection
Dependencies are injected via constructors:
```go
service := NewProofService(factory, proofRepo)
handler := NewProofHandler(service)
```

## Configuration

Configuration is environment-based using the `internal/config` package:

```go
cfg, _ := config.Load()
// Loads from environment variables
// Provides validation
// Generates connection strings
```

## Security

### Authentication
- API key-based authentication
- Keys stored with secure hashing
- Rate limiting per key
- Key expiration support

### Data Protection
- Sensitive input data not stored by default
- Proofs stored separately (future: encrypted S3)
- User data isolation

### Rate Limiting
- Sliding window algorithm
- Redis-backed
- Per-API-key limits
- Configurable by tier

## Scalability

### Horizontal Scaling
- Stateless API servers
- Load balancer ready
- Shared PostgreSQL and Redis

### Vertical Scaling
- Connection pooling
- Async processing for heavy workloads
- Background worker pool

### Caching Strategy
- Redis for:
  - Rate limiting counters
  - Verification key cache (future)
  - Job queue

## Monitoring & Observability

### Logging
- Structured logging (future)
- Request/response logging
- Error tracking

### Metrics (Future)
- Prometheus metrics
- Proof generation time histograms
- API request counters
- Error rates

### Tracing (Future)
- OpenTelemetry integration
- Distributed tracing
- Performance profiling

## Testing Strategy

### Unit Tests
- Each proof system independently tested
- Service layer logic tested
- Repository mocking

### Integration Tests
- End-to-end API tests
- Database integration tests
- Redis integration tests

### Performance Tests
- Proof generation benchmarks
- API load testing
- Concurrent request handling

## Future Enhancements

### Phase 2: Async Processing
- Worker pool implementation
- Job queue with retry logic
- Job status polling
- WebSocket updates

### Phase 3-6: Advanced Proof Systems
- gnark integration for SNARKs
- Winterfell for STARKs
- Circuit compilation
- Trusted setup management

### Phase 7: Production Hardening
- Comprehensive monitoring
- Security audit
- Performance optimization
- Load testing

### Phase 8: Advanced Features
- Go/JavaScript SDKs
- Circuit IDE
- Proof explorer UI
- Batch proof generation
