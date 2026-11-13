# Tech

## Summary
- **Stack**: React 18 + TypeScript (Vite) + Tailwind CSS frontend; Go 1.22+ (Gin) REST API; MySQL 8.0 as the real database (mock/sample data seeded into the DB, not JSON files in production).
- **Architecture**: SPA (layered: pages → components → hooks → services) calling a real Go REST API (handler → service → repository → MySQL). Frontend is built first against mock JSON for fast UI iteration, then the data layer is switched to the Go API.
- **Infra**: Local via Docker Compose (web, api, mysql). Target: AWS ECS Fargate, RDS MySQL, ElastiCache Redis, S3 + CloudFront.

## Build Direction (confirmed with user)
- Full stack: **React frontend + Go (Gin) backend + MySQL**.
- Sample/mock data is **seeded into MySQL** (migrations + seed); backend reads from the real DB. Mock external systems (bank/payment, fund house, bond market, gold price, notification) are backed by the same MySQL data per FR-26.
- Execution order (hands-on policy): Setup → Frontend on mock JSON (see the UI working) → convert data layer to Go API + MySQL → verify → deploy (optional).

## Stack
- **Languages**: TypeScript (frontend); Go (backend — later phase)
- **Frameworks**: React 18, React Router, TanStack Query, Recharts, Tailwind CSS (frontend); Gin, JWT, sqlc/GORM (backend)
- **Build system**: Vite + npm (frontend); Go modules + golang-migrate (backend)
- **Package manager**: npm
- **Testing**: Vitest + React Testing Library (frontend); Go testing + testify (backend). Pending D3 confirmation.

## Architecture (settled at D3)
- **Pattern**: Layered SPA (pages → components → hooks → service interface) calling a layered Go API (handler → service → repository/sqlc → MySQL). Data layer swappable via `VITE_DATA_SOURCE` (mock | api).
- **API style**: REST/JSON, envelope `{ data, error, meta }` with typed error codes.

## Infrastructure
- **Cloud provider**: AWS (target). Local dev via Docker Compose.
- **Compute**: ECS Fargate (target); local Vite dev server + Go API container.
- **Database**: MySQL 8.0 — the real data store. Sample data seeded via migrations/seed scripts. DECIMAL for all money/unit values. (Frontend-only step uses mock JSON temporarily before the API is wired.)
- **IaC tool**: Docker Compose for local; CloudFormation/CDK/Terraform TBD (D3).

## Patterns & Conventions
- **Architecture pattern**: Layered frontend (route pages → feature components → hooks → service interface); layered Go backend (handler → service → repo).
- **Data access**: Frontend — typed service interface with `mock.ts` / `http.ts` impls (env-selected). Backend — **sqlc** type-safe queries over MySQL; money via **shopspring/decimal** ↔ DECIMAL(18,4).
- **API response format**: Envelope `{ data, error, meta }`; typed error `{ code, message }` with spec codes (INSUFFICIENT_BALANCE, QUOTE_EXPIRED, BOND_OFFER_INVALID, ...).
- **Error handling**: typed codes per api-spec.md; middleware recovery; structured error logs.
- **Authentication**: JWT access (15 min, in memory) + refresh (7 days, httpOnly secure cookie), silent refresh on 401; bcrypt ≥ 12; lockout after 5 fails.
- **Validation**: request validation at handler boundary (Go) + form validation on FE.
- **Logging**: Go `log/slog` structured JSON; request-id correlation; Minimal observability (logs + health).
- **Code style**: ESLint + Prettier (FE); gofmt/go vet (BE).
- **Naming conventions**: camelCase variables/functions, PascalCase components/classes, UPPER_SNAKE_CASE constants, private methods prefixed `_`. One component per file; group related components in folders; `index.ts` for exports. Interfaces over types for public APIs; explicit return types on exported functions; avoid `any`. (From steering: js-conventions, code-conventions.)
- **Branch strategy**: Pending D3 (`git init -b main` at setup per workshop policy).

## Environment Configuration
- **Config approach**: 12-factor — config via env vars (`.env`, Vite `import.meta.env`). No secrets in source.
- **Environments**: local (phase 1). Dev/staging/prod later with dev/prod parity.
- **Secrets management**: env vars locally; secrets manager in cloud (target). Never commit secrets.

## CI/CD Pipeline
- **Tool**: GitHub Actions (target). Pending D3.
- **Stages**: lint → test → build (→ deploy later).
- **Deploy target**: S3 + CloudFront (web), ECS Fargate (API) — target.

## Dependency Management
- **Lockfile**: package-lock.json (npm).
- **Version strategy**: pinned/exact versions for new dependencies (per safety guidance).
- **Monorepo tooling**: Pending D3 — recommend a single repo with `web/` (and later `api/`).
