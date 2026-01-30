# Load Testing Zapiki

This directory contains load testing scripts for Zapiki using k6.

## Prerequisites

Install k6:
```bash
# macOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows (via Chocolatey)
choco install k6
```

## Running Load Tests

### Local Testing

```bash
# Test local instance
k6 run scripts/load-test.js

# Test with custom base URL
k6 run --env BASE_URL=http://localhost:8080 scripts/load-test.js

# Test with custom API key
k6 run --env API_KEY=your_api_key_here scripts/load-test.js
```

### Production Testing

```bash
# Test production API
k6 run --env BASE_URL=https://zapiki-production.up.railway.app \
       --env API_KEY=your_prod_api_key \
       scripts/load-test.js
```

### Custom Load Profiles

```bash
# Quick smoke test (1 VU for 30s)
k6 run --vus 1 --duration 30s scripts/load-test.js

# Stress test (200 VUs for 5 minutes)
k6 run --vus 200 --duration 5m scripts/load-test.js

# Spike test
k6 run --stage 10s:10 \
       --stage 20s:100 \
       --stage 10s:10 \
       scripts/load-test.js
```

## Test Scenarios

The load test script includes:

1. **Health Check** - Verify API is responding
2. **List Systems** - Test /api/v1/systems endpoint
3. **Generate Commitment Proof** - Fast synchronous proofs
4. **List Templates** - Test template listing
5. **Verify Proof** - End-to-end proof verification

## Performance Thresholds

Current thresholds:
- 95th percentile response time: < 5 seconds
- Error rate: < 10%
- Failed requests: < 5%

## Metrics Collected

- **HTTP metrics**: request duration, success rate
- **Proof generation time**: time to generate proofs
- **Verification time**: time to verify proofs
- **Error rate**: percentage of failed requests

## Results

Results are saved to `load-test-results.json` after each run.

## Analyzing Results

```bash
# View JSON results
cat load-test-results.json | jq .

# View specific metrics
cat load-test-results.json | jq '.metrics.http_req_duration'
```

## CI/CD Integration

Load tests can be run in CI:

```yaml
- name: Run load tests
  run: k6 run --quiet scripts/load-test.js
```

## Tips

1. **Start small**: Begin with low VU counts and gradually increase
2. **Monitor resources**: Watch CPU, memory, database connections
3. **Check Prometheus**: Use /metrics endpoint to see detailed metrics
4. **Database tuning**: Adjust connection pool size if needed
5. **Redis tuning**: Monitor Redis memory usage

## Expected Performance

| Proof System | Generation Time | Throughput (req/s) |
|--------------|----------------|-------------------|
| Commitment   | < 100ms        | ~1000             |
| Groth16      | ~30s (async)   | Limited by workers|
| PLONK        | ~35s (async)   | Limited by workers|
| STARK        | ~40s (async)   | Limited by workers|

## Troubleshooting

### High error rate
- Check API logs: `railway logs --service zapiki`
- Verify database connections
- Check Redis connectivity

### Slow response times
- Increase worker count
- Scale API replicas
- Check database query performance

### Connection refused
- Verify BASE_URL is correct
- Check API is running
- Verify firewall rules

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Test Types](https://k6.io/docs/test-types/)
- [Prometheus Metrics](https://zapiki-production.up.railway.app/metrics)
