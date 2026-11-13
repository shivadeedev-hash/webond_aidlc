# Tasks — WeBond

## Summary
- **Total Tasks**: 46 across 12 phases (incl. Add-ons: W3.5 Design Documentation, W3.6 Production readiness)
- **Execution Waves**: 7 (sequential by wave per D4-7; stop-and-verify at each wave end)
- **Coverage**: ~13 backend services, 20 entities, REST endpoints under `/api`, 9 UI pages
- **Testing**: TDD per wave for critical Go logic (money/trading/DCA/wallet); test-after for UI; property tests for money math
- **Strategy (D4)**: wave-based breakdown, mock-JSON→API swap, single-concern tasks, sequential
- **Derived from**: requirements.md (US-001..032), design.md + design/*, decisions-tasks.md

## Overview
Waves follow the hands-on policy order: **W0 Setup → W1 Frontend on mock JSON → W2 Backend + MySQL + swap → W3 Verify → W3.5 Design Documentation (Add-on) → W4 Deploy (optional)**. TDD applies within waves that contain critical backend logic (write tests first). UI is test-after. Each wave ends with a verify checkpoint; stop and wait for user confirmation before the next wave.

Legend: `- [ ] N. Phase` / `  - [ ] N.M Task`. **Deps** = prerequisite task IDs. **Ref** = design source.

---

## Task Phases

- [x] 1. Wave 0 — Project Setup & Environment
  - [x] 1.1 Initialize repo & root scaffolding
    - **Deps**: none | **Ref**: implementation.md — directory structure
    - `git init -b main` at project root; add root `.gitignore` (node_modules, dist, .env, /api/bin, *.log, .DS_Store), `README.md` (overview + run instructions placeholder), `.env.example`.
  - [x] 1.2 Environment install script (`scripts/setup-env.sh`)
    - **Deps**: 1.1 | **Ref**: operations.md — Local Environment Setup; decisions-tasks.md D4-8
    - Bash script: detect OS; check Node/npm/git/Go/Docker/MySQL presence + versions; install missing **real-build** tools via Homebrew (`go`, `--cask docker`, optional `mysql-client`); print a clear summary table; idempotent (skip installed); never fail the dev flow. Note in output: "development uses `npm run dev` (mock mode); Docker/Go only needed for the real/backend build." Make executable.
  - [x] 1.3 Scaffold frontend app (Vite + React + TS + Tailwind)
    - **Deps**: 1.1 | **Ref**: implementation.md; tech.md
    - `web/` Vite React-TS app; install Tailwind, React Router, TanStack Query, Recharts; `web/.env.example` (`VITE_DATA_SOURCE=mock`, `VITE_API_URL`); ESLint + Prettier; `npm run dev` works (blank shell) on Node now.
  - [x] 1.4 Design tokens & Tailwind theme
    - **Deps**: 1.3 | **Ref**: design-system-policies; components.md (Shared FE)
    - `src/styles/tokens.css` CSS variables (white bg, dark-navy header, black text, navy-blue primary button, semantic colors); map into `tailwind.config.js`; create `.kiro/steering/design-system.md` recording token decisions + rationale (per design-system policy). No hardcoded colors in components.
  - [x] 1.5 Scaffold backend module (Go + Gin) + Docker Compose
    - **Deps**: 1.1 | **Ref**: implementation.md; integration.md
    - `api/` Go module, Gin server skeleton with `/health`; `cmd/webond/main.go`; `sqlc.yaml`; empty `db/{migrations,queries,seed}`; root `docker-compose.yml` (web, api, mysql) + `api/Dockerfile`, `web/Dockerfile`. (Builds only after user installs Go/Docker — dev continues via npm.)
  - **Verify W0**: `cd web && npm run dev` serves on :5173; `scripts/setup-env.sh` runs and reports; repo committed.

- [ ] 2. Wave 1a — Frontend foundation (types, services, mock data)
  - [x] 2.1 Shared types & formatting utils (TDD)
    - **Deps**: 1.3 | **Ref**: components.md (Shared FE); data-model.md; BR-15
    - `src/types/*` domain models; `src/lib/format.ts` (THB 2dp, units/grams 4dp, dates Asia/Bangkok). Write `format.test.ts` first (test-after allowed for pure UI, but formatting is logic → tests first).
  - [x] 2.2 Typed service interfaces + env-selected impl wiring
    - **Deps**: 2.1 | **Ref**: components.md (Service Layer); D3-6/D4-4
    - Per-domain `services/<d>/{index.ts,mock.ts,http.ts}`; `index` picks impl via `VITE_DATA_SOURCE`; `lib/apiClient.ts` (envelope + error-code map, silent refresh stub); `lib/errors.ts`.
  - [x] 2.3 Mock JSON seed data
    - **Deps**: 2.1 | **Ref**: requirements US-032; data-model.md
    - `src/mocks/*.json`: catalog (funds/bonds/gold + price/NAV history), one demo investor with wallet, holdings (MF/Bond/Gold), DCA plans, ~3 months transactions, notifications. `mock.ts` reads these. Mirrors eventual API shapes.
  - [x] 2.4 App shell, routing, i18n scaffold
    - **Deps**: 1.4, 2.2 | **Ref**: components.md (AppShell); NFR-07
    - Router, dark-navy Header/Nav (Dashboard/Market/Portfolio/DCA/History/Profile), responsive container, `ProtectedRoute` (guest redirect US-005), `useAuth` (token in memory), i18n scaffold (Thai default).

- [ ] 3. Wave 1b — Frontend pages on mock data
  - [x] 3.1 Auth pages (Login/Register/Forgot/Reset/Verify)
    - **Deps**: 2.4 | **Ref**: components.md (AuthModule); US-001..005; wireframe 01
    - Forms + validation, consent checkbox, generic error, lockout/duplicate messaging; verification via mock inbox.
  - [x] 3.2 Profile & Wallet page
    - **Deps**: 2.4 | **Ref**: US-006..010; wireframe 09
    - Personal info + bank (masked), risk questionnaire → risk level, wallet balances (balance/reserved/available), deposit/withdraw, wallet history.
  - [x] 3.3 Market page (catalog, filters, period return + sort)
    - **Deps**: 2.4 | **Ref**: US-011..012; wireframe 03; Add-on
    - Type tabs, search, filters, period toggle 3M/6M/1Y/3Y/5Y, sort by performance, asset cards.
  - [x] 3.4 Asset Detail + Buy/Sell panel
    - **Deps**: 3.3 | **Ref**: US-012..018; wireframe 04
    - Type-specific detail, price/NAV chart (Recharts), Buy/Sell panel, quote dialog (gold 30s), risk-ack box, "Set up DCA" shortcut. Error states (insufficient/quote-expired/price-unavailable).
  - [x] 3.5 DCA page
    - **Deps**: 2.4 | **Ref**: US-019..021; wireframe 05
    - Plan table (status/invested/next run), create/edit form (MF/Gold only), round history; min-amount + bond-exclusion validation.
  - [x] 3.6 Portfolio page
    - **Deps**: 2.4 | **Ref**: US-022; wireframe 06
    - Summary cards, holdings table (qty, avg cost, market value, unrealized P/L, next event), bond accrued/coupon/maturity row.
  - [x] 3.7 Dashboard page
    - **Deps**: 3.6 | **Ref**: US-023; wireframe 02; Add-on
    - Value/P&L/cash cards, allocation pie, value-trend chart (1M/6M/1Y/3Y/5Y/All), top movers, DCA summary, upcoming events.
  - [x] 3.8 History list + Transaction detail
    - **Deps**: 2.4 | **Ref**: US-024..028; wireframes 07/08
    - Dated infinite-scroll list, tabs/filters/search/sort in URL, signed amounts + badges, detail page with status timeline, cancel-pending button (enabled per rule), Export (Excel) + statement buttons (wired to API later).
  - [x] 3.9 FE component tests on core flows
    - **Deps**: 3.1–3.8 | **Ref**: testing-strategy.md
    - Vitest + RTL: Market render/filter, AssetDetail buy flow, History filter/empty-state, DCA form validation.
  - **Verify W1**: all 9 pages usable in mock mode; `npm run test` green; responsive check (mobile/desktop). STOP for user confirmation.

- [x] 4. Wave 2a — Backend foundation (DB, config, money, middleware)
  - [x] 4.1 Config, logging, health, graceful shutdown
    - **Deps**: 1.5 | **Ref**: operations.md
    - Env config + startup validation, `log/slog` structured logging + request-id middleware, `/health` + `/health/ready` (DB ping), graceful shutdown.
  - [x] 4.2 MySQL migrations (all 20 entities)
    - **Deps**: 1.5 | **Ref**: data-model.md
    - golang-migrate up/down SQL for all tables, DECIMAL(18,4), indexes, enums, FKs.
  - [x] 4.3 sqlc setup + base queries + repositories
    - **Deps**: 4.2 | **Ref**: data-model.md; components.md (Repositories)
    - `db/queries/*.sql`, generate sqlc, repo interfaces + impl; connection pool; transaction helper.
  - [x] 4.4 Money package (TDD + property tests)
    - **Deps**: 1.5 | **Ref**: correctness.md P1,P4,P5,P6; D4-2
    - shopspring/decimal helpers: allotment, avg-cost, gold grams↔baht-weight, coupon, sign rules. Write gopter property tests FIRST, then implement.
  - [x] 4.5 HTTP envelope, error codes, auth middleware
    - **Deps**: 4.1 | **Ref**: api-spec.md
    - Response envelope + typed error codes, JWT access/refresh, RBAC, idempotency middleware, CORS.

- [x] 5. Wave 2b — Backend services (TDD for business logic)
  - [x] 5.1 AuthService + endpoints (TDD)
    - **Deps**: 4.3, 4.5 | **Ref**: api-spec Auth; US-001..005; EX-01/02/03
    - bcrypt, register/verify/login/refresh/logout/reset, lockout, notification_outbox entries. Tests first.
  - [x] 5.2 Profile & Wallet services + endpoints (TDD)
    - **Deps**: 4.3, 4.4 | **Ref**: US-006..010; BR-11
    - Profile + risk assessment; deposit/withdraw/reserve/release in DB transactions; balances. Tests first (cash conservation P2).
  - [x] 5.3 CatalogService + endpoints
    - **Deps**: 4.3 | **Ref**: US-011..012; Add-on
    - List/search/filter, period returns + performance sort, detail, price/NAV history.
  - [x] 5.4 TradingService + order endpoints (TDD)
    - **Deps**: 4.4, 5.2, 5.3 | **Ref**: US-013..018; BR-02/03/04→Add-on/09/10; EX-04..08/11/12
    - Buy/Sell MF/Bond/Gold, cut-off, gold quote lock, idempotency, risk-ack, holdings avg-cost update, transactions + status history, single DB txn. Tests first (P3,P5).
  - [x] 5.5 DCAService + scheduler + manual trigger (TDD)
    - **Deps**: 5.4 | **Ref**: US-019..021; FR-15; EX-09
    - CRUD plans (MF/Gold), next-run calc, idempotent rounds, pause-after-3, `/api/admin/run/dca`. Tests first.
  - [x] 5.6 Portfolio & Dashboard services + endpoints
    - **Deps**: 5.4 | **Ref**: US-022..023; BR-10
    - Holdings valuation + P/L, dashboard aggregates/allocation/trend/top-movers/events.
  - [x] 5.7 HistoryService + endpoints + Excel export
    - **Deps**: 5.4 | **Ref**: US-024..028; BR-12/13/14/15; Add-on xlsx
    - List/filter/search/sort/paginate, detail (+ownership), cancel pending, excelize `.xlsx` export, monthly statement stub.
  - [x] 5.8 BondLifecycle + Notification + Admin + manual triggers
    - **Deps**: 5.4 | **Ref**: US-029..031; FR-23/24/25
    - Coupon/maturity processing (`/run/coupons`, `/run/maturities`), notifications + email on/off, minimal admin master-data CRUD + audit.
  - [x] 5.9 MockExternal + auto price-generator + toggle
    - **Deps**: 4.3 | **Ref**: US-030 Add-on; integration.md
    - Goroutine ticker writing prices to MySQL, persisted toggle (`price_generator_state`), `/api/admin/price-generator` on/off + publish-nav/tick-gold.

- [x] 6. Wave 2c — Seed, integration tests, data-layer swap
  - [x] 6.1 Seed data (rich + demo investor)
    - **Deps**: 4.2 | **Ref**: US-032; D1-5
    - `db/seed/seed.sql`: full catalog + price/NAV history, demo investor (wallet, holdings, DCA, ~3mo transactions), admin; run on compose startup. Print demo credentials to README.
  - [x] 6.2 Backend integration tests (service + MySQL)
    - **Deps**: 5.1–5.9, 6.1 | **Ref**: testing-strategy.md
    - Build-tagged tests against a MySQL test DB for auth, orders (+idempotency/quote), wallet, DCA, history/cancel/export.
  - [x] 6.3 Wire frontend HTTP impl + flip to API
    - **Deps**: 5.1–5.9, W1 | **Ref**: D3-6/D4-4
    - Implement `services/*/http.ts` against real endpoints; silent refresh; set `VITE_DATA_SOURCE=api`; keep mock for tests/offline.
  - **Verify W2**: `docker compose up` brings web+api+mysql; migrations+seed run; demo login works; core flows work against real API; `go test ./...` + integration green. STOP for user confirmation.

- [x] 7. Wave 3 — Full verification
  - [x] 7.1 End-to-end flow verification (manual + scripted checks)
    - **Deps**: 6.3 | **Ref**: testing-strategy.md critical paths
    - Register→verify→risk→deposit→buy MF→Portfolio/History; gold DCA→run→round; cancel pending before cut-off→cash released.
  - [x] 7.2 Coverage + quality gate
    - **Deps**: 6.2 | **Ref**: NFR-08
    - Backend business-logic coverage ≥ 70%; `go vet`/lint + FE lint clean; fix gaps.
  - [x] 7.3 Responsive + accessibility pass
    - **Deps**: W1 | **Ref**: design-system-policies §9; NFR-07
    - Verify mobile/tablet/desktop, focus-visible, tap targets ≥44px, contrast AA on key screens.
  - **Verify W3**: all checks pass; no console/terminal errors. STOP for user confirmation.

- [ ] 8. Wave 3.5 — Design Documentation (Add-on)
  - [ ] 8.1 Consolidated design document + diagrams (single file)
    - **Deps**: W3 | **Ref**: design-conventions steering; design.md + design/* (components, data-model, api-spec, integration, implementation, operations, correctness, testing-strategy); requirements.md; personas.md
    - **Add-on (user request)**: produce ONE human-readable design document `docs/design-document.md` that lets a reader understand the whole system, with a complete standard diagram set. Source of truth = existing spec/design/* — summarize and visualize, do NOT invent content (guardrails). All diagrams as **Mermaid** embedded in the markdown (version-controlled, no external tools, no licensed assets).
    - **Sections**: (1) Overview & context; (2) Technical architecture (layered FE/BE + infra) with **C4-style context + container diagrams**; (3) **Component diagram**; (4) **Deployment diagram** (Docker Compose web/api/mysql; AWS target); (5) Data flow diagrams; (6) **BPMN 2.0 business process with swimlanes** for key processes (buy order, DCA run, coupon/maturity) per design-conventions; (7) **Sequence diagrams** for core flows (register→verify, place MF buy, gold quote 30s, DCA execution, cancel pending); (8) **User journey** (Investor/Guest); (9) Interfaces / **API design** summary from api-spec.md; (10) Data model + **ER diagram** (20 entities) from data-model.md; (11) Error handling (typed codes/EX map); (12) Testing strategy; (13) Deployment procedures; (14) Implementation considerations / NFRs. Include a traceability note back to requirements.
    - **Verify**: all Mermaid blocks parse (valid syntax), every diagram type above present, content matches design/* (no fabricated facts), single file renders as a coherent document.
  - **Verify W3.5**: `docs/design-document.md` exists, complete section + diagram set, diagrams valid, no invented content. STOP for user confirmation.

- [ ] 9. Wave 4 — Deploy (optional)
  - [ ] 9.1 Production build & container hardening
    - **Deps**: W3.5 | **Ref**: tech.md infra
    - Multi-stage Dockerfiles, prod env config, static web build.
  - [ ] 9.2 CI + deploy config (optional)
    - **Deps**: 9.1 | **Ref**: tech.md CI/CD
    - GitHub Actions (lint/test/build); AWS deploy notes. Optional per policy.
    - **Add-on**: 9.1/9.2 are delivered by phase 10 (tasks 10.4 web image, 10.6 AWS deploy + CI).

- [x] 10. Wave 3.6 — Production readiness (Add-on, user request 2026-11-12)
  - [x] 10.1 API operability: auto-migration + deploy-friendly config
    - **Deps**: W3 | **Ref**: operations.md; 12-factor (III Config, XII Admin processes)
    - Embed `api/db/migrations/*.sql` and run them with golang-migrate on boot (`MIGRATE_ON_BOOT`, default true when APP_ENV != production, false in production); add `webond migrate` subcommand (run pending migrations then exit; needs only DB config) for one-off deploy tasks; DB config from `DATABASE_URL` OR parts `DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD`; optional `DB_TLS_CA_FILE` (RDS CA); `CORS_ALLOWED_ORIGINS` env (default localhost:5173 in non-prod, none in prod); in production reject weak/default JWT secrets (`change-me-dev` or < 32 chars). Must be compatible with a DB already at schema_migrations version 9.
  - [x] 10.2 Bond valuation and sell-price correctness (money bug)
    - **Deps**: W2b | **Ref**: requirements US-016/US-022, BR-10; data-model bond_prices
    - Evidence (seed + schedules): `clean_price` is % of par (100.5), `accrued_interest` is THB per unit (8.125 = half of 16.25 coupon/unit on face 1,000), units = amount / face_value. Fix portfolio/dashboard valuation and bond sell proceeds to use THB-per-unit = face_value × clean_price / 100 (+ accrued per unit on sell quote). Add regression tests. No schema change.
  - [x] 10.3 Monthly statement as real PDF (US-028)
    - **Deps**: W2b | **Ref**: US-028, api-spec `/api/statements/{yyyyMM}`
    - Replace the text stub with a real PDF (`application/pdf`, attachment) using an MIT-licensed Go PDF library; English labels + asset codes (no Thai text, to avoid unshaped glyphs); account shown by email/id.
  - [x] 10.4 Web: wire History Export/Statement + web image in real API mode
    - **Deps**: 10.3 | **Ref**: US-025/028; D3-6
    - Enable Export (Excel) and Statement (PDF, month picker) buttons in API mode (authenticated blob download; disabled with tooltip in mock mode). `web/Dockerfile` becomes a static nginx image taking `ARG VITE_DATA_SOURCE`/`VITE_API_URL` at build time; `docker-compose.yml` passes build args so `docker compose up` yields a web UI that talks to the API.
  - [x] 10.5 One-command local bootstrap
    - **Deps**: 10.1, 10.4 | **Ref**: scripts/setup-env.sh
    - `scripts/gen-env.sh` (creates `.env` with random JWT secrets; `--rotate` replaces weak ones) and `scripts/dev-up.sh` (checks Docker, ensures `.env`, `docker compose up -d --build`, waits for readiness, prints URLs + demo credentials).
  - [x] 10.6 AWS deployment scripts + CI (prepared, NOT executed)
    - **Deps**: 10.1, 10.4 | **Ref**: tech.md infra (ECS Fargate, RDS MySQL, S3+CloudFront); well-architected + ASA security rules
    - `deploy/aws/` CloudFormation template + `deploy.sh`/`destroy.sh` (dry-run by default; explicit apply flag), hardened `api/Dockerfile`, `.github/workflows/{ci,deploy}.yml`, `docs/DEPLOYMENT.md` (Thai, beginner). Validated with static tooling only (cfn-lint, shellcheck/bash -n, dry-run, docker build); never run against a real AWS account without the user.
  - [x] 10.7 Integrated verification + docs
    - **Deps**: 10.1–10.6 | **Ref**: Definition of Done
    - Full stack via `docker compose` in real API mode end-to-end, update `docs/GETTING-STARTED.md` and README, remove the manual migration step, record results in audit.
  - [x] 10.8 Real email delivery over SMTP, 100% free (Add-on, user request 2026-11-12)
    - **Deps**: 10.7 | **Ref**: FR-01, FR-24, FR-26; US-001/002/004/029; requirements Add-on 2026-11-12
    - (Superseded in part by 10.9: e-mails carry an OTP, not a link.)
    - Bug: registering via the web UI dead-ends (verify page reads the auth-only inbox). Fix: migration 000010 adds `sent_at/send_attempts/last_error` to `notification_outbox` (existing email rows marked sent); `internal/mail` (stdlib `net/smtp`, implicit TLS / STARTTLS / plain-for-local, no new Go deps) + outbox dispatcher goroutine (poll, send, mark sent, bounded retries, graceful stop); config `SMTP_HOST/PORT/USERNAME/PASSWORD/FROM/TLS`, `APP_PUBLIC_URL`; no SMTP host → log mode (recipient + subject only, never tokens). Emails include links; `VerifyEmail` reads `?token=` and auto-submits; register/forgot return the token only when not production. Compose adds pinned Mailpit (`:8025` UI). AWS: optional `SmtpHost/SmtpPort/SmtpUsername/SmtpFrom` params, SMTP password in Secrets Manager, API egress on the SMTP port, prod Rule requires SmtpHost. Docs: Gmail app-password how-to (free). Verify: unit tests (fake sender), integration, browser register → Mailpit → click link → login.
  - [x] 10.9 OTP instead of links for verification and password reset (Add-on, user CR 2026-11-12b)
    - **Deps**: 10.8 | **Ref**: FR-01/FR-03, US-001/002/004; requirements Add-on 2026-11-12b; api-spec Auth
    - 6-digit numeric OTP (crypto/rand), HMAC-SHA256 stored (key derived from the JWT refresh secret, bound to user + purpose), constant-time compare, 15 min TTL, 5 attempts (`token_attempts`, migration 000011), new code invalidates older ones, resend cooldown 60 s; plain code removed from `notification_outbox.body` after sending. API: verify {email,code}, reset {email,code,newPassword}, new resend-verification {email}; no enumeration. Web: VerifyEmail (email + 6-digit code + resend with countdown), ResetPassword (email + code + new password), Forgot → Reset carries the e-mail; login of a pending account links to verification. Docs updated. Verify: unit/integration (wrong code, attempts, expiry, invalidation, cooldown, prod no echo), web tests, browser flow via Mailpit.
  - [x] 10.10 Red and white theme, modern and friendly UI (Add-on, user final request 2026-11-12c)
    - **Deps**: 10.9 | **Ref**: design-system-policies; `.kiro/steering/design-system.md`; requirements Add-on 2026-11-12c
    - Re-value tokens only (names stable) in `web/src/styles/tokens.css` + add `primary-soft`, `header-text`, `gradient-brand` (synced in tailwind.config.js + design-system.md). White sticky header with red logo and soft-red active nav; pill buttons; friendlier radii/shadows; red gradient auth panel and dashboard welcome; danger = brick red distinct from brand. Verify: AA contrast, all pages desktop + mobile screenshots, web tests, no emoji, no hardcoded hex.
  - [x] 10.11 Order entry: sell by amount or units + easier amount input (Add-on, user report 2026-11-12d)
    - **Deps**: 10.10 | **Ref**: US-013..017, requirements Add-on 2026-11-12d
    - API: fund/bond sell accept `orderMode` amount (THB → units at NAV / dirty price) or units; clear INSUFFICIENT_HOLDING message with the maximum. Web: text amount field (commas, ฿, spaces OK; no minus), THB ↔ units/grams switch for sell, holding shown, quick chips, specific inline validation. Also the DCA plan form amount field (same input component; BR-07 minimums and messages unchanged). Verify: service tests (sell by amount, over-holding), web tests (parsing, chips, messages), browser buy + sell for fund, bond and gold.

---

## Task Summary

| ID | Title | Deps | Status |
|----|-------|------|--------|
| 1.1 | Init repo & scaffolding | — | ☐ |
| 1.2 | Env install script | 1.1 | ☐ |
| 1.3 | Scaffold frontend | 1.1 | ☐ |
| 1.4 | Design tokens & theme | 1.3 | ☐ |
| 1.5 | Scaffold backend + compose | 1.1 | ☐ |
| 2.1 | Types & formatting (TDD) | 1.3 | ☑ |
| 2.2 | Service interfaces + env swap | 2.1 | ☑ |
| 2.3 | Mock JSON seed | 2.1 | ☑ |
| 2.4 | App shell, routing, i18n | 1.4, 2.2 | ☑ |
| 3.1 | Auth pages | 2.4 | ☑ |
| 3.2 | Profile & Wallet page | 2.4 | ☑ |
| 3.3 | Market page | 2.4 | ☑ |
| 3.4 | Asset Detail + Buy/Sell | 3.3 | ☑ |
| 3.5 | DCA page | 2.4 | ☑ |
| 3.6 | Portfolio page | 2.4 | ☑ |
| 3.7 | Dashboard page | 3.6 | ☑ |
| 3.8 | History + detail | 2.4 | ☑ |
| 3.9 | FE component tests | 3.1–3.8 | ☑ |
| 4.1 | Config/log/health/shutdown | 1.5 | ☑ |
| 4.2 | MySQL migrations | 1.5 | ☑ |
| 4.3 | sqlc + repositories | 4.2 | ☑ |
| 4.4 | Money package (TDD/PBT) | 1.5 | ☑ |
| 4.5 | Envelope/errors/auth mw | 4.1 | ☑ |
| 5.1 | AuthService (TDD) | 4.3, 4.5 | ☑ |
| 5.2 | Profile & Wallet svc (TDD) | 4.3, 4.4 | ☑ |
| 5.3 | CatalogService | 4.3 | ☑ |
| 5.4 | TradingService (TDD) | 4.4, 5.2, 5.3 | ☑ |
| 5.5 | DCAService + scheduler (TDD) | 5.4 | ☑ |
| 5.6 | Portfolio & Dashboard svc | 5.4 | ☑ |
| 5.7 | HistoryService + Excel | 5.4 | ☑ |
| 5.8 | BondLifecycle/Notif/Admin | 5.4 | ☑ |
| 5.9 | MockExternal + price-gen | 4.3 | ☑ |
| 6.1 | Seed + demo investor | 4.2 | ☑ |
| 6.2 | Integration tests | 5.1–5.9, 6.1 | ☑ |
| 6.3 | Wire http impl + flip to API | 5.1–5.9 | ☑ |
| 7.1 | E2E flow verification | 6.3 | ☑ |
| 7.2 | Coverage + quality gate | 6.2 | ☑ (service 73.6%; money 89/auth 85/domain 75) |
| 7.3 | Responsive + a11y | W1 | ☑ |
| 8.1 | Design doc + diagrams (single file, Add-on) | W3 | ☐ |
| 9.1 | Prod build (optional) — delivered by 10.4/10.6 | W3.5 | ☐ |
| 9.2 | CI/deploy (optional) — delivered by 10.6 | 9.1 | ☐ |
| 10.1 | API auto-migration + deploy-friendly config (Add-on) | W3 | ☑ |
| 10.2 | Bond valuation / sell-price fix (Add-on) | W2b | ☑ |
| 10.3 | Monthly statement real PDF (Add-on) | W2b | ☑ |
| 10.4 | Web Export/Statement + web image in API mode (Add-on) | 10.3 | ☑ |
| 10.5 | One-command local bootstrap (Add-on) | 10.1, 10.4 | ☑ |
| 10.6 | AWS deploy scripts + CI, prepared not executed (Add-on) | 10.1, 10.4 | ☑ |
| 10.7 | Integrated verification + docs (Add-on) | 10.1–10.6 | ☑ |
| 10.8 | Real email delivery over SMTP, free (Add-on) | 10.7 | ☑ |
| 10.9 | OTP instead of links for verify/reset (Add-on) | 10.8 | ☑ |
| 10.10 | Red and white theme, modern and friendly UI (Add-on) | 10.9 | ☑ |
| 10.11 | Order entry: sell by amount/units + easier input (Add-on) | 10.10 | ☑ |

## Requirements Coverage

| US | Tasks |
|----|-------|
| US-001..005 (Account) | 3.1, 5.1 |
| US-006..010 (Profile/Wallet) | 3.2, 5.2 |
| US-011..012 (Market) | 3.3, 5.3 |
| US-013..018 (Trading) | 3.4, 4.4, 5.4 |
| US-019..021 (DCA) | 3.5, 5.5 |
| US-022 (Portfolio) | 3.6, 5.6 |
| US-023 (Dashboard) | 3.7, 5.6 |
| US-024..028 (History) | 3.8, 5.7 |
| US-029 (Notifications) | 5.8 |
| US-030 (Mock + price-gen) | 5.9 |
| US-031 (Admin) | 5.8 |
| US-032 (Seed/demo) | 2.3, 6.1 |

## Design Coverage
- **Services**: Auth 5.1, Profile/Wallet 5.2, Catalog 5.3, Trading 5.4, DCA 5.5, Portfolio/Dashboard 5.6, History 5.7, BondLifecycle/Notif/Admin 5.8, MockExternal/PriceGen 5.9.
- **Entities (20)**: migrations 4.2 + seed 6.1.
- **Endpoints**: per-service tasks 5.1–5.9; integration tests 6.2.
- **FE pages (9)**: 3.1–3.8.
- **Design documentation (Add-on)**: consolidated `docs/design-document.md` + full diagram set → 8.1.

## Testing Coverage
- **Unit (BE)**: within 5.1–5.9 (TDD for 5.1/5.2/5.4/5.5).
- **Property (money)**: 4.4 (P1,P4,P5,P6), 5.2 (P2), 5.4 (P3).
- **Integration**: 6.2.
- **FE component/format**: 2.1 (format), 3.9.
- **Coverage summary**: all 9 services → tests; all FE pages → component or flow coverage; endpoints → integration (6.2).

## Definition of Done
- Code complete per task Ref; dev server/app runs; tests for the task green.
- Each wave: build/compile passes, app runs, core flows click through, no console/terminal errors — verified before next wave.
- Backend business-logic coverage ≥ 70% (by W3).
- Acceptance criteria of mapped US-* satisfied.

## Execution Waves

| Wave | Phases | Resolved deps |
|------|--------|---------------|
| W0 Setup | 1 | — |
| W1 Frontend (mock) | 2, 3 | W0 |
| W2 Backend+MySQL+swap | 4, 5, 6 | W0 (W1 for 6.3) |
| W3 Verify | 7 | W2 |
| W3.5 Design Documentation (Add-on) | 8 | W3 |
| W3.6 Production readiness (Add-on) | 10 | W3 |
| W4 Deploy (optional) | 9 (delivered by 10.4/10.6; real AWS deploy remains a user action) | W3.6 |

**File ownership** (sequential, so no overlap risk): W0 root + `web/` scaffold + `api/` scaffold + `scripts/`; W1 `web/src/**`; W2 `api/**` + `web/src/services/*/http.ts`; W3 tests/verify; W3.5 `docs/`; W4 Dockerfiles/CI.

## Add-on (Change Requests)
- **W3.5 Design Documentation** (user request, 2026-10-28): added phase 8 — a single consolidated `docs/design-document.md` with a complete standard diagram set (C4 context/container, component, deployment, data flow, BPMN 2.0 w/ swimlanes, sequence, ER, user journey) in Mermaid, summarizing existing design/* (no invented content). Inserted as Wave W3.5 between Verify (W3) and Deploy (W4); Deploy renumbered phase 8→9. Totals 38→39 tasks / 10→11 phases / 5→6 waves.

- **Beginner Getting-Started guide** (user request, 2026-11-12): added `docs/GETTING-STARTED.md` (Thai, step-by-step for a reader with zero background) and linked it from README; README "Full stack" section corrected. Docs-only; no code or task-count change. Commands in it were verified on macOS from a fresh database in an isolated compose project (migrations → 20 tables, API ready, seed applied, demo login OK, Vite dev server in api mode, CORS for :5173).

- **W3.6 Production readiness** (user request, 2026-11-12): phase 10 (10.1–10.7) — auto-migration, bond valuation fix, real PDF statement, Export/Statement downloads, nginx web image in API mode, one-command bootstrap, AWS CloudFormation + dry-run deploy scripts + CI. Found during 10.7 and fixed in scope: web↔API contract mismatch (Portfolio/Dashboard crashed in API mode) → `web/src/services/apiMappers.ts` + small API enrichments; nginx `/assets/` regex vs SPA route `/assets/:type/:id`; TxnRow mobile overflow; test-DB isolation (`WEBOND_TEST_DSN`, guard against `webond`). Demo DB reset with user approval.

## Notes
- **Resolved (phase 10)**: the earlier known issues — compose `web` built in mock mode, History Export/Statement buttons disabled, no automatic migration, bond valuated as % instead of THB, statement PDF stub — are all fixed. Verified: api unit+integration on `webond_test`, web 89/89 tests, real-browser E2E 34/34 against `docker compose` (9 pages, buy, F5 session restore, xlsx/pdf downloads, mobile).
- **Remaining limitations**: never deployed to a real AWS account (9.1/9.2 delivered as scripts; running `deploy.sh --apply` is a user action and creates billable resources); RDS password rotation needs an ECS force-new-deployment; CSP is report-only; no real email (SES); admin UI minimal (D1-1); statement PDF text is English; reset-password does not revoke existing refresh JWTs; scheduler is manual-trigger not cron (D1-3); load tests deferred (D4-5).
- **Env**: dev uses `npm run dev` (mock, Node only). `scripts/setup-env.sh` + Docker/Go needed only for the real/backend build (W2+). This is the "setup for real build" per your D4-8 answer.
