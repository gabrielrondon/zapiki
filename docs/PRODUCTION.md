# Zapiki Production Deployment Guide

## Overview

This guide covers deploying Zapiki to production on Railway.app. Railway provides managed PostgreSQL, Redis, and easy deployment from GitHub.

## Prerequisites

- Railway account (https://railway.app)
- GitHub repository (already set up)
- Railway CLI (optional but recommended)

## Architecture

```
┌─────────────────────────────────────────┐
│            Railway Project              │
│                                         │
│  ┌──────────┐  ┌──────────┐           │
│  │   API    │  │  Worker  │           │
│  │ Service  │  │ Service  │           │
│  └────┬─────┘  └────┬─────┘           │
│       │             │                  │
│       │   ┌─────────┴──────┐          │
│       │   │                 │          │
│  ┌────▼───▼────┐   ┌────────▼─────┐  │
│  │  PostgreSQL │   │     Redis     │  │
│  └─────────────┘   └───────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

## Step-by-Step Deployment

### 1. Create Railway Project

```bash
# Install Railway CLI (optional)
npm i -g @railway/cli

# Login to Railway
railway login

# Create new project
railway init
```

Or use the Railway dashboard at https://railway.app

### 2. Add Services

#### A. PostgreSQL Database

1. Go to your Railway project
2. Click "New Service"
3. Select "Database" → "PostgreSQL"
4. Railway will automatically create the database
5. Note: `DATABASE_URL` environment variable is auto-created

#### B. Redis

1. Click "New Service"
2. Select "Database" → "Redis"
3. Railway will automatically create Redis
4. Note: `REDIS_URL` environment variable is auto-created

#### C. API Service

1. Click "New Service"
2. Select "GitHub Repo"
3. Connect your `zapiki` repository
4. Railway will detect the `Dockerfile` and build automatically

**Environment Variables for API**:
```bash
# Server
API_PORT=8080
ENV=production

# Database (automatically set by Railway if using their PostgreSQL)
# If using Railway PostgreSQL, it sets DATABASE_URL
# We need to parse it or set individual variables:
POSTGRES_HOST=${{Postgres.PGHOST}}
POSTGRES_PORT=${{Postgres.PGPORT}}
POSTGRES_USER=${{Postgres.PGUSER}}
POSTGRES_PASSWORD=${{Postgres.PGPASSWORD}}
POSTGRES_DB=${{Postgres.PGDATABASE}}
POSTGRES_SSLMODE=require

# Redis (automatically set by Railway if using their Redis)
REDIS_HOST=${{Redis.REDIS_HOST}}
REDIS_PORT=${{Redis.REDIS_PORT}}
REDIS_PASSWORD=${{Redis.REDIS_PASSWORD}}

# Proof Systems
ENABLE_COMMITMENT=true
ENABLE_GROTH16=true
ENABLE_PLONK=true
ENABLE_STARK=false

# Rate Limiting
RATE_LIMIT_FREE_TIER=100
RATE_LIMIT_PRO_TIER=10000
```

**Start Command**: `./zapiki-api`

#### D. Worker Service

1. Click "New Service"
2. Select "GitHub Repo" (same repository)
3. Use the same `Dockerfile`

**Environment Variables for Worker**:
Same as API (can reference the same database and Redis)

**Start Command**: `./zapiki-worker`

### 3. Database Migration

Railway doesn't automatically run migrations, so we need to do it manually:

**Option 1: Railway CLI**
```bash
# Connect to your project
railway link

# Run migration
railway run psql $DATABASE_URL < deployments/docker/schema.sql
```

**Option 2: Railway Dashboard**
1. Go to PostgreSQL service
2. Click "Query" tab
3. Paste contents of `deployments/docker/schema.sql`
4. Execute

**Option 3: One-time Job**
Create a one-time deployment that runs migrations:
```bash
# In Railway dashboard, add a service with:
# Start command: psql $DATABASE_URL < /app/deployments/docker/schema.sql
```

### 4. Initialize Templates

After migration, initialize templates:

```bash
# Using Railway CLI
railway run psql $DATABASE_URL < scripts/seed-templates.sql

# Or use the Railway dashboard Query tab
```

### 5. Verify Deployment

**Check API Health**:
```bash
curl https://your-railway-app.railway.app/health
```

**Expected Response**:
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

**Get API Key**:
```bash
# Connect to database
railway run psql $DATABASE_URL

# Query for API key
SELECT key FROM api_keys WHERE name = 'Test API Key' LIMIT 1;
```

**Test Proof Generation**:
```bash
curl -X POST https://your-railway-app.railway.app/api/v1/proofs \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "proof_system": "commitment",
    "data": {
      "type": "string",
      "value": "production test"
    }
  }'
