# Design Document: WeBond

## Summary
<!-- 10-line max digest for downstream phases. -->
- **Architecture**: Layered full-stack SPA + REST API — React talks to a Go/Gin API over MySQL; frontend built mock-JSON-first then swapped to the API via one env flag.
- **Stack**: React 18 + TS (Vite) + Tailwind / Go 1.23 (Gin) / MySQL 8.0 / Docker Compose.
- **Components**: ~12 FE modules + ~13 BE services (Auth, Profile, Wallet, Catalog, Trading, DCA+Scheduler, Portfolio, Dashboard, History, BondLifecycle, MockExternal+PriceGenerator, Notification, Admin).
- **Entities**: 20 — users, risk_assessments, wallets, wallet_transactions, funds, fund_navs, bonds, bond_prices, bond_coupon_schedules, gold_products, gold_prices, orders, holdings, dca_plans, dca_rounds, transactions, transaction_status_history, notification_outbox, audit_logs, price_generator_state.
- **Endpoints**: REST under `/api` + `/health`, `/health/ready`; envelope `{data,error,meta}`, spec error codes, Idempotency-Key on orders.
- **Integrations**: 5 mocked systems backed by MySQL + auto price-generator (toggle).
- **Testing**: PBT Yes (gopter, money invariants) — NFR Yes.
- **Key Decisions**: sqlc + shopspring/decimal for exact money; token in memory + httpOnly refresh cookie; auto price-generator goroutine with on/off toggle.

## Architecture

### System Context Diagram
```
                       ┌─────────────────────────────┐
   Browser  ──HTTPS──▶ │  Go (Gin) API  :8080        │
 (React SPA :5173)     │  router→service→repo(sqlc)  │ ──▶ MySQL 8.0 :3306
   mock|api (env)      │  PriceGenerator goroutine   │ ◀── migrations + seed
                       │  mock external services     │
                       └─────────────────────────────┘
   All external providers (bank, fund house, bond market, gold, notify) = internal mocks over the same DB.
```

### Technology Stack
- **Frontend**: React 18, TypeScript, Vite, Tailwind, React Router, TanStack Query, Recharts.
- **Backend**: Go 1.23, Gin, golang-jwt v5, shopspring/decimal, sqlc, go-sql-driver/mysql, xuri/excelize, gopter.
- **Database**: MySQL 8.0 (InnoDB, DECIMAL(18,4)), golang-migrate.
- **Infrastructure**: Docker Compose (web + api + mysql) for local; AWS target later.

### Key Design Decisions
1. **Exact money math**: sqlc + shopspring/decimal ↔ DECIMAL(18,4), order/balance mutations in a single DB transaction, idempotency keys (NFR-05).
2. **Auth posture**: access token in memory + refresh token in httpOnly secure cookie with silent refresh (NFR-01, OWASP XSS/CSRF).
3. **Mock→API swap**: one typed service interface per domain, `mock.ts`/`http.ts` picked by `VITE_DATA_SOURCE`, so UI code never changes between phases (D3-6).
4. **Simulated realtime**: in-process goroutine price-generator with a persisted on/off toggle (US-030 Add-on), frontend polls.

## Traceability

