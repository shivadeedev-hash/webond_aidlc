# Requirements — WeBond

## Summary
<!-- 10-line max, downstream reads ONLY this -->
- **Total Stories**: 32 across 9 functional areas
- **Priority**: 22 High, 8 Medium, 2 Low
- **User Types**: Investor (registered), Guest, Admin
- **Key Entities**: user, risk_assessment, wallet, wallet_transaction, fund, bond, gold_product, price/nav, order, holding, dca_plan, dca_round, transaction, notification
- **Integrations (mocked, DB-backed)**: Bank/Payment, Fund House/NAV, Bond Market, Gold Price, Notification
- **Core flows**: (1) Register→verify→profile/risk→deposit; (2) Browse Market→Asset detail→Buy/Sell; (3) Create & run DCA; (4) View Portfolio & Dashboard; (5) Review History→detail→cancel pending
- **Scope note (D1-1)**: Investor core journey = must build + test. Admin (FR-25) and scheduler (FR-15/FR-23) implemented as logic + manual trigger + rich seed, not always-on/ full back-office UI.

---

## Functional Areas

### Area 1 — Account & Authentication

#### US-001 — Register a new account
- **As a** Guest, **I want** to sign up with my details and accept consent, **so that** I can become an investor.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN a guest submits email, password, full name, mobile no., date of birth and checks T&C/PDPA consent, THEN the system shall validate all fields, verify email uniqueness, hash the password with bcrypt (cost ≥ 12), create the account in `Pending Verification` status, and e-mail a **6-digit verification OTP** (Add-on 2026-11-12b: OTP instead of link).
  - IF the email already exists, THEN the system shall reject with HTTP 409 and message "This email is already registered", ELSE it shall create the account.
  - IF T&C/PDPA consent is not accepted, THEN the system shall block submission and show a validation error.
  - WHEN registration succeeds, THEN the account shall not be able to trade until email is verified (BR-01).
- **Dependencies**: none
- **Source**: FR-01, EX-02, NFR-01, BR-01

#### US-002 — Verify email
- **As an** Investor, **I want** to verify my email via the OTP/link, **so that** my account becomes Active.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens the verification link / submits the OTP from the notification inbox, THEN the system shall set the account status to `Active`.
  - WHILE the account is `Pending Verification`, IF the user attempts any trade or DCA action, THEN the system shall block it and prompt to verify email.
- **Dependencies**: US-001
- **Source**: FR-01, FR-26, BR-01

#### US-003 — Log in and log out
- **As an** Investor, **I want** to log in and out securely, **so that** my session is protected.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN a user submits a valid email and password, THEN the system shall issue a JWT access token (15 min) and a refresh token (7 days) and redirect to the Dashboard.
  - IF login fails, THEN the system shall show the generic message "Invalid email or password" without revealing which field was wrong.
  - IF login fails 5 times in a row, THEN the system shall lock the account for 15 minutes.
  - WHEN the user clicks Logout, THEN the system shall revoke the refresh token.
  - WHILE a request has an expired access token, IF the refresh token is valid, THEN the system shall silently refresh; ELSE it shall redirect to Login.
- **Dependencies**: US-002
- **Source**: FR-02, EX-01, EX-03, NFR-01

#### US-004 — Forgot / reset password
- **As an** Investor, **I want** to reset a forgotten password, **so that** I can regain access.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHEN a user requests a password reset, THEN the system shall record a one-time reset link (valid 15 minutes) in the notification inbox, regardless of whether the email exists (no user enumeration).
  - WHEN the password is successfully changed, THEN the system shall invalidate all existing sessions.
  - IF the reset link is expired or already used, THEN the system shall reject it and allow requesting a new one.
- **Dependencies**: US-001
- **Source**: FR-03

#### US-005 — Guest access boundary
- **As a** Guest, **I want** to view public pages only, **so that** I can evaluate before signing up.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHERE the user is a guest, WHEN they open the Market catalog, THEN the system shall display assets and latest prices without requiring login.
  - WHERE the user is a guest, WHEN they attempt to open Dashboard, Portfolio, Wallet, DCA, History, or any Buy/Sell action, THEN the system shall redirect to Login.