```

## Configuration Details

### Railway Environment Variables

Railway uses a special syntax for referencing other services:

```bash
# Reference PostgreSQL service
POSTGRES_HOST=${{Postgres.PGHOST}}

# Reference Redis service
REDIS_HOST=${{Redis.REDIS_HOST}}
```

This automatically connects services together.

### Dockerfile

The `Dockerfile` uses multi-stage build:
1. **Builder stage**: Compiles Go binaries
2. **Runtime stage**: Minimal Alpine image with just the binary

**Size**: ~50MB (vs 29MB local binary due to Alpine base)

### Health Checks

Railway uses the `/health` endpoint to monitor service health:
- **Healthy**: Returns 200 with service status
- **Unhealthy**: Returns 503 if any service is down

### Auto-Deployment

Railway automatically deploys when you push to GitHub:
```bash
git push origin main
# Railway detects push and deploys automatically
```

## Scaling

### Horizontal Scaling

**API Service**:
```
railway scale --replicas 3
```

**Worker Service**:
```
railway scale --replicas 5
```

Railway load balances across replicas automatically.

### Vertical Scaling

Upgrade Railway plan for more resources:
- **Starter**: 512MB RAM, 0.5 vCPU
- **Developer**: 8GB RAM, 8 vCPU
- **Team**: 32GB RAM, 32 vCPU

### Database Scaling

Railway PostgreSQL auto-scales storage. For more performance:
1. Upgrade Railway plan
2. Or migrate to dedicated database (AWS RDS, etc.)

## Monitoring

### Railway Dashboard

Railway provides built-in monitoring:
- **Logs**: Real-time logs for all services
- **Metrics**: CPU, Memory, Network usage
- **Deployments**: History and rollback

### Application Logs

View logs:
```bash
# API logs
railway logs --service api

# Worker logs
railway logs --service worker
```

### Custom Monitoring (Optional)

Add external monitoring:
- **Sentry**: Error tracking
- **Datadog**: Metrics and APM
- **Prometheus**: Custom metrics (requires setup)

## Custom Domain

### Setup Custom Domain

1. Go to API service in Railway
2. Click "Settings" → "Domains"
3. Click "Add Domain"
4. Enter your domain: `api.zapiki.io`
5. Add DNS records as shown by Railway

**DNS Records**:
```
Type: CNAME
Name: api
Value: your-service.railway.app
```

### SSL/TLS

Railway automatically provisions SSL certificates for custom domains using Let's Encrypt.

## Security Checklist

### Before Production

- [ ] Change default API key
- [ ] Set strong database password
- [ ] Enable PostgreSQL SSL (`POSTGRES_SSLMODE=require`)
- [ ] Set appropriate rate limits
- [ ] Review CORS settings
- [ ] Enable production mode (`ENV=production`)
- [ ] Backup database regularly
- [ ] Setup monitoring/alerting
- [ ] Review and rotate secrets

### Create Production API Keys

```sql
-- Connect to production database
railway run psql $DATABASE_URL

