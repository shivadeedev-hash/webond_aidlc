# Integration

## Overview
All external systems are **mocked** in phase 1 and backed by the WeBond MySQL database (FR-26) — no real provider network calls (workshop guardrail). The Go API calls internal mock services (not HTTP out). An auto price-generator simulates market movement.

## External Integrations (mocked, DB-backed)

### Mock Bank / Payment
- **Purpose**: deposit, withdraw, direct debit for DCA.
- **Type**: internal service (no real HTTP). Reads/writes `wallets`, `wallet_transactions`.
- **Error handling**: validation errors → envelope error codes; no retries (synchronous, local).

### Mock Fund House / NAV
- **Purpose**: fund master + daily NAV + async order allotment.
- **Type**: internal service. Reads `funds`, `fund_navs`.
- **Behaviour**: MF buy/sell confirm NAV per cut-off (BR-02); manual `publish-nav` endpoint simulates async confirmation.

### Mock Bond Market
- **Purpose**: bond master, offering, price/yield, coupons.
- **Type**: internal service. Reads `bonds`, `bond_prices`, `bond_coupon_schedules`.

### Mock Gold Price
- **Purpose**: bid/ask tick.
- **Type**: internal service + price-generator ticker. Reads/writes `gold_prices`.
- **Behaviour**: 30s quote lock on trade (EX-08); block trade if price > 5 min old (EX-10).

### Mock Notification
- **Purpose**: email/SMS/OTP simulation.
- **Type**: internal service writing `notification_outbox`; viewable inbox endpoint. No real email send (D1-2).

## Auto Price-Generator (US-030 Add-on)
- **Mechanism**: in-process Go goroutine ticker. WHILE `price_generator_state.enabled` = true, every `interval_seconds` (default 60) it appends a new `gold_prices` row (bounded random walk on bid/ask) and periodically new `fund_navs` / `bond_prices` points; WHILE false, prices stay fixed.
- **Control**: `POST /api/admin/price-generator {enabled,intervalSeconds?,volatilityPct?}`; state persisted in `price_generator_state` (singleton) so it survives restarts.
- **Client**: frontend polls latest prices via TanStack Query refetch interval (no WebSocket in phase 1).
- **Error handling**: ticker errors logged (structured) and skipped; a failing tick never crashes the server; generator is isolated from request handling.

## Inter-unit Communication
Single deployable unit (comprehensive mode). Frontend ↔ backend is the only network boundary: REST/JSON over HTTP, JWT Bearer + refresh cookie. No message bus/queue in phase 1.

### Sync contract (frontend → backend)
- Transport: HTTPS REST; envelope `{data,error,meta}`; error codes per api-spec.md.
- Auth: access token (memory) in `Authorization`; silent refresh via `POST /api/auth/refresh` on 401 (EX-03).

## Integration Testing
- **Strategy**: backend integration tests run services + sqlc repos against a MySQL test database (Docker), seeded via `db/seed`. Scheduler/price events exercised via the manual-trigger endpoints.
- **Mocking**: no external HTTP to mock (all internal). Time-dependent logic (cut-off, DCA next-run, quote expiry) uses an injectable clock so tests control time.
- **Frontend**: service interface mocked by the `mock.ts` implementation (JSON fixtures); component tests assert rendering + interactions.