| Requirement | Component(s) | API Endpoint(s) | Data Entity | Design File | Status |
|---|---|---|---|---|---|
| US-001 Register | AuthModule / AuthService | POST /api/auth/register | users, notification_outbox | components, api-spec, data-model | ✅ |
| US-002 Verify email | AuthModule / AuthService | POST /api/auth/verify; GET /api/inbox | users, notification_outbox | api-spec | ✅ |
| US-003 Login/Logout | AuthModule / AuthService | POST /api/auth/login,refresh,logout | users | api-spec, operations | ✅ |
| US-004 Forgot/Reset | AuthModule / AuthService | POST /api/auth/forgot-password,reset-password | users, notification_outbox | api-spec | ✅ |
| US-005 Guest boundary | AppShell/ProtectedRoute | (public GET /api/assets*) | — | components | ✅ |
| US-006 Profile/bank | ProfileWallet / ProfileService | GET/PUT /api/profile | users | api-spec | ✅ |
| US-007 Risk assessment | ProfileWallet / ProfileService | GET/POST /api/risk-assessment | risk_assessments | data-model | ✅ |
| US-008 Deposit | ProfileWallet / WalletService | POST /api/wallet/deposit | wallets, wallet_transactions | data-model | ✅ |
| US-009 Withdraw | WalletService | POST /api/wallet/withdraw | wallets, wallet_transactions | api-spec | ✅ |
| US-010 Balances/reserve | WalletService | GET /api/wallet | wallets | correctness P2 | ✅ |
| US-011 Browse/search (+periods,sort) | Market / CatalogService | GET /api/assets | funds,bonds,gold,*_navs/prices | api-spec | ✅ |
| US-012 Asset detail | AssetDetail / CatalogService | GET /api/assets/{t}/{id}(/prices) | fund_navs,bond_prices,gold_prices | data-model | ✅ |
| US-013 Buy MF | Trading / TradingService | POST /api/orders | orders,holdings,transactions,wallets | correctness P1 | ✅ |
| US-014 Sell MF | TradingService | POST /api/orders | orders,holdings,transactions | correctness P3 | ✅ |
| US-015 Buy Bond (100k lot) | TradingService | POST /api/orders | orders,holdings,bonds | correctness P5 | ✅ |
| US-016 Sell Bond | TradingService | POST /api/orders | orders,holdings | — | ✅ |
| US-017 Buy/Sell Gold | TradingService | POST /api/orders/quote,{id}/confirm | orders,holdings,gold_prices | correctness P4 | ✅ |
| US-018 Risk-mismatch ack | RiskAckBox / TradingService | POST /api/orders (riskAck) | audit_logs | components | ✅ |
| US-019 Create DCA | DCA / DCAService | POST /api/dca-plans | dca_plans | data-model | ✅ |
| US-020 Manage DCA | DCA / DCAService | PUT/DELETE/pause/resume, GET rounds | dca_plans, dca_rounds | api-spec | ✅ |
| US-021 DCA auto-exec | DCAService + Scheduler | POST /api/admin/run/dca | dca_rounds, orders | integration | ✅ |
| US-022 Portfolio | Portfolio / PortfolioService | GET /api/portfolio | holdings | correctness P3 | ✅ |
| US-023 Dashboard | Dashboard / DashboardService | GET /api/dashboard?trend= | holdings, transactions, prices | components | ✅ |
| US-024 History list | History / HistoryService | GET /api/transactions | transactions | data-model | ✅ |
| US-025 Filter/search/sort | History / HistoryService | GET /api/transactions (params) | transactions | api-spec | ✅ |
| US-026 Txn detail | TransactionDetail / HistoryService | GET /api/transactions/{ref} | transactions, status_history | data-model | ✅ |
| US-027 Cancel pending | HistoryService | POST /api/transactions/{ref}/cancel | orders, wallets | integration | ✅ |
| US-028 Export Excel/statement | HistoryService | GET /api/transactions/export; /statements/{ym} | transactions | api-spec (excelize) | ✅ |
| US-029 Notifications | NotificationService | GET /api/notifications; settings | notification_outbox | integration | ✅ |
| US-030 Mock + price-gen toggle | MockExternal+PriceGenerator | /api/admin/price-generator; run/* | *_prices, price_generator_state | integration | ✅ |
| US-031 Admin master/price | AdminService | /api/admin/funds|bonds|gold, audit-logs | funds,bonds,gold,audit_logs | api-spec | ✅ |
| US-032 Seed + demo account | db/seed | (seed on startup) | all tables | implementation | ✅ |

### Coverage Summary
- **Requirements covered**: 32 / 32.
- **Gaps**: none.
- **Components without requirement**: HTTP middleware, repositories, money helpers — infrastructure/scaffolding supporting the above (justified).

## Open Questions & Risks
| # | Question/Risk | Impact | Status |
|---|--------------|--------|--------|
| 1 | Go, Docker, MySQL not installed on the machine | High (blocks full-stack run) | Open — see operations.md; FE mock mode runs on Node now |
| 2 | US-015 min lot 100k overrides source BR-04 | Medium (business value) | Mitigated — logged as Add-on per user instruction |
| 3 | Monthly statement PDF (US-028) | Low | Open — minimal/stub in phase 1 |
| 4 | Real scheduler vs manual trigger (D1-3) | Low | Mitigated — manual trigger + seeded data chosen |

## Detailed Specifications
- [Components](design/components.md)
- [Data Model](design/data-model.md)
- [API Specification](design/api-spec.md)
- [Integration](design/integration.md)
- [Implementation](design/implementation.md)
- [Operations](design/operations.md)
- [Correctness Properties](design/correctness.md)
- [Testing Strategy](design/testing-strategy.md)

## External References
| Source | Type | Used in |
|--------|------|---------|
| (MSC) AIDLC Workshop - Usecase WeBond.md | Requirements | all design files |
| images/01..09 *.png | Wireframes | components.md (pages/layout) |
| decisions-design.md (D3) | Tech decisions | all design files |
