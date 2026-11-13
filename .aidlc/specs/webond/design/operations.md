# Operations Design

## Summary
- **Observability Level**: Minimal (D3-12) — structured logging + health endpoints.
- **Error Tracking**: Log-based only (D3-13).
- **Lifecycle Management**: Health + readiness + graceful shutdown (D3-14).
- **Logging Format**: Structured JSON (Go `log/slog`), human-readable in dev.
- **Logging Library**: Go standard `log/slog`.

---

## Local Environment Setup (prerequisite for running the project)

> Detected on this machine: ✅ Node 24.14.1, ✅ npm 11.11.0, ✅ git 2.54.0 · ❌ Go, ❌ Docker, ❌ MySQL.

Two ways to run WeBond:

### Option A — Docker Compose (recommended, one command)
Requires Docker Desktop (includes Compose v2). After install:
```
docker compose up --build
# web → http://localhost:5173 (or served build), api → http://localhost:8080, mysql → 3306
# migrations + seed run automatically on first start
```
Install Docker Desktop: https://www.docker.com/products/docker-desktop/ (brings docker + compose + a MySQL container, so no separate Go/MySQL install needed to run).

### Option B — Run on host (no Docker)
Requires Go + a MySQL 8 server:
```
# macOS via Homebrew:
brew install go mysql
brew services start mysql
# create DB + user, set DATABASE_URL, then:
cd api && go mod tidy && go run ./cmd/webond     # :8080
cd web && npm install && npm run dev             # :5173 (VITE_DATA_SOURCE=api)
```

### Frontend-only demo (works right now, Node only)
```
cd web && npm install && npm run dev     # VITE_DATA_SOURCE=mock → full UI on JSON fixtures
```
This needs no Go/Docker/MySQL and is the phase-1 "see the UI working" path.

---

## Logging
- **Format**: structured JSON in `APP_ENV=production`, pretty text in development.
- **Levels**: error, warn, info, debug (`LOG_LEVEL`).
- **Correlation**: `request-id` middleware — generate if absent, echo in responses + logs.

### Per-Component Logging
| Component | Key Log Events | Level | Context Fields |
|---|---|---|---|
| HTTP middleware | request start/end, status, duration | info | request-id, method, path, status, duration_ms, user-id? |
| AuthService | login success/fail, lockout, reset | info/warn | request-id, user-id (masked email) |
| TradingService | order created/completed/failed, reserve/release | info | request-id, user-id, txn_ref, asset_type, code |
| DCAService/Scheduler | plan run start/result, round success/fail | info/warn | plan-id, run-date, round, reason |
| BondLifecycleService | coupon paid, maturity closed | info | bond-id, user-id, amount |
| PriceGenerator | tick written, toggle change, tick error | info/warn | asset_type, enabled, error? |
| Repositories | query error only | error | operation, table |

### Sensitive Data Rules
- **Never log**: passwords, token values, full bank account, national ID.
- **Mask**: email (first 3 + domain), bank account (last 4).
- **Always include**: request-id, timestamp, operation, user-id (if authed), duration.

## Health & Readiness
| Endpoint | Purpose | Auth | Response |
|---|---|---|---|
| `GET /health` | liveness | none | `200 {"status":"ok","uptime":N}` |
| `GET /health/ready` | readiness | none | `200 {"status":"ready","checks":{"database":{"status":"ok","latency_ms":N}}}` or `503` |

### Readiness checks
| Dependency | Check | Timeout | On failure |
|---|---|---|---|
| MySQL | `db.PingContext` | 2s | 503 not_ready |

## Graceful Shutdown
```
SIGTERM:
 1. stop accepting new connections
 2. /health/ready → 503 (drain)
 3. wait in-flight requests (SHUTDOWN_TIMEOUT_MS, default 30s)
 4. stop price-generator goroutine (ctx cancel)
 5. close DB pool, flush logs
 6. exit 0 (exit 1 on timeout)
```
| Setting | Default | Env |
|---|---|---|
| Shutdown timeout | 30s | `SHUTDOWN_TIMEOUT_MS` |

## Error Handling & Reporting
| Category | HTTP | Log Level | Retry? | Example |
|---|---|---|---|---|
| Client error | 4xx | warn | No | validation, INSUFFICIENT_BALANCE, QUOTE_EXPIRED |
| Operational | 5xx | error | caller may retry | DB timeout |
| Programming | 5xx | error | No | nil deref (recovered by middleware) |
| Fatal | — | fatal | No | cannot connect DB on startup → exit 1 |

Error log carries request-id, error code, operation, sanitized input; stack only in non-production.

## Configuration Management
| Variable | Required | Default | Sensitive | Description |
|---|---|---|---|---|
| `PORT` | No | 8080 | No | API port |
| `APP_ENV` | No | development | No | runtime env |
| `LOG_LEVEL` | No | info | No | log level |
| `DATABASE_URL` | Yes | — | Yes | MySQL DSN |
| `JWT_ACCESS_SECRET` | Yes | — | Yes | access token secret |
| `JWT_REFRESH_SECRET` | Yes | — | Yes | refresh token secret |
| `SHUTDOWN_TIMEOUT_MS` | No | 30000 | No | graceful shutdown |

### Startup validation
On boot: validate required env vars (exit 1 with clear message if missing), validate DSN format, ping DB (exit 1 if unreachable), run migrations, optionally seed.

## Traceability
| Operations concern | Related design |
|---|---|
| Per-component logging | components.md |
| Readiness targets | integration.md, data-model.md |
| Error classification | api-spec.md error codes |
| Config variables | implementation.md env table |