- **Dependencies**: none
- **Source**: FR-05, user classes

---

### Area 2 — Profile & Risk Suitability

#### US-006 — Edit profile & withdrawal bank account
- **As an** Investor, **I want** to maintain my personal info and bank account, **so that** I can withdraw and keep my data current.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens Profile, THEN the system shall display personal info (full name, mobile, verified email) and the withdrawal bank account.
  - WHEN the user saves changes, THEN the system shall validate and persist them, masking sensitive fields (bank account) on display.
- **Dependencies**: US-003
- **Source**: FR-04, NFR-02

#### US-007 — Complete risk-suitability assessment
- **As an** Investor, **I want** to complete a risk questionnaire, **so that** I get a Risk Level that governs what I can buy.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user completes the questionnaire, THEN the system shall compute a Risk Level (1–5) and store it with the assessment date and a 2-year expiry.
  - WHILE the assessment is missing or expired, IF the user attempts to place an order or create a DCA plan, THEN the system shall block it and prompt to (re)take the assessment (BR-01).
  - WHEN the user chooses "Retake risk assessment", THEN the system shall start a new questionnaire and replace the prior result on completion.
- **Dependencies**: US-003
- **Source**: FR-04, BR-01, BR-08

---

### Area 3 — Wallet

#### US-008 — Deposit cash
- **As an** Investor, **I want** to deposit cash, **so that** I have balance to invest.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user confirms a deposit (mocked payment), THEN the system shall record a wallet transaction and increase the available balance atomically.
  - WHEN a deposit completes, THEN it shall appear in wallet history with type `Deposit` and a `+` amount.
- **Dependencies**: US-003
- **Source**: FR-07, BR-11, BR-15

#### US-009 — Withdraw cash
- **As an** Investor, **I want** to withdraw available cash, **so that** I can take money out.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user requests a withdrawal, THEN the system shall allow it only up to the available balance (balance − reserved).
  - IF the requested amount exceeds available balance, THEN the system shall reject it and prompt accordingly.
  - WHEN a withdrawal completes, THEN it shall appear in wallet history with type `Withdraw` and a `−` amount.
- **Dependencies**: US-008
- **Source**: FR-07, BR-11

#### US-010 — Wallet balances & reservation display
- **As an** Investor, **I want** to see balance, reserved, and available amounts, **so that** I understand my spendable cash.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - The system shall display Balance, Reserved, and Available (= Balance − Reserved) on the Profile & Wallet page.
  - WHEN an order is placed, THEN the system shall reserve the order amount from available balance; WHEN the order completes or is cancelled, THEN it shall release/settle the reservation.
- **Dependencies**: US-008
- **Source**: FR-07, BR-11

---

### Area 4 — Market (Asset Catalog)

#### US-011 — Browse & search assets
- **As an** Investor or Guest, **I want** to browse and search assets by type, **so that** I can find investments.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens Market, THEN the system shall list assets grouped by type tabs (Mutual Fund / Bond / Gold) showing latest price/NAV/yield and return over a selectable period.
  - WHEN the user selects a return period (3M / 6M / 1Y / 3Y / 5Y), THEN the system shall show each asset's return for that period and use it as the sort key for performance.
  - WHEN the user sorts by performance (ascending/descending) for the selected period, THEN the system shall reorder the asset list accordingly.
  - WHEN the user searches by name or code, THEN the system shall filter the list to matching assets.
  - WHEN the user applies filters (type, risk level, fund category, maturity), THEN the system shall return only assets matching all selected filters.
- **Dependencies**: none
- **Source**: FR-05; **Add-on** (user CR 2026-10-28): multi-period return filter (3M/6M/1Y/3Y/5Y) + sort by performance

#### US-012 — View asset detail by type
- **As an** Investor, **I want** to see detail appropriate to each asset type, **so that** I can make an informed decision.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user selects a Mutual Fund, THEN the system shall show a NAV history chart, risk level, fee (0% — free), min. purchase, and a fund fact sheet link.
  - WHEN the user selects a Bond, THEN the system shall show issuer, coupon rate, coupon frequency, maturity date, YTM, face value, credit rating, and offering period.
  - WHEN the user selects Gold, THEN the system shall show current buy (ASK) / sell (BID) price and a price history chart.
  - WHEN price/NAV data is older than the allowed threshold (gold 5 min, NAV 1 business day), THEN the system shall display "Price unavailable" with the last-updated time.
