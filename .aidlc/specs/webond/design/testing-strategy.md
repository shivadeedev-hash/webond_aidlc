# Testing Strategy

## Overview
- **Philosophy**: Hybrid — test-after for UI, property + example tests for financial logic (the riskiest code). Lazy about ceremony, strict about money correctness.
- **Coverage target**: ≥ 70% for backend business logic (NFR-08); frontend covers core flows + formatting.
- **Test pyramid**: many unit (money/services/components) → fewer integration (service+DB) → a few happy-path flow checks.

## Frameworks
- **Backend unit**: `go test` + stretchr/testify.
- **Backend property**: leanovate/gopter (money invariants — correctness.md).
- **Backend integration**: `go test -tags=integration` against MySQL test DB (Docker).
- **Frontend unit/component**: Vitest + React Testing Library.
- **Load**: not in phase 1 (NFR-04 targets documented, not load-tested now).

## Test Architecture
```
api/
  internal/money/money_test.go           # example unit
  internal/money/money_property_test.go  # gopter P1,P4,P6
  internal/service/*_test.go             # service unit (mock repos / clock)
  internal/service/trading_property_test.go     # P2,P3,P5
  test/integration/*_test.go             # service+sqlc+MySQL (build tag)
web/
  src/**/*.test.tsx                      # component tests
  src/lib/format.test.ts                 # THB/units/date formatting (BR-15)
  src/test/setup.ts                      # RTL setup
```
- **Naming**: Go `TestXxx` / `Property_Xxx`; FE `*.test.tsx`.
- **Shared utilities**: Go test factories for users/orders/holdings + injectable clock; FE JSON fixtures reused from `mocks/`.
- **Isolation**: integration tests wrap each test in a DB transaction rolled back at the end (or truncate+reseed).

## Mock Strategy
- **External services**: none real — internal mock services; nothing to stub over HTTP.
- **Database**: unit tests use in-memory fakes/mock repos; integration tests use a real MySQL (Docker).
- **Time-dependent**: injectable `Clock` interface controls cut-off (BR-02), DCA next-run, 30s quote expiry (EX-08), 5-min price staleness (EX-10).
- **Frontend**: service interface swapped to `mock.ts`; no network in component tests.

## Environment
- **Test DB**: ephemeral MySQL via Docker Compose (`docker compose -f docker-compose.test.yml up`) or the dev DB with a `_test` schema; migrated before run.
- **Env vars**: `APP_ENV=test`, separate `DATABASE_URL`.
- **CI**: GitHub Actions (later) runs `go test ./...` + `npm run test` + lint.

## Coverage Mapping
| Component | Test type |
|---|---|
| money/* (allotment, avg-cost, gold, coupon, signs) | property + unit |
| TradingService (orders, reserve, idempotency, rules) | unit + integration |
| WalletService (deposit/withdraw/reserve) | unit + integration |
| DCAService + scheduler (idempotent rounds, pause-after-3) | unit + integration |
| PortfolioService / DashboardService | unit |
| HistoryService (filter/cancel/export) | unit + integration |
| Auth (bcrypt, JWT, lockout, reset) | unit + integration |
| FE formatting (format.ts) | unit |
| FE pages (Market, AssetDetail, History, DCA) | component |

| Endpoint group | Test type |
|---|---|
| /api/auth/* | integration |
| /api/orders* (+ idempotency, quote) | integration |
| /api/wallet/* | integration |
| /api/transactions* (+ cancel, export) | integration |
| /api/dca-plans* | integration |

## Critical paths (flow checks)
- Register → verify → risk assessment → deposit → buy MF → appears in Portfolio + History.
- Create gold DCA → manual run → round + transaction created.
- Cancel pending MF buy before cut-off → cash released.

## Run commands
- Backend unit+property: `cd api && go test ./...`
- Backend integration: `cd api && go test -tags=integration ./test/integration/...`
- Frontend: `cd web && npm run test`
- Coverage: `cd api && go test -cover ./...` ; `cd web && npm run test -- --coverage`
