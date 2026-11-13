# Design Decisions (D3)

## Context Summary
WeBond — full-stack: React 18 + TS (Vite) + Tailwind frontend, Go (Gin) REST API, MySQL 8.0 (seeded). Comprehensive mode, scope=new, solo dev. 32 stories, 3 user types. Stack is already fixed by requirements + user, so D3 focuses on the open architecture/library/pattern choices within that stack (not re-choosing the stack). Key features: auth (bcrypt+JWT), trading (MF/Bond/Gold) with DECIMAL money, DCA, portfolio/dashboard with charts, history + Excel export, mock external systems with an auto price-generator toggle.

---

## Decision Questions

### D3-1: Go data access layer
**Question**: How should the Go backend talk to MySQL?
- 1) **sqlc** — write SQL, generate type-safe Go; precise control over DECIMAL and transactions **(Recommended for financial accuracy + explicit queries)**
- 2) GORM — ORM, faster CRUD, less boilerplate but less control over generated SQL
- 3) database/sql + sqlx — thin, manual mapping
- 4) Other: _______

**Answer**: sqlc

---

### D3-2: Monetary/decimal handling in Go
**Question**: How to represent money & units (must be DECIMAL, never float — NFR-05)?
- 1) **shopspring/decimal** in Go mapped to MySQL DECIMAL(18,4); format at the edge **(Recommended)**
- 2) Integer minor units (store satang/4-dp scaled ints)
- 3) string passthrough from DB, parse only when computing
- 4) Other: _______

**Answer**: shopspring/decimal

---

### D3-3: API response & error format
**Question**: What response envelope and error shape should the REST API use?
- 1) **Envelope `{ data, error, meta }`** with typed error `{ code, message }` using the spec's codes (INSUFFICIENT_BALANCE, QUOTE_EXPIRED, BOND_OFFER_INVALID, ...) **(Recommended — matches exception spec)**
- 2) Bare JSON body + HTTP status only
- 3) RFC 7807 problem+json
- 4) Other: _______

**Answer**: Envelope

---

### D3-4: Auth token storage on the client
**Question**: Where does the React app keep the JWT access/refresh tokens?
- 1) **Access token in memory + refresh token in httpOnly secure cookie**; silent refresh on 401 **(Recommended — best XSS/CSRF posture)**
- 2) Both tokens in localStorage (simplest, weaker against XSS)
- 3) Both in httpOnly cookies + CSRF token
- 4) Other: _______

**Answer**: *Access token in memory + refresh token in httpOnly secure cookie**; silent refresh on 401 **(Recommended — best XSS/CSRF posture)

---

### D3-5: Frontend data-fetching & state
**Question**: How does the React app manage server data and caching? (tech.md lists TanStack Query)
- 1) **TanStack Query** for server state + React context for light UI state **(Recommended)**
- 2) Redux Toolkit (incl. RTK Query)
- 3) Plain fetch + useState/useEffect
- 4) Other: _______

**Answer**: TanStack Query

---

### D3-6: Frontend data-layer swap strategy (mock → API)
**Question**: Workshop order builds the UI on mock JSON first, then wires the Go API. How to structure that swap cleanly?
- 1) **Single typed service interface with two implementations (mock JSON / HTTP client), chosen by an env flag** so UI code never changes **(Recommended)**
- 2) Build directly against the API from the start (skip the mock-JSON step)
- 3) Mock via MSW (Mock Service Worker) intercepting fetch
- 4) Other: _______

**Answer**: Single typed service interface with two implementations (mock JSON / HTTP client), chosen by an env flag

---

### D3-7: Auto price-generator implementation (US-030)
**Question**: How to run the simulated realtime price generator with on/off toggle?
- 1) **In-process Go goroutine ticker writing to MySQL, controlled by a persisted toggle row + start/stop endpoints; frontend polls latest prices** **(Recommended — simple, no extra infra)**
- 2) Goroutine + push updates over WebSocket/SSE to the client
- 3) Separate worker process / cron container
- 4) Other: _______

**Answer**: In-process Go goroutine ticker writing to MySQL, controlled by a persisted toggle row + start/stop endpoints; frontend polls latest prices

---

### D3-8: Excel export generation (US-028)
**Question**: Where/how to produce the .xlsx history export?
- 1) **Backend generates .xlsx with a Go library (excelize) and streams the file** **(Recommended — handles Thai text, large data server-side)**
- 2) Frontend generates .xlsx in-browser (SheetJS)
- 3) Backend returns data, frontend builds file
- 4) Other: _______

**Answer**: Backend generates .xlsx with a Go library (excelize) and streams the file

---

