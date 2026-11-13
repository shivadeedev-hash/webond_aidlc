# Components

## Overview
Layered architecture on both tiers. **Frontend (React SPA)**: pages → feature components → hooks (TanStack Query) → service interface (mock-JSON impl | HTTP impl, env-selected). **Backend (Go/Gin)**: router/handlers → services (business logic) → repositories (sqlc over MySQL). Money handled with shopspring/decimal; API uses an envelope `{ data, error, meta }`.

---

## Frontend Components

### AppShell / Layout
- **Purpose**: White sticky top nav with red brand accents (Add-on 2026-11-12c; was dark navy) (Dashboard, Market, Portfolio, DCA, History, Profile), auth-aware routing, responsive container.
- **Exposes**: route layout, nav, auth guard (redirect guests per US-005).
- **Consumes**: `useAuth`.
- **Internal**: `components/layout/{Header,Nav,Container,ProtectedRoute}.tsx`.

### AuthModule
- **Purpose**: Login, Register, Forgot/Reset, email verification (FR-01/02/03).
- **Exposes**: `pages/{Login,Register,ForgotPassword,ResetPassword}.tsx`.
- **Consumes**: `authService`, `useAuth` (token in memory + silent refresh).
- **Error handling**: generic invalid-credential message (EX-01), lockout notice, duplicate email (EX-02).

### ProfileWallet
- **Purpose**: Profile + risk assessment + wallet deposit/withdraw + wallet history (FR-04/07).
- **Exposes**: `pages/Profile.tsx`, `components/wallet/*`, `components/risk/RiskQuestionnaire.tsx`.
- **Consumes**: `profileService`, `walletService`.

### Market
- **Purpose**: Catalog with type tabs, search, filters, multi-period return filter + sort by performance (FR-05, US-011 Add-on).
- **Exposes**: `pages/Market.tsx`, `components/market/{AssetCard,FilterBar,PeriodToggle}.tsx`.
- **Consumes**: `assetService`.

### AssetDetail
- **Purpose**: Type-specific detail + NAV/price chart + Buy/Sell panel + DCA shortcut (FR-06, FR-08..12).
- **Exposes**: `pages/AssetDetail.tsx`, `components/trade/{BuySellPanel,QuoteDialog,RiskAckBox}.tsx`, `components/charts/PriceChart.tsx`.
- **Consumes**: `assetService`, `orderService`, `walletService`.
- **Error handling**: insufficient balance/holding, quote-expired re-quote (EX-08), risk-mismatch ack (EX-12), price-unavailable (EX-10).

### DCA
- **Purpose**: Plan list + create/edit form (MF/Gold only) + round history (FR-13/14).
- **Exposes**: `pages/DCA.tsx`, `components/dca/{PlanTable,PlanForm}.tsx`.
- **Consumes**: `dcaService`.

### Portfolio
- **Purpose**: Summary cards + holdings table with P/L and next event (FR-21).
- **Exposes**: `pages/Portfolio.tsx`, `components/portfolio/HoldingsTable.tsx`.
- **Consumes**: `portfolioService`.

### Dashboard
- **Purpose**: Value/P&L cards, allocation pie, value-trend chart (1M/6M/1Y/3Y/5Y/All — Add-on), top movers, DCA summary, upcoming events (FR-22).
- **Exposes**: `pages/Dashboard.tsx`, `components/dashboard/*`, `components/charts/{AllocationPie,TrendChart}.tsx`.
- **Consumes**: `dashboardService`.

### History
- **Purpose**: Dated list (infinite scroll), tabs/filters/search/sort in URL, detail with status timeline, cancel pending, Excel export + monthly statement (FR-16..20, Add-on xlsx).
- **Exposes**: `pages/History.tsx`, `pages/TransactionDetail.tsx`, `components/history/{TxnRow,FilterBar,StatusTimeline}.tsx`.
- **Consumes**: `historyService`.

### Shared FE utilities
- **Purpose**: formatting (THB 2dp, units/grams 4dp, dates Asia/Bangkok dd MMM yyyy HH:mm — BR-15), i18n scaffold (Thai default), design tokens, API client, error-code→message map.
- **Internal**: `lib/{format,i18n,apiClient,errors}.ts`, `types/*`.

### Service Layer (swap point — D3-6)
- **Purpose**: One typed interface per domain with two implementations selected by `VITE_DATA_SOURCE` (`mock` | `api`).
- **Internal**: `services/{auth,asset,wallet,order,dca,portfolio,dashboard,history,notification}/{index.ts, mock.ts, http.ts}`.

---

## Backend Components (Go)

### HTTP Layer (Gin)
- **Purpose**: Routing, middleware (JWT auth, RBAC, request-id, structured logging, recovery, CORS, idempotency), envelope responses.
- **Exposes**: all REST endpoints (see api-spec.md), `GET /health`, `GET /health/ready`.
- **Internal**: `api/internal/http/{router.go, middleware/*, response.go}`.