- **Dependencies**: US-011
- **Source**: FR-06, BR-17, EX-10

---

### Area 5 — Trading

#### US-013 — Buy Mutual Fund
- **As an** Investor, **I want** to buy a fund by THB amount, **so that** I invest in it.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user submits a buy order with a THB amount, THEN the system shall verify KYC/suitability, min. purchase (default 500 THB) and available balance, reserve the cash, and create an order in `Pending`.
  - WHEN NAV is confirmed, THEN the system shall allot units = amount ÷ NAV (4 decimals) and set the order `Completed`.
  - IF available balance < order amount, THEN the system shall reject with code INSUFFICIENT_BALANCE and prompt to deposit.
  - IF the order is placed after the 15:30 cut-off or on a non-business day, THEN the system shall warn the user and process it at the next business day's NAV (order still accepted).
  - WHEN the same order is submitted twice with the same Idempotency-Key, THEN the system shall return the original order and not create a duplicate.
- **Dependencies**: US-007, US-008, US-012
- **Source**: FR-08, BR-02, BR-03, EX-04, EX-06, EX-11, NFR-05

#### US-014 — Sell (redeem) Mutual Fund
- **As an** Investor, **I want** to redeem fund units, **so that** I realize proceeds.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user places a sell order by units or amount, THEN the system shall verify available (unlocked) units and lock them.
  - WHEN NAV is confirmed, THEN the system shall credit proceeds to the wallet per the fund's settlement (default T+3) and set status accordingly.
  - IF the sell quantity exceeds available units, THEN the system shall reject with code INSUFFICIENT_HOLDING.
- **Dependencies**: US-013
- **Source**: FR-09, BR-03, EX-05

#### US-015 — Buy Bond
- **As an** Investor, **I want** to buy a bond, **so that** I earn coupons and principal.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user buys a bond, THEN the system shall validate the offering window, min. lot (100,000 THB face value) and that the amount is a multiple of 100,000 THB, deduct cash, and record the holding with face value, units, purchase price and coupon schedule.
  - IF the offering is closed, sold out, below the 100,000 THB min lot, or not a multiple of 100,000 THB, THEN the system shall reject with code BOND_OFFER_INVALID and the reason.
  - The system shall not allow bonds to be added to a DCA plan (BR-06).
- **Dependencies**: US-007, US-008, US-012
- **Source**: FR-10, EX-07, BR-06; **Add-on** (user CR 2026-10-28): min lot changed from BR-04's 1,000 THB to **100,000 THB** (and multiples of 100,000) per user instruction — overrides source BR-04

#### US-016 — Sell Bond
- **As an** Investor, **I want** to sell a bond before maturity, **so that** I can exit early.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHEN the user sells a bond, THEN the system shall quote the current clean price plus accrued interest and show it for confirmation.
  - WHEN the user confirms, THEN the system shall reduce the holding and credit proceeds to the wallet.
  - IF the sell quantity exceeds the held bond units, THEN the system shall reject with code INSUFFICIENT_HOLDING.
- **Dependencies**: US-015
- **Source**: FR-11, EX-05

#### US-017 — Buy / Sell Gold
- **As an** Investor, **I want** to trade gold by THB amount or weight, **so that** I can invest in gold.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user enters a THB amount or weight (gram / baht-weight), THEN the system shall lock the current price for 30 seconds and show a quote (price, weight, total). Buy uses ASK, Sell uses BID.
  - WHEN the user confirms within 30 seconds, THEN the system shall execute and update the gold holding in grams (4 decimals).
  - IF the user confirms after the 30-second lock expires, THEN the system shall re-quote with the new price and require re-confirmation (code QUOTE_EXPIRED).
  - IF the gold price is older than 5 minutes, THEN the system shall block gold trading and show "Price unavailable".
  - IF available balance < buy total (or holding < sell weight), THEN the system shall reject (INSUFFICIENT_BALANCE / INSUFFICIENT_HOLDING).