### D3-9: Charts on the frontend
**Question**: Which charting approach for dashboard pie + value-trend + NAV/price history?
- 1) **Recharts** (already in tech.md) — React-native, declarative **(Recommended)**
- 2) Chart.js (react-chartjs-2)
- 3) ECharts
- 4) Other: _______

**Answer**: Recharts

---

### D3-10: Local orchestration
**Question**: How to run the full stack locally for the reviewer?
- 1) **Docker Compose (web + api + mysql), seed on startup** **(Recommended — one command, reproducible, matches target)**
- 2) Run each service manually (node, go, local mysql)
- 3) Compose for mysql only; run web/api on host
- 4) Other: _______

**Answer**: Docker Compose (web + api + mysql), seed on startup

---

### D3-11: Correctness & Property-Based Testing (MANDATORY)
**Question**: Which financial logic needs property-based / invariant testing beyond example tests?
- 1) **Core money math** — unit allotment (amount÷NAV), average-cost & P/L, cash reserve/release conservation, gold grams↔baht-weight, bond coupon math: property tests asserting invariants (no negative balances, reserved ≤ balance, totals conserved) **(Recommended)**
- 2) Example-based unit tests only
- 3) Full property-based coverage across all services
- 4) Other: _______

**Answer**: Core money math

---

### D3-12: Observability Strategy (MANDATORY)
**Question**: What level of observability does this service need?
- 1) **Minimal — structured logging + health endpoint** (fits solo/hands-on build) **(Recommended)**
- 2) Standard — logging + metrics + health + readiness
- 3) Full — logging + metrics + tracing + alerting + dashboards
- 4) None (prototype only)
- 5) Other: _______

**Answer**: Minimal

---

### D3-13: Error Tracking (MANDATORY)
**Question**: How should runtime errors be captured?
- 1) **Log-based only — errors in structured logs** **(Recommended for this build)**
- 2) Dedicated service (Sentry/Datadog/Rollbar)
- 3) Cloud-native (CloudWatch/GCP/Azure)
- 4) None
- 5) Other: _______

**Answer**: Log-based only

---

### D3-14: Health & Lifecycle Management (MANDATORY)
**Question**: What lifecycle management does the Go service need?
- 1) Basic health endpoint only
- 2) **Health + readiness + graceful shutdown** **(Recommended — containerized via Compose)**
- 3) Health + readiness + graceful shutdown + startup probe + drain delay
- 4) Other: _______

**Answer**: Health + readiness + graceful shutdown

---

## Decisions Summary
<!-- Populated after you answer. One line per decision. -->
- D3-1 Go data access: sqlc (type-safe SQL)
- D3-2 Decimal handling: shopspring/decimal ↔ MySQL DECIMAL(18,4)
- D3-3 API/error format: Envelope { data, error, meta } with typed { code, message } (spec error codes)
- D3-4 Token storage: Access token in memory + refresh token in httpOnly secure cookie; silent refresh on 401
- D3-5 FE data/state: TanStack Query + React context
- D3-6 Mock→API swap: Single typed service interface, mock-JSON vs HTTP impl chosen by env flag
- D3-7 Price generator: In-process Go goroutine ticker → MySQL, persisted toggle + start/stop endpoints, FE polls
- D3-8 Excel export: Backend generates .xlsx via excelize, streams file
- D3-9 Charts: Recharts
- D3-10 Local orchestration: Docker Compose (web + api + mysql), seed on startup
- D3-11 Correctness/PBT: Core money math invariants (allotment, avg-cost/P&L, cash conservation, gold units, coupon)
- D3-12 Observability: Minimal (structured logging + health endpoint)
- D3-13 Error tracking: Log-based only
- D3-14 Health/lifecycle: Health + readiness + graceful shutdown

---

## Environment Note (machine check 2026-10-28)
Detected: Node 24.14.1 ✅, npm 11.11.0 ✅, git 2.54.0 ✅. Missing: Go ❌, Docker ❌, MySQL ❌. Frontend mock mode runs on Node alone; full stack needs Go + Docker (see design/operations.md § Local Environment Setup).

---

## Validation Notes
- All answers = recommended options; mutually consistent. No conflicts.
- Checked: observability=Minimal fits solo/hands-on (OK, not production-critical); error-tracking=Log-based pairs with Minimal (OK); lifecycle=Health+readiness+graceful matches Docker Compose container target (OK); sqlc + shopspring/decimal + DECIMAL(18,4) satisfies NFR-05 financial accuracy; token storage (memory + httpOnly cookie) satisfies NFR-01/OWASP XSS/CSRF posture.

---

**Instructions**: Fill in your answers above and respond with "done" — or say "use recommendations" to auto-fill all recommended options.
