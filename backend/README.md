# Restaurant API — Backend

Go · Echo · GORM · MySQL · Redis · Meilisearch. Stateless HTTP API plus a
separate background worker. Built phase-by-phase per
[`../restaurant-app-plan.md`](../restaurant-app-plan.md).

> **Status:** Phase 1 (Foundation) complete. Auth, catalog, search, i18n,
> orders, geo, payments, etc. follow in later phases.

## Layout

```
cmd/
  api/        HTTP server entrypoint
  worker/     Asynq worker entrypoint (jobs wired in a later phase)
  migrate/    embedded-migration runner (up/down/version/force)
internal/
  config/     env loading + validation (the only reader of os env)
  repository/ data access; GORM connection + pool config
  cache/      Redis client (cache DB index)
  middleware/ request-id, access log, recover, CORS, body limit, rate limit
  handlers/   thin Echo handlers (health: /healthz, /readyz)
  server/     Echo assembly, error envelope handler, graceful shutdown
pkg/
  logger/     slog setup (console|json) + context helpers
  response/   standard JSON envelope + stable error codes
  pagination/ offset pagination params + meta
migrations/   embedded SQL migrations (golang-migrate)
```

**Layering rule:** `handlers → services → repository → DB`. Handlers never touch
GORM; services never write SQL.

## Quick start

### 1. Configure

```bash
cp .env.example .env
# edit .env if your MySQL/Redis differ from the defaults
```

### 2. Infrastructure

With Docker:

```bash
docker compose up -d mysql redis meilisearch
```

Or point `.env` at an existing local MySQL/Redis. The dev defaults expect a
`restaurant` database and `restaurant`/`restaurant` user:

```sql
CREATE DATABASE restaurant CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'restaurant'@'%' IDENTIFIED BY 'restaurant';
GRANT ALL PRIVILEGES ON restaurant.* TO 'restaurant'@'%';
FLUSH PRIVILEGES;
```

### 3. Migrate + run

```bash
go run ./cmd/migrate up   # apply migrations
go run ./cmd/api          # start API on :8080
```

### 4. Verify

```bash
curl localhost:8080/healthz   # liveness  -> 200 {"data":{"status":"ok"}}
curl localhost:8080/readyz    # readiness -> 200 if mysql+redis up, else 503
```

## Common tasks

```bash
make build           # build api/worker/migrate into bin/
make test            # run tests
make check           # fmt + vet + test
make migrate-up      # apply migrations
make migrate-down    # roll back one
```

## Conventions

- **Money:** integer cents everywhere (`*_cents`), never floats. Currency `CAD`.
- **Responses:** every endpoint returns the `response.Envelope`
  (`{data,meta}` or `{error:{code,message}}`). Error `code`s are stable.
- **Config:** all configuration flows through `internal/config`; nothing else
  reads the environment. Production fails fast on weak secrets / wildcard CORS.
- **Context:** request contexts carry a deadline + request-scoped logger; pass
  them to every DB/Redis call.
- **Migrations:** embedded in the binary; create new ones as
  `migrations/0000NN_name.up.sql` / `.down.sql`.
