# AIPass

AIPass is a production-like training project for QR-based access control and subscriptions. It is intentionally understandable for junior/middle onboarding while still using a real backend stack: Go, Echo, PostgreSQL, Redis, Redpanda/Kafka, MinIO, Prometheus, Grafana, Jaeger, Docker, Helm, GitLab CI, and tests.

## Services

- `access-api`: admin API for auth, users, plans, subscriptions, QR passes, payments, reports, metrics.
- `scanner-gateway`: serves `/scanner`, validates QR tokens, writes access events, publishes Kafka events.
- `notification-report-service`: consumes access events and sends Telegram messages when configured.

## Local Start

Generate local JWT keys first. Run this before `docker compose up --build` so keys are copied into service images:

```powershell
.\scripts\gen_rsa_keys.ps1
```

Start the stack:

```powershell
docker compose up --build
```

Useful URLs:

- access API: `http://localhost:8080`
- scanner UI: `http://localhost:8081/scanner`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` with `admin/admin`
- MinIO console: `http://localhost:9001` with `minioadmin/minioadmin`
- Jaeger: `http://localhost:16686`

Local seed admin:

- email: `admin@aipass.local`
- password: `password`

This seed is for local training only.

## Main Request Flow

1. Admin logs in through `POST /api/v1/auth/login`.
2. Admin creates a member through `POST /api/v1/users`.
3. Admin creates a plan through `POST /api/v1/plans`.
4. Admin assigns a subscription through `POST /api/v1/users/:id/subscriptions`.
5. Admin activates the subscription through `PATCH /api/v1/subscriptions/:id/status`.
6. Admin generates QR token through `POST /api/v1/subscriptions/:id/qr-pass`.
7. Scanner UI reads the QR token and calls `POST /api/v1/scans/validate`.
8. Scanner gateway validates PostgreSQL state, creates an access event, and publishes `access.events.v1`.
9. Notification service consumes the event and sends Telegram when env vars are configured.

## Example API Flow

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@aipass.local","password":"password"}'
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

This first implementation is a working baseline. It includes the service layout, schema, auth, users, plans, subscriptions, QR generation, scanner validation, Kafka publishing, Telegram consumer, Excel reports, metrics, Docker Compose, Helm/CI scaffolding, and a real browser scanner page.

MinIO binary upload endpoints and full Swagger generation are scaffolded but intentionally left for later versions so the project stays teachable step by step.