- **Dependencies**: US-007, US-008, US-012
- **Source**: FR-12, BR-09, EX-08, EX-10, EX-04, EX-05

#### US-018 — Risk-mismatch acknowledgement
- **As an** Investor, **I want** to be warned when an asset exceeds my risk level, **so that** I invest knowingly.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - IF the asset risk level is greater than the user's risk level, THEN the system shall display a risk warning and require an explicit acknowledgement before placing the order, ELSE it shall proceed normally.
  - WHEN the user acknowledges the risk, THEN the system shall log the acknowledgement (audit).
- **Dependencies**: US-007, US-012
- **Source**: EX-12, BR-08

---

### Area 6 — DCA

#### US-019 — Create a DCA plan
- **As an** Investor, **I want** to create a recurring investment plan, **so that** I invest automatically.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user creates a DCA plan for a Mutual Fund or Gold, THEN the system shall capture asset, amount per round (THB), frequency (weekly Mon–Fri / monthly day 1–28), run day, start date and optional end date / number of rounds, and show the next execution date.
  - IF the selected asset is a Bond, THEN the system shall reject it (DCA eligible for MF and Gold only).
  - IF amount per round is below the minimum (500 THB for MF, 100 THB for Gold), THEN the system shall reject with a validation message.
- **Dependencies**: US-007
- **Source**: FR-13, BR-06, BR-07

#### US-020 — Manage DCA plans
- **As an** Investor, **I want** to view and change my plans, **so that** I stay in control.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens DCA Plans, THEN the system shall list all plans with status (Active / Paused / Completed / Cancelled), total invested, rounds executed and next run date.
  - WHEN the user edits amount or run day, pauses, resumes, or cancels a plan, THEN the system shall apply the change from the next round.
- **Dependencies**: US-019
- **Source**: FR-14

#### US-021 — DCA auto-execution (manual trigger in phase 1)
- **As an** Admin/System, **I want** due DCA plans to execute, **so that** investments happen on schedule.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHEN a DCA plan's run date is reached (or the dev "run scheduler now" endpoint is invoked), THEN the system shall create a buy order for the plan amount, link it to the plan and round, and record a notification of success/failure.
  - The system shall be idempotent per plan + run date (no duplicate round for the same date).
  - IF a round fails (insufficient balance or suspended fund), THEN the system shall mark the round `Failed`, notify the user, and not retry automatically; after 3 consecutive failures the plan becomes `Paused`.
- **Dependencies**: US-019, US-013, US-017
- **Source**: FR-15, EX-09 (phase-1 handling per D1-3)

---

### Area 7 — Portfolio

#### US-022 — View portfolio holdings
- **As an** Investor, **I want** to see all my holdings with cost and P/L, **so that** I know how I'm doing.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens Portfolio, THEN the system shall show each holding with quantity, average cost, market value, and unrealized P/L in THB and %.
  - WHERE a holding is a bond, WHEN displayed, THEN the system shall also show accrued interest, next coupon date and maturity date.
  - The system shall compute cost with the average-cost method and value MF at latest NAV, Bond at latest clean price, Gold at BID price.
  - The system shall display summary cards: Market value, Total cost, Unrealized P/L, Realized P/L (YTD).
- **Dependencies**: US-013, US-015, US-017
- **Source**: FR-21, BR-10

---

### Area 8 — Dashboard

#### US-023 — Insight dashboard
- **As an** Investor, **I want** a consolidated insight dashboard, **so that** I see my whole position at a glance.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens the Dashboard, THEN the system shall display total portfolio value and total P/L (today and all-time) and cash available.
  - The system shall display an asset allocation pie (Bond / MF / Gold) and a portfolio value trend chart with range toggles (1M / 6M / 1Y / 3Y / 5Y / All).
  - The system shall display top gainers/losers, a DCA summary (invested to date, avg cost vs current), and upcoming events (coupon payments, bond maturities, next DCA runs).
  - WHEN latest price/NAV updates, THEN dashboard values shall reflect the latest data and show the last-updated time.
- **Dependencies**: US-022
- **Source**: FR-22, EX-10; **Add-on** (user CR 2026-10-28): value trend range extended to 1M / 6M / 1Y / 3Y / 5Y / All

---

### Area 9 — History

