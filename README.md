# Zapiki - Zero-Knowledge Proof as a Service

**Vision**: "Stripe for Zero-Knowledge Proofs" - making ZK accessible to every developer and company.

Zapiki is a universal ZK-as-a-Service platform that abstracts cryptographic complexity behind a simple REST API. Users submit data and receive proofs - no cryptography expertise required.

## Features

- **Simple REST API** - Generate and verify proofs with HTTP requests
- **Multiple Proof Systems** - Support for commitments, zk-SNARKs, and zk-STARKs
- **Async Processing** - Handle long-running proof generation with job queue
- **Rate Limiting** - Built-in rate limiting per API key
- **Scalable** - PostgreSQL, Redis, and S3 for production workloads

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, but recommended)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/gabrielrondon/zapiki.git
cd zapiki
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Start the infrastructure services:
```bash
make docker-up
```

4. Run the API server:
```bash
make run
```

The API will be available at `http://localhost:8080`.

### Testing the API

1. **Check system health**:
```bash
curl http://localhost:8080/health
```

2. **List available proof systems**:
```bash
curl -H "X-API-Key: test_zapiki_key_<your-key>" \
     http://localhost:8080/api/v1/systems
```

3. **Generate a commitment proof**:
```bash
curl -X POST http://localhost:8080/api/v1/proofs \
  -H "X-API-Key: test_zapiki_key_<your-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "string",
      "value": "my secret data"
    }
  }'
```

4. **Verify a proof**:
```bash
curl -X POST http://localhost:8080/api/v1/verify \
  -H "X-API-Key: test_zapiki_key_<your-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "proof": {...},
    "verification_key": {...}
  }'
```

## API Documentation

### Endpoints

#### Health Check
- `GET /health` - Check service health (no auth required)

#### System Information
- `GET /api/v1/systems` - List available proof systems and capabilities

#### Proof Generation
- `POST /api/v1/proofs` - Generate a proof
- `GET /api/v1/proofs` - List user's proofs
- `GET /api/v1/proofs/{id}` - Get specific proof
- `DELETE /api/v1/proofs/{id}` - Delete a proof

#### Verification
- `POST /api/v1/verify` - Verify a proof

### Authentication

All API endpoints (except `/health`) require an API key. Include it in the request header:

```
X-API-Key: your_api_key_here
```

Or as a Bearer token:

```
Authorization: Bearer your_api_key_here
```

### Supported Proof Systems

#### Commitment (Phase 1 - Available Now)
- **Type**: Simple hash-based commitment with Ed25519 signature
- **Speed**: < 100ms
- **Use Cases**: Data commitments, timestamping
- **Example**:
```json
{
  "proof_system": "commitment",
  "data": {
    "type": "string",
    "value": "secret_data"
  }
}
```

#### Groth16 (Phase 3 - Coming Soon)
- **Type**: zk-SNARK with trusted setup
- **Speed**: ~10-60s
- **Use Cases**: Privacy-preserving proofs, identity verification

#### PLONK (Phase 5 - Coming Soon)
- **Type**: zk-SNARK with universal setup
- **Speed**: ~15-90s
- **Use Cases**: Flexible circuit proofs

#### STARK (Phase 6 - Coming Soon)
- **Type**: Transparent proof (no trusted setup)
- **Speed**: ~30-120s
- **Use Cases**: Post-quantum secure proofs

## Development

### Project Structure

```
zapiki/
├── cmd/api/          # API server entry point
├── internal/
│   ├── api/          # HTTP handlers, middleware, routes
│   ├── config/       # Configuration management
│   ├── models/       # Data models
│   ├── prover/       # Proof system implementations
│   ├── service/      # Business logic
│   └── storage/      # Database and cache layers
├── deployments/      # Docker configs
└── scripts/          # Helper scripts
```

### Available Make Commands

```bash
make help          # Show available commands
make build         # Build the binary
make run           # Run the server
make test          # Run tests
make docker-up     # Start Docker services
make docker-down   # Stop Docker services
make db-reset      # Reset database
make dev           # Start development environment
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

## Configuration

Configuration is done via environment variables. See `.env.example` for all available options.

Key settings:
- `API_PORT` - API server port (default: 8080)
- `POSTGRES_*` - PostgreSQL connection settings
- `REDIS_*` - Redis connection settings
- `ENABLE_*` - Enable/disable proof systems
- `RATE_LIMIT_*` - Rate limiting configuration

## Database

The database schema is automatically initialized when running `docker-compose up`.

To manually reset the database:
```bash
make db-reset
```

## Roadmap

- [x] **Phase 1**: Foundation & Commitment proofs
- [ ] **Phase 2**: Async job processing
- [ ] **Phase 3**: Groth16 SNARK integration
- [ ] **Phase 4**: Template system
- [ ] **Phase 5**: PLONK support
- [ ] **Phase 6**: STARK integration
- [ ] **Phase 7**: Production hardening
- [ ] **Phase 8**: Advanced features (SDKs, UI)

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - see LICENSE file for details.

## Support

For questions or issues, please open a GitHub issue or contact support@zapiki.io.

---

Built with ❤️ using Go, PostgreSQL, Redis, and cryptography.