### AuthService
- **Purpose**: register, verify, login/logout, refresh, reset (FR-01/02/03), bcrypt, JWT, lockout.
- **Consumes**: UserRepo, NotificationService.

### ProfileService / WalletService
- **Purpose**: profile + risk assessment (FR-04); deposit/withdraw, reserve/release, balances (FR-07, BR-11).
- **Consumes**: UserRepo, RiskRepo, WalletRepo (transactional).

### CatalogService
- **Purpose**: list/search/filter assets, period returns + performance sort, asset detail, price/NAV history (FR-05/06, Add-on).
- **Consumes**: FundRepo, BondRepo, GoldRepo, PriceRepo.

### TradingService
- **Purpose**: Buy/Sell MF, Bond, Gold with all BR/EX rules; idempotency; quote lock; risk-mismatch ack; order lifecycle; holdings update (avg cost); writes transactions + status history (FR-08..12, BR-02/03/04→Add-on/09/10, EX-04..08/11/12).
- **Consumes**: OrderRepo, HoldingRepo, WalletRepo, PriceRepo, TransactionRepo (single DB transaction per order — NFR-05).

### DCAService + Scheduler
- **Purpose**: create/manage plans (MF/Gold, BR-06/07); compute next_run; execute due plans (manual trigger endpoint in phase 1 — D1-3); idempotent rounds (FR-15); pause after 3 fails (EX-09).
- **Consumes**: DcaRepo, TradingService.

### PortfolioService / DashboardService
- **Purpose**: holdings valuation + P/L (BR-10); dashboard aggregates, allocation, trend series, top movers, upcoming events (FR-21/22).
- **Consumes**: HoldingRepo, PriceRepo, TransactionRepo, DcaRepo, BondRepo.

### HistoryService
- **Purpose**: list/filter/search/sort/paginate transactions, detail (+ownership check EX-15), cancel pending (BR-13/EX-13), Excel export via excelize, monthly statement stub (FR-16..20).
- **Consumes**: TransactionRepo, OrderRepo.

### BondLifecycleService
- **Purpose**: coupon crediting + maturity closeout (FR-23, BR-05) — manual trigger in phase 1.
- **Consumes**: BondRepo, HoldingRepo, WalletRepo, TransactionRepo.

### MockExternalService + PriceGenerator
- **Purpose**: mock endpoints backed by MySQL (FR-26); auto price-generator goroutine with persisted on/off toggle + start/stop + manual publish/tick/pay-coupon (US-030 Add-on).
- **Consumes**: PriceRepo, PriceGeneratorStateRepo, BondLifecycleService.

### NotificationService
- **Purpose**: write in-app + simulated email rows to notification_outbox incl. OTP/verify/reset (FR-24, honors email on/off).
- **Consumes**: NotificationRepo.

### AdminService
- **Purpose**: minimal master-data CRUD + audit (FR-25, RBAC admin) — phase-1 scope-limited.
- **Consumes**: Fund/Bond/Gold repos, AuditRepo.

### Repositories (sqlc)
- **Purpose**: type-safe queries; money as decimal; all order/balance mutations inside DB transactions.
- **Internal**: `api/internal/repo/*` + `api/db/queries/*.sql` + sqlc-generated `api/internal/db/*`.

---

## Interactions (ASCII)
```
Browser (React SPA)
  │  HTTPS, Bearer access token (memory), refresh cookie (httpOnly)
  ▼
Gin router → middleware (auth/RBAC/req-id/log/idempotency)
  ▼
Handlers → Services (business logic, decimal math, DB txn)
  ▼
Repositories (sqlc) → MySQL 8.0
         ▲
PriceGenerator goroutine ──(ticker, if enabled)── writes prices → MySQL
```

### Data flow (place MF buy)
```
BuySellPanel → orderService.buyFund(amount, idemKey)
  → POST /api/orders (envelope)
  → TradingService: check KYC/suitability (BR-01), min purchase, balance
     → reserve cash (WalletRepo, txn) → create order 'pending'
     → cut-off rule (BR-02) sets trade_date/NAV date
     → on NAV confirm: units = amount/NAV (4dp), holding avg-cost update,
       transaction + status history, release/settle reserve
  → envelope { data: order }  → UI invalidates queries (TanStack Query)
```

## Component → Story map (summary)
Auth→US-001..005; ProfileWallet→US-006..010; Market→US-011..012; AssetDetail/Trading→US-013..018; DCA→US-019..021; Portfolio→US-022; Dashboard→US-023; History→US-024..028; Notification→US-029; MockExternal/PriceGenerator→US-030; Admin→US-031; Seed→US-032. (Full matrix in design.md Traceability.)