#### US-024 — Transaction history list
- **As an** Investor, **I want** a dated list of all my transactions, **so that** I can review activity.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user opens History, THEN the system shall list the user's transactions (Buy, Sell, DCA Buy, Coupon, Maturity, Deposit, Withdraw) newest first, grouped by date, defaulting to the last 3 months, 20 per page with infinite scroll.
  - Each row shall show date/time, a type badge, asset code & name, signed amount (+/− THB), quantity @ price, and a status badge.
  - The system shall apply amount signs by wallet view: Sell/Coupon/Maturity/Deposit = `+`, Buy/DCA Buy/Withdraw = `−`.
- **Dependencies**: US-003
- **Source**: FR-16, BR-15

#### US-025 — Filter, search & sort history
- **As an** Investor, **I want** to filter and search history, **so that** I find specific transactions.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHEN the user changes a tab (All / Orders / DCA / Cash / Income), asset type, transaction type, status, or date range (7D / 1M / 3M / 6M / 1Y / custom), THEN the system shall filter the list accordingly and keep the filters in the URL query string.
  - WHEN the user searches by asset code/name or reference no., THEN the system shall return matching transactions; WHEN they sort by date or amount, THEN the list shall reorder.
  - IF the selected range exceeds 12 months, THEN the system shall reject with code RANGE_TOO_LARGE and ask to narrow the range.
  - IF no transaction matches, THEN the system shall show an empty state with "Clear filters" and a shortcut to Market.
- **Dependencies**: US-024
- **Source**: FR-17, BR-12, EX-14, EX-16

#### US-026 — Transaction detail
- **As an** Investor, **I want** to see full detail of a transaction, **so that** I understand it.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - WHEN the user selects a transaction, THEN the system shall show the fields relevant to its type (amount, fee = 0, net amount, quantity, price/NAV & date, accrued interest, realized P/L, DCA plan & round, trade/settlement date, reference no.) and a status timeline (Submitted → Processing → Completed/Failed/Cancelled).
  - IF a reference no. does not exist or belongs to another user, THEN the system shall return 404 "Transaction not found" (code TXN_NOT_FOUND) and never reveal another user's data.
- **Dependencies**: US-024
- **Source**: FR-18, EX-15, BR-16, BR-17

#### US-027 — Cancel a pending order
- **As an** Investor, **I want** to cancel an eligible pending order, **so that** I can undo a mistake.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHERE a transaction is a `Pending` MF Buy/Sell or DCA Buy before cut-off, WHEN the user clicks Cancel and confirms, THEN the system shall set status `Cancelled`, release reserved cash/units, and append the change to the status timeline.
  - IF the order is past cut-off, already Processing/Completed, or a non-cancellable type (Gold, Bond, Coupon, Maturity, Deposit, Withdraw), THEN the system shall reject with code CANCEL_NOT_ALLOWED and refresh the status.
- **Dependencies**: US-026
- **Source**: FR-19, BR-13, EX-13

#### US-028 — Export CSV & monthly statement
- **As an** Investor, **I want** to export my history, **so that** I can keep records.
- **Priority**: Low
- **Acceptance Criteria** (EARS):
  - WHEN the user clicks Export, THEN the system shall download the current filtered list as an Excel workbook (.xlsx) with a header row and Thai text rendered correctly.
  - IF the export exceeds 10,000 rows, THEN the system shall reject with code EXPORT_TOO_LARGE and ask to narrow the range or use a monthly statement.
  - WHERE a month has ended, WHEN the user opens History, THEN a downloadable monthly statement (PDF) shall be available with opening/closing balance, transactions and holdings.
- **Dependencies**: US-025
- **Source**: FR-20, EX-14; **Add-on** (user CR 2026-10-28): export format changed from CSV to **Excel (.xlsx)**

---

### Area 10 — Notifications, Admin & Mock Systems (phase-1 simplified)

#### US-029 — Notifications
- **As an** Investor, **I want** to be notified of key events, **so that** I stay informed.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - WHEN an order completes/fails, a DCA round runs, a coupon is paid, or a bond is within 30 days of maturity, THEN the system shall create an in-app notification and an email entry in the notification inbox.
  - WHERE the user has turned email notifications off, WHEN an event occurs, THEN the system shall still create the in-app notification but skip the email entry.
