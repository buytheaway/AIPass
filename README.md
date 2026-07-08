# AIPass

AIPass is a production-like training project for QR-based access control and subscriptions. It is intentionally understandable for junior/middle onboarding while still using a real backend stack: Go, Echo, PostgreSQL, Redis, Redpanda/Kafka, MinIO, Prometheus, Grafana, Jaeger, Docker, Helm, GitLab CI, and tests.

## Services

- `access-api`: admin API for auth, users, plans, subscriptions, QR passes, payments, reports, metrics.
- `scanner-gateway`: serves `/scanner`, validates QR tokens, writes access events, publishes Kafka events.
- `notification-report-service`: consumes access events and sends Telegram messages when configured.

## Local Start

JWT keys are optional in local V1. If PEM files are missing, `access-api` creates an in-memory RSA keypair so login can still return a JWT. To use stable local keys, run:

```powershell
.\scripts\gen_rsa_keys.ps1
```

Start the V1 stack:

```powershell
docker compose up --build -d postgres redis minio redpanda access-api
```

Apply migrations:

```powershell
.\scripts\migrate.ps1
```

Useful URLs:

- access API: `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`
- MinIO console: `http://localhost:9001` with `minioadmin/minioadmin`

Local seed admin:

- email: `admin@aipass.local`
- password: `admin123`

This seed is for local training only.

## Main Request Flow

1. Admin logs in through `POST /api/v1/auth/login`.
2. `access-api` reads the user and bcrypt hash from PostgreSQL.
3. On success, `access-api` returns an RSA JWT access token.

## Example API Flow

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@aipass.local","password":"admin123"}'
```

Create plan:

```bash
curl -X POST http://localhost:8080/api/v1/plans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Monthly","duration_days":30,"price":"25000","currency":"KZT"}'
```

## What Each Version Teaches

- V0: Go service startup, config, logging, health checks.
- V1: PostgreSQL schema, SQL migrations, repository pattern.
- V2: RSA JWT auth, password hashing, middleware.
- V3: Echo handlers, DTOs, validation, Swagger placeholder.
- V4: QR token generation, SHA-256 token storage, subscription rules.
- V5: Real scanner UI, browser camera API, check-in/check-out flow.
- V6: Kafka/Redpanda events and Telegram worker.
- V7: Redis rate limiting and short-lived scanner helpers.
- V8: MinIO file buckets and payment receipt flow.
- V9: Excel reports, metrics, Grafana, Jaeger basics.
- V10: Docker multi-stage, Helm, GitLab CI, integration tests.

## Current Scope

This first implementation is a working baseline. V1 focuses on PostgreSQL, migrations, seed admin, real auth login, health/readiness, and basic Swagger.

Redis, Redpanda, and MinIO are present as optional local infrastructure, but they are not required for the login flow in V1.
