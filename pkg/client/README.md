# Zapiki Go SDK

Official Go client library for the Zapiki Zero-Knowledge Proof as a Service API.

## Installation

```bash
go get github.com/gabrielrondon/zapiki/pkg/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gabrielrondon/zapiki/pkg/client"
)

func main() {
    // Create client
    c := client.NewClient(
        "https://zapiki-production.up.railway.app",
        "your_api_key_here",
    )

    ctx := context.Background()

    // Generate a commitment proof
    req := &client.GenerateProofRequest{
        ProofSystem: client.ProofSystemCommitment,
        Data: client.DataInput{
            Type:  "string",
            Value: "Hello, Zero-Knowledge!",
        },
    }

    resp, err := c.GenerateProof(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Proof generated! ID: %s\n", resp.ProofID)
}
```

## Features

- ✅ Full API coverage (proofs, verification, templates, systems)
- ✅ Type-safe proof system selection
- ✅ Context support for cancellation and timeouts
- ✅ Automatic JSON encoding/decoding
- ✅ Clean error handling
- ✅ Batch operations support (coming soon)

## Usage Examples

### 1. Generate a Proof

```go
// Commitment proof (fast, synchronous)
req := &client.GenerateProofRequest{
    ProofSystem: client.ProofSystemCommitment,
    Data: client.DataInput{
        Type:  "string",
        Value: "my secret data",
    },
}

proof, err := c.GenerateProof(ctx, req)
```

### 2. Verify a Proof

```go
verifyReq := &client.VerifyRequest{
    ProofSystem:     client.ProofSystemCommitment,
    Proof:           proof.Proof,
    VerificationKey: proof.VerificationKey,
}

result, err := c.VerifyProof(ctx, verifyReq)
fmt.Println("Valid:", result.Valid)
```

### 3. Use a Template

```go
// List available templates
templates, err := c.ListTemplates(ctx)

// Generate proof using template
proof, err := c.GenerateFromTemplate(
    ctx,
    templateID,
    map[string]interface{}{
        "age":       25,
        "threshold": 18,
        "over_threshold": 1,
    },
)
```

### 4. List Proof Systems

```go
systems, err := c.ListSystems(ctx)
for _, sys := range systems {
    fmt.Printf("%s: %v\n", sys.Name, sys.Capabilities.Features)
}
```

### 5. Check API Health

```go
health, err := c.Health(ctx)
fmt.Printf("Status: %s\n", health["status"])
```

### 6. Async Proofs (SNARKs/STARKs)

```go
// Generate async proof
req := &client.GenerateProofRequest{
    ProofSystem: client.ProofSystemGroth16,
    Data: client.DataInput{
        Type:  "json",
        Value: map[string]interface{}{"a": 7, "b": 8, "c": 56},
    },
}

resp, err := c.GenerateProof(ctx, req)
fmt.Println("Job ID:", resp.JobID)
fmt.Println("Status:", resp.Status) // "pending"

// Poll for completion
for {
    time.Sleep(5 * time.Second)

    updated, err := c.GetProof(ctx, resp.ProofID)
    if err != nil {
        break
    }

    if updated.Status == "completed" {
        fmt.Println("Proof ready!")
        break
    }
}
```

### 7. Custom Timeout

```go
// Set custom timeout (default is 30s)
c := client.NewClient(baseURL, apiKey).
    WithTimeout(2 * time.Minute)
```

## Proof Systems

| System | Type | Speed | Setup Required |
|--------|------|-------|----------------|
| `ProofSystemCommitment` | Hash-based | ~50ms | No |
| `ProofSystemGroth16` | zk-SNARK | ~30s | Yes (trusted) |
| `ProofSystemPLONK` | zk-SNARK | ~35s | Yes (universal) |
| `ProofSystemSTARK` | Transparent | ~40s | No |

## Error Handling

```go
proof, err := c.GenerateProof(ctx, req)
if err != nil {
    // Check if it's an API error
    if strings.Contains(err.Error(), "status 401") {
        fmt.Println("Authentication failed")
    } else if strings.Contains(err.Error(), "status 429") {
        fmt.Println("Rate limit exceeded")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Best Practices

1. **Reuse clients**: Create one client and reuse it
   ```go
   // Good
   client := client.NewClient(url, key)
   defer client.Close()

   for _, data := range items {
       client.GenerateProof(ctx, data)
   }
   ```

2. **Use contexts**: Always pass context for cancellation
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()

   proof, err := c.GenerateProof(ctx, req)
   ```

3. **Handle async proofs**: SNARKs/STARKs are async
   ```go
   if resp.Status == "pending" {
       // Poll or use webhooks (future feature)
       time.Sleep(30 * time.Second)
       updated, _ := c.GetProof(ctx, resp.ProofID)
   }
   ```

4. **Check capabilities**: Different systems have different features
   ```go
   systems, _ := c.ListSystems(ctx)
   for _, sys := range systems {
       if !sys.Capabilities.RequiresTrustedSetup {
           // Use STARK or Commitment
       }
   }
   ```

## Complete Example

See [examples/main.go](examples/main.go) for a complete working example.

```bash
cd pkg/client/examples
go run main.go
```

## API Documentation

For full API documentation, see the [OpenAPI specification](../../openapi.yaml).

## Support

- **Issues**: https://github.com/gabrielrondon/zapiki/issues
- **API Docs**: https://zapiki-production.up.railway.app/docs
- **Main Repo**: https://github.com/gabrielrondon/zapiki

## License

MIT
