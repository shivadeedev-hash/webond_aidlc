# WeBond

WeBond is a retail investment platform to buy, sell, and manage **Bonds**, **Mutual Funds**, and **Gold** in one consolidated portfolio. It supports automated investing with DCA (Dollar-Cost Averaging), a cash wallet, transaction history, and a portfolio dashboard with insights.

This is a full-stack application:

- **web/** — React 18 + TypeScript (Vite) + Tailwind CSS single-page app.
- **api/** — Go 1.23 + Gin REST API.
- **MySQL 8** (8.0 locally, RDS 8.4 on AWS) — the real data store (sample data seeded via migrations/seed).

The frontend is built **mock-JSON-first** so the UI is demonstrable on Node alone, then the data layer is switched to the Go API via a single env flag (`VITE_DATA_SOURCE`).

## Repository layout

```
WeBond/
├── docker-compose.yml   # web + api + mysql + mailpit (local full stack)
├── .env.example         # root env sample (copy to .env)
├── scripts/             # setup-env, gen-env, dev-up, live-prices, test-db
├── docs/                # GETTING-STARTED.md (beginner guide), DEPLOYMENT.md (AWS)
├── deploy/aws/          # CloudFormation template + dry-run-by-default deploy/destroy scripts
├── .github/workflows/   # CI (lint/test/build) + manual deploy workflow
├── web/                 # React + TS + Vite + Tailwind frontend
└── api/                 # Go + Gin backend
```

## Prerequisites

| Tool | Needed for | Status on a fresh machine |
|------|-----------|---------------------------|
| Node.js 20+ / npm | Frontend (mock mode + build) | Required now |
| Go 1.22+ | Backend (real build) | Install via `scripts/setup-env.sh` |
| Docker + Compose v2 | One-command full stack + MySQL | Install via `scripts/setup-env.sh` |
| MySQL client (optional) | Local DB inspection | Optional |

> Only Node is required to run the frontend demo. Go, Docker, and MySQL are only needed for the real/backend build.

Check and install the real-build tooling (macOS/Homebrew, idempotent, safe to re-run):

```bash
./scripts/setup-env.sh
```

## How to run

```bash
git clone https://github.com/shivadeedev-hash/webond_aidlc.git && cd webond_aidlc
./scripts/dev-up.sh     # needs Docker Desktop; opens http://localhost:5173
```

> New to the project? Follow the step-by-step beginner guide (Thai): [docs/GETTING-STARTED.md](docs/GETTING-STARTED.md).
> It covers installing tools, creating the database tables, starting everything, logging in, and troubleshooting.

### 1) Frontend only (mock mode) — works with Node alone

```bash
cd web
cp .env.example .env     # VITE_DATA_SOURCE=mock
npm install
npm run dev              # http://localhost:5173
```

Other frontend scripts:

```bash
npm run build    # type-check + production build
npm run preview  # preview the production build
npm run test     # Vitest
npm run lint     # ESLint
```

### 2) Full stack in Docker (one command)

```bash
./scripts/dev-up.sh          # creates .env with random secrets if missing, builds and starts web + api + mysql
# web  → http://localhost:5173  (nginx, API mode, same-origin /api)
# api  → http://localhost:8080  (GET /health, GET /health/ready)
# mysql→ localhost:3306
```

E-mails (verification, password reset, notifications) are sent over SMTP. By default they go to the
local Mailpit inbox at http://localhost:8025 (nothing leaves the machine). To deliver to real inboxes for
free, set `SMTP_*` in `.env`, for example Gmail with an app password (see `.env.example` and
`docs/GETTING-STARTED.md` §4.4).

The API runs database migrations automatically on boot (`MIGRATE_ON_BOOT`, default on outside
production) and seeds the demo dataset (`SEED_ON_BOOT`). One-off admin commands are also available:
`docker compose exec api /app/webond migrate` and `docker compose exec api /app/webond seed`.

Helper scripts:

| Script | Purpose |
|--------|---------|
| `scripts/gen-env.sh [--rotate]` | Create `.env` with random JWT secrets (`--rotate` regenerates them) |
| `scripts/dev-up.sh [--no-build]` | Start the full stack and wait until healthy |
| `scripts/live-prices.sh on\|off\|status` | Toggle the simulated live price/NAV generator |
| `scripts/test-db.sh [--reset]` | Create/migrate the isolated `webond_test` database for integration tests |

### 3) Backend on the host (Go + a MySQL 8 server)

```bash
cd api
go run ./cmd/webond    # :8080, migrates + seeds on boot in non-production
```

### Tests

```bash
cd web && npm run test -- --run                     # frontend unit tests
./scripts/test-db.sh                                # once: create webond_test
export WEBOND_TEST_DSN='webond:webond@tcp(127.0.0.1:3306)/webond_test?parseTime=true&loc=Asia%2FBangkok'
cd api && go test -count=1 -tags=integration ./...  # backend unit + integration tests
```

Integration tests only read `WEBOND_TEST_DSN` and refuse to run against the demo database `webond`.

### Deploy to AWS

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) (Thai). `deploy/aws/deploy.sh` is dry-run by default and
only creates resources with `--apply` plus typing the stack name. Deploying creates billable resources.

## Configuration

All configuration is via environment variables (12-factor). Copy `.env.example` to `.env` and fill real values. `.env` is gitignored — no secrets are committed. See `.env.example` for the full list.

## Health endpoints

- `GET /health` — liveness, returns `{"status":"ok"}`.
- `GET /health/ready` — readiness (pings the database).

## Demo credentials

The database is populated with a demo dataset (idempotent seed at `api/db/seed/seed.sql`,
applied automatically on API boot in non-production via `SEED_ON_BOOT`, and re-runnable at
any time). Two pre-verified accounts are available:

| Role | Email | Password |
|------|-------|----------|
| Investor | `demo@webond.test` | `Demo@1234` |
| Admin | `admin@webond.test` | `Admin@1234` |

> These are **sample/dev-only** credentials for the demo dataset — not real users. Passwords
> are stored as bcrypt hashes (cost 12). Do not reuse them outside local/demo environments;
> seeding is disabled when `APP_ENV=production` (or `SEED_ON_BOOT=false`).

The demo investor comes with: an active status + valid risk assessment (level 4), a funded
wallet (balance + reserved), holdings across a mutual fund, a bond, and gold, two DCA plans
(fund + gold), and ~3 months of transaction history (with status timelines and a wallet
ledger). The catalog is seeded with several funds, bonds, and a gold product plus ~90 days of
NAV/price history, so the Market, Portfolio, Dashboard, and History pages show real data.

Quick login check against the running API:

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@webond.test","password":"Demo@1234"}'
```

### Seeding manually

The seed also applies on API boot, but you can re-run it at any time:

```bash
docker compose exec api /app/webond seed
```

To reset the demo database completely (deletes all local data): `docker compose down -v && ./scripts/dev-up.sh`.