- **Dependencies**: US-013, US-021, US-023
- **Source**: FR-24

#### US-030 — Mock external systems backed by the DB
- **As the** System, **I want** mock REST endpoints reading/writing WeBond MySQL, **so that** external systems are simulated without real providers.
- **Priority**: Medium
- **Acceptance Criteria** (EARS):
  - The system shall provide mock endpoints for Bank/Payment, Fund House/NAV, Bond Market, Gold Price and Notification whose data comes from the WeBond MySQL tables (no separate DB).
  - The system shall provide an **auto-generate price feed** that periodically produces simulated gold prices (BID/ASK), fund NAVs and bond prices and writes them to the DB, so prices move on their own without manual admin updates.
  - The system shall expose a **toggle (on/off)** for the auto-generator; WHILE the toggle is ON, the system shall generate new price/NAV points on an interval (default ~1 min) with bounded random-walk movement and append them to price history; WHILE the toggle is OFF, prices shall stay fixed at the last generated values.
  - WHEN the toggle state changes, THEN the system shall persist the state and report the current state and interval.
  - WHEN a dev/admin endpoint is invoked to pay a coupon or run DCA, THEN the system shall update the corresponding DB rows and let dependent orders/holdings reflect the change.
  - The system shall not perform any real outbound network call to a real provider.
- **Dependencies**: seed data
- **Source**: FR-26; **Add-on** (user CR 2026-10-28): since data is mocked, replace manual admin price updates with an **auto price-data generator + on/off toggle** (simulated realtime)

#### US-031 — Admin asset & price management (minimal, phase 1)
- **As an** Admin, **I want** to manage master data and prices, **so that** the catalog stays correct.
- **Priority**: Low
- **Acceptance Criteria** (EARS):
  - WHERE the user has the Admin role, WHEN they create/edit/deactivate a fund, bond or gold product or update a NAV/price, THEN the system shall persist the change and write an audit log entry (who / what / when).
  - IF a non-admin attempts an admin action, THEN the system shall reject with HTTP 403.
- **Dependencies**: US-003
- **Source**: FR-25, NFR-02 (scope-limited per D1-1)

#### US-032 — Rich seed data & demo account
- **As a** Reviewer, **I want** pre-seeded data and a ready demo login, **so that** I can evaluate the app immediately.
- **Priority**: High
- **Acceptance Criteria** (EARS):
  - The system shall seed the full asset catalog (funds, bonds, gold) with price/NAV history, and one pre-verified demo investor with wallet balance, holdings across all three asset types, DCA plans, and ~3 months of transactions.
  - WHEN the reviewer uses the printed demo credentials, THEN the system shall log them in and all pages shall show realistic, non-empty data.
- **Dependencies**: US-030
- **Source**: D1-5, user request ("Help to initial all necessary mock data")

---

## Story Summary

| ID | Title | Area | Priority | Dependencies |
|----|-------|------|----------|--------------|
| US-001 | Register a new account | Account | High | — |
| US-002 | Verify email | Account | High | US-001 |
| US-003 | Log in and log out | Account | High | US-002 |
| US-004 | Forgot / reset password | Account | Medium | US-001 |
| US-005 | Guest access boundary | Account | Medium | — |
| US-006 | Edit profile & bank account | Profile | High | US-003 |
| US-007 | Risk-suitability assessment | Profile | High | US-003 |
| US-008 | Deposit cash | Wallet | High | US-003 |
| US-009 | Withdraw cash | Wallet | High | US-008 |
| US-010 | Wallet balances & reservation | Wallet | High | US-008 |
| US-011 | Browse & search assets | Market | High | — |
| US-012 | Asset detail by type | Market | High | US-011 |
| US-013 | Buy Mutual Fund | Trading | High | US-007, US-008, US-012 |
| US-014 | Sell Mutual Fund | Trading | High | US-013 |
| US-015 | Buy Bond | Trading | High | US-007, US-008, US-012 |
| US-016 | Sell Bond | Trading | Medium | US-015 |
| US-017 | Buy / Sell Gold | Trading | High | US-007, US-008, US-012 |
| US-018 | Risk-mismatch acknowledgement | Trading | High | US-007, US-012 |
| US-019 | Create DCA plan | DCA | High | US-007 |
| US-020 | Manage DCA plans | DCA | High | US-019 |
| US-021 | DCA auto-execution (manual trigger) | DCA | Medium | US-019, US-013, US-017 |
| US-022 | Portfolio holdings | Portfolio | High | US-013, US-015, US-017 |
| US-023 | Insight dashboard | Dashboard | High | US-022 |
| US-024 | Transaction history list | History | High | US-003 |
| US-025 | Filter, search & sort | History | Medium | US-024 |
| US-026 | Transaction detail | History | High | US-024 |
| US-027 | Cancel pending order | History | Medium | US-026 |
| US-028 | Export CSV & statement | History | Low | US-025 |
| US-029 | Notifications | Notification | Medium | US-013, US-021, US-023 |
| US-030 | Mock external systems | Integration | Medium | seed |
| US-031 | Admin asset & price mgmt | Admin | Low | US-003 |
| US-032 | Rich seed & demo account | Data | High | US-030 |

