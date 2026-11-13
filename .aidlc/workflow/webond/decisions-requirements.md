# Requirements Decisions (D1)

## Context Summary
WeBond — greenfield full-stack retail investment platform (Bond, Mutual Fund, Gold). Stack: React 18 + TS (Vite) + Tailwind, Go (Gin) REST API, MySQL 8.0 (sample data seeded into DB). Mode: comprehensive (no unit decomposition). The requirements document (`initial-requirements/*`) already defines 26 FRs, 16 exceptions (EX), 17 business rules (BR), NFRs, and 9 wireframes — these are the source of truth. The questions below only cover choices the source leaves open or that shape phase-1 scope.

---

## Decision Questions

### D1-1: Phase-1 feature scope (which FRs are "must build and pass" now)
**Question**: The spec has 26 FRs. For this hands-on build, which scope do we target for "usable + tested"?
- 1) **Investor core journey** — Account (register/login/reset), Profile+Wallet, Market+Detail, Buy/Sell (MF/Bond/Gold), DCA, Portfolio, Dashboard, History+Detail (the 9 wireframe pages). Admin back-office (FR-25) and scheduler/batch jobs (FR-15, FR-23) mocked/simplified. **(Recommended)**
- 2) Everything in the spec including Admin back-office UI, real scheduler for DCA auto-execution, and coupon/maturity batch jobs
- 3) Read-only MVP — catalog, portfolio, dashboard, history viewing only (no trading/DCA)
- 4) Other (please specify): _______

**Answer**: *Investor core journey

---

### D1-2: Authentication depth for phase 1
**Question**: How real should auth be in this build (NFR-01 targets bcrypt + JWT access/refresh)?
- 1) **Full real auth** — bcrypt password hashing, JWT access (15m) + refresh (7d), account lockout after 5 fails (EX-01), reset flow; email verification/OTP simulated via a viewable notification inbox (FR-26). **(Recommended)**
- 2) Full auth above PLUS real email sending (SMTP) and real 2FA OTP
- 3) Simplified — session token only, no refresh/lockout, verification auto-passed
- 4) Other (please specify): _______

**Answer**: Full real auth

---

### D1-3: DCA auto-execution & bond coupon/maturity (scheduler-driven FRs)
**Question**: FR-15 (DCA auto-run), FR-23 (coupon/maturity) need a background scheduler. How to handle in phase 1?
- 1) **Manual trigger + seeded history** — logic implemented and unit-tested; an admin/dev "run scheduler now" endpoint triggers it on demand; seed sample DCA rounds, coupons, maturities so the UI shows realistic data. No always-on cron. **(Recommended)**
- 2) Real always-on background worker (Go goroutine/cron) running on schedule
- 3) Skip execution — only create/manage DCA plans and show static seeded data
- 4) Other (please specify): _______

**Answer**: Manual trigger + seeded history

---

### D1-4: Language & localization for phase 1 UI
**Question**: NFR-07 says Thai/English supported, default Thai. What do we implement now?
- 1) **Thai default + i18n-ready structure**, with English strings where quick; full EN translation can follow. Number/THB/date formatting per BR-15 done properly. **(Recommended)**
- 2) Full bilingual Thai + English with a language switcher, all strings translated
- 3) Thai only, no i18n scaffolding
- 4) Other (please specify): _______

**Answer**: Thai default + i18n-ready structure

---

### D1-5: Seeded demo data & demo login
**Question**: "Help to initial all necessary mock data." How much seeded data + how does the reviewer log in?
- 1) **Rich seed + ready demo account** — one pre-verified investor with wallet balance, holdings (MF/Bond/Gold), DCA plans, ~3 months of transactions, plus full asset catalog with price/NAV history; printed demo credentials. **(Recommended)**
- 2) Minimal seed — catalog + prices only; user registers fresh to see empty states
- 3) Rich seed as (1) PLUS multiple demo users (incl. an admin)
- 4) Other (please specify): _______

**Answer**: Rich seed + ready demo account

---

### D1-6: Personas
**Question**: Should we generate detailed personas? (3 user classes: Investor, Guest, Admin with different goals/access.)
- 1) **Yes** — generate personas for Investor, Guest, Admin (drives access rules & journeys). **(Recommended)**
- 2) No — user types are clear enough from the spec; skip personas
- 3) Other (please specify): _______

**Answer**: Yes

---

### D1-7: Team size
**Question**: How many developers will work on this project? (Used by later design/task gates.)
- 1) **Solo (1 developer)** **(Recommended — this is an AI-assisted hands-on build)**
- 2) Small team (2–3)
- 3) Medium team (4–8)
- 4) Large team (9+)

**Answer**: Solo

---

## Decisions Summary
<!-- Populated after you answer. One line per decision. -->
- D1-1 Phase-1 scope: Investor core journey (9 wireframe pages; Admin FR-25 + scheduler FR-15/FR-23 simplified)
- D1-2 Auth depth: Full real auth (bcrypt, JWT access 15m + refresh 7d, lockout after 5 fails, reset; email/OTP via viewable notification inbox)
- D1-3 Scheduler FRs: Manual trigger + seeded history (logic unit-tested, dev "run now" endpoint, realistic seeded rounds/coupons/maturities; no always-on cron)
- D1-4 Language: Thai default + i18n-ready structure; proper THB/number/date formatting per BR-15
- D1-5 Seed data: Rich seed + one ready pre-verified demo investor (wallet, holdings, DCA, ~3mo transactions, full catalog + price/NAV history; printed credentials)
- D1-6 Personas: Yes (Investor, Guest, Admin)
- D1-7 Team size: Solo (1 developer)

---

## Validation Notes
- All answers = recommended options. No conflicts detected.
- Checked: Enterprise-scope-solo (D1-1 scoped to investor core, not full enterprise → OK); Compliance-without-security-stories (PDPA + OWASP present → will include security/privacy stories, NFR enabled → OK); Complex-features-without-personas (3 user types, personas=Yes → OK); Heavy-integration-load (5 external systems, all mocked/seeded in phase 1 → OK); Performance-without-NFR (NFR targets present, NFR analysis enabled → OK).

---

**Instructions**: Fill in your answers above and respond with "done" — or say "use recommendations" to auto-fill all recommended options.