-- Create production user
INSERT INTO users (email, name, tier)
VALUES ('production@yourcompany.com', 'Production User', 'pro');

-- Create production API key
INSERT INTO api_keys (user_id, key, name, rate_limit)
SELECT id,
       'prod_' || encode(gen_random_bytes(32), 'hex'),
       'Production API Key',
       10000
FROM users
WHERE email = 'production@yourcompany.com';

-- Get the key
SELECT key FROM api_keys WHERE name = 'Production API Key';
```

## Backup & Recovery

### Database Backup

Railway provides automatic backups. To create manual backup:

```bash
# Backup database
railway run pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql

# Restore database
railway run psql $DATABASE_URL < backup-20240130.sql
```

### Disaster Recovery

If something goes wrong:

1. **Rollback Deployment**:
   - Go to Railway dashboard
   - Click "Deployments"
   - Click "Rollback" on previous working deployment

2. **Restore Database**:
   ```bash
   railway run psql $DATABASE_URL < backup.sql
   ```

## Cost Estimation

### Railway Pricing (as of 2024)

**Free Tier**:
- $5 free credit/month
- Good for testing

**Starter Plan** ($5/month):
- $5 credit + pay-as-you-go
- ~$20-50/month for small production

**Developer Plan** ($20/month):
- $20 credit + pay-as-you-go
- ~$50-100/month for medium traffic

**Estimated Monthly Cost** (small production):
```
API Service:       $10-20
Worker Service:    $10-20
PostgreSQL:        $5-10
Redis:             $5-10
────────────────────────
Total:            $30-60/month
```

## Performance Optimization

### Database

1. **Connection Pooling**: Already configured (25 max connections)
2. **Indexes**: All critical columns indexed
3. **Query Optimization**: Use `EXPLAIN` for slow queries

### API

1. **Compression**: Enable gzip (add middleware)
2. **Caching**: Redis caching for verification keys
3. **CDN**: Use Cloudflare for static assets

### Worker

1. **Concurrency**: Increase worker count based on load
2. **Queue Priority**: Already configured (high/normal/low)
3. **Batch Processing**: Group similar jobs

## Troubleshooting

### API Not Starting

Check logs:
```bash
railway logs --service api
```

Common issues:
- Database connection failed → Check `POSTGRES_*` vars
- Redis connection failed → Check `REDIS_*` vars
- Port binding → Ensure `API_PORT=8080`

### Worker Not Processing Jobs

Check logs:
```bash
railway logs --service worker
```

Common issues:
- Redis connection failed
- No jobs in queue → Create test proof
- Worker crashed → Check for panic/errors

### Database Migration Failed

Manually run migration:
```bash
railway run psql $DATABASE_URL < deployments/docker/schema.sql
```

### Slow Proof Generation

- Check worker count: `railway ps`
- Scale workers: `railway scale --replicas 3 --service worker`
- Check job queue: Monitor Redis

## Maintenance

### Regular Tasks

**Weekly**:
- Review error logs
- Check disk usage
- Monitor rate limit hits

**Monthly**:
- Database backup verification
- Performance review
- Cost review
- Security updates

### Updates

Deploy updates:
```bash
git push origin main
# Railway auto-deploys
```

Zero-downtime deployment:
1. Railway builds new version
2. Health check passes
3. Traffic switches to new version
4. Old version terminated

## Support & Resources

- **Railway Docs**: https://docs.railway.app
- **Railway Discord**: https://discord.gg/railway
- **Zapiki Issues**: https://github.com/gabrielrondon/zapiki/issues

## Next Steps

After deployment:
1. [ ] Test all endpoints in production
2. [ ] Monitor for 24-48 hours
3. [ ] Setup alerting
4. [ ] Document API endpoints
5. [ ] Share API with users
6. [ ] Plan Phase 6-8 implementation

---

**Your Zapiki API is now in production!** 🚀

Access it at: `https://your-project.railway.app`