## Story × Persona Matrix

| ID | Investor | Guest | Admin |
|----|----------|-------|-------|
| US-001 | — | Primary | — |
| US-002 | Primary | — | — |
| US-003 | Primary | — | Secondary |
| US-004 | Primary | — | — |
| US-005 | Secondary | Primary | — |
| US-006 | Primary | — | — |
| US-007 | Primary | — | — |
| US-008–US-010 | Primary | — | — |
| US-011 | Primary | Primary | — |
| US-012 | Primary | Secondary | — |
| US-013–US-018 | Primary | — | — |
| US-019–US-021 | Primary | — | Secondary (trigger) |
| US-022–US-023 | Primary | — | — |
| US-024–US-028 | Primary | — | — |
| US-029 | Primary | — | — |
| US-030 | — | — | Secondary |
| US-031 | — | — | Primary |
| US-032 | Primary (demo) | — | Secondary |

## Non-Functional Notes
- **Security (NFR-01)**: bcrypt ≥ 12, JWT access/refresh, RBAC (Investor/Admin), encrypt sensitive PII at rest, OWASP Top 10 (SQLi via parameterized queries, XSS, CSRF). HTTPS in deployed envs.
- **Compliance/PDPA (NFR-02)**: consent capture at registration, on-screen data masking, audit log for login/orders/DCA/admin changes, right to export/delete (export covered by US-028).
- **Data integrity (NFR-05)**: money & units as DECIMAL(18,4), order/balance updates in a single DB transaction, Idempotency-Key on order APIs.
- **Performance (NFR-04)**: API p95 < 500 ms, dashboard < 2 s (targets; verified qualitatively in phase 1).
- **Usability (NFR-07)**: responsive desktop + mobile, Thai default + i18n-ready, THB/number/date formatting per BR-15.
- **Observability/quality (NFR-08)**: structured logs, health-check endpoint, OpenAPI docs, backend unit-test coverage ≥ 70% for business logic.
- **Fees (BR-17)**: all fees 0 in phase 1; fee column retained in the model.

## External References

| Source | Stories derived | What was used |
|--------|-----------------|---------------|
| (MSC) AIDLC Workshop - Usecase WeBond.md | US-001..US-032 | FR-01..FR-26, EX-01..EX-16, BR-01..BR-17, NFR-01..NFR-08 |
| images/01..09 *.png | US-003,006–012,019–028 | Page layouts, fields, navigation, badges, status timeline |
| decisions-requirements.md (D1) | Scope/auth/scheduler/seed decisions | Phase-1 scope boundaries |

---

## Add-on (Change Requests)

Changes made at the user's request that extend or override the source spec. Recorded per workshop policy.

| Date | Story | Change | Overrides source? |
|------|-------|--------|-------------------|
| 2026-10-28 | US-011 | Market return filter supports periods 3M/6M/1Y/3Y/5Y; add sort by performance for the selected period | Extends FR-05 (spec only mentioned 1Y return) |
| 2026-10-28 | US-015 | Bond min lot changed to **100,000 THB** and multiples of 100,000 | **Overrides BR-04** (was 1,000 THB) — per user instruction, not source |
| 2026-10-28 | US-023 | Dashboard value-trend range extended to 1M/6M/1Y/**3Y/5Y**/All | Extends FR-22 (was 1M/6M/1Y/All) |
| 2026-10-28 | US-028 | History export format changed from CSV to **Excel (.xlsx)** | **Overrides FR-20** (was CSV UTF-8 BOM); monthly PDF statement unchanged |
| 2026-10-28 | US-030 | Add **auto price-data generator** (gold/NAV/bond price simulated realtime) with **on/off toggle**, replacing manual admin price updates | Extends FR-26 (spec had admin-triggered publish + 1-min gold tick) |
| 2026-11-12d | US-013/014/015/016/017 | **Easier, working order entry** (user report: "amount must be positive" whatever is typed). Bug: fund/bond SELL always failed because the web sent a THB amount while the API only read units (US-014 already requires selling "by units or amount"). Fix: API sells funds/bonds by `amount` (THB → units at the current NAV/price, 4 dp) or by `units`; INSUFFICIENT_HOLDING states the maximum. Web: plain text input accepting `1,000` / `฿1000` / spaces, live thousands formatting, minus sign impossible; sell shows the holding and lets the user switch THB ↔ units (gold: THB ↔ grams); quick-pick chips (buy: minimum / common amounts / all cash; bond: lot multiples; sell: 25% / 50% / 100%); specific inline messages (minimum, lot multiple, more than cash, more than holding) instead of the generic validation text. The DCA plan form's amount-per-round field gets the same input (comma-friendly text, no minus, quick picks at/above the BR-07 minimum); its rules and messages are unchanged. | Fixes US-014/016 against the source; UX extends US-013..017 |
| 2026-11-12c | All UI | **Red and white theme, modern and friendly** (user final request). White surfaces, red brand and primary actions, soft red tints for active/selected states, pill buttons, larger radii, red gradient hero on auth/dashboard. Loss/danger uses a distinct brick red and always keeps the +/- sign and icon so it is not confused with brand red. WCAG 2.2 AA contrast kept. | **Overrides** the original theme (white BG, dark-navy header, navy-blue buttons) in product success criteria |
| 2026-11-12b | US-001/002/004 | **OTP instead of links** (user CR: the web expects an OTP, and users read mail on a phone that cannot open `localhost` links). Verification and password-reset e-mails carry a **6-digit numeric OTP**, no link. Verify = `{email, code}`; reset = `{email, code, newPassword}`; new `POST /api/auth/resend-verification {email}`. Rules: OTP valid **15 minutes**; **5 wrong attempts** invalidate it; a new code invalidates older ones; resend cooldown **60 s**; only an HMAC of the code is stored for checking and the plain code is removed from the outbox once the e-mail is sent; responses never reveal whether an e-mail exists (except the existing EX-02 duplicate-email on register); the code is echoed in API responses only outside production. | Replaces the link part of the 2026-11-12 row; FR-01 allows "link/OTP"; FR-03 "reset link" delivered as an OTP per user instruction |
| 2026-11-12 | US-001/002/004/029/030 | **Real email delivery, 100% free** (bug fix: a new user could not verify through the UI because the simulated inbox requires login). Every `notification_outbox` email row is delivered by a background sender over **SMTP** (transactional outbox: written in the business transaction, sent after commit, retried, `sent_at` recorded). Local default = **Mailpit** container (free, MIT, nothing leaves the machine; inbox UI at :8025). Real inboxes = any SMTP server by config, documented with **Gmail SMTP + app password** (free, no domain). Emails carry clickable links `/verify-email?token=…` and `/reset-password?token=…`; the verify page accepts the token from the URL. The register/forgot API responses return the token **only outside production**. SES is not used (not free). | Implements FR-01 "send an email verification link/OTP" + FR-26 "swap to real provider by config"; **overrides US-030's "no real outbound call"** for email only, and only when the operator configures an external SMTP server |

> Note on US-030 vs US-031: with the auto-generator handling price movement, admin price editing in US-031 becomes optional/secondary in phase 1. Admin still manages master data (create/edit/deactivate assets) and can pay coupon / run DCA via dev endpoints.
