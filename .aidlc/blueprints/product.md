# Product

## Summary
- **Product**: WeBond — a retail investment platform to buy, sell and manage Bonds, Mutual Funds and Gold in one portfolio.
- **Users**: Investor (registered), Guest, Admin (back-office).
- **Type + Scope**: Greenfield, new product.

## Overview
WeBond is a greenfield web application where an individual investor holds and manages multiple asset classes (Bond, Mutual Fund, Gold) in one place: buy/sell, automate investing with DCA (Dollar-Cost Averaging), and see a consolidated portfolio with insights (allocation, P/L, performance, upcoming coupons/maturities). It is built as a full stack: a React web frontend, a Go (Gin) REST API, and a MySQL database. Sample/mock data is seeded into MySQL; external systems (bank/payment, fund house, bond market, gold price, notification) are mocked but backed by the same database.

## Problem Statement
Retail investors juggle separate apps and accounts for bonds, mutual funds and gold, making it hard to see a single consolidated portfolio, automate recurring investments, and track income events (coupons, maturities). WeBond unifies catalog, trading, DCA, wallet, history and insights in one place.

## Target Users
- **Investor (Registered User)**: register, log in, complete profile/risk assessment, deposit/withdraw cash, buy/sell all three asset types, manage DCA plans, view portfolio and dashboard, review history. Primary goal: grow and monitor investments easily.
- **Guest**: view the landing page and public asset catalog/prices only. Primary goal: evaluate before signing up.
- **Admin (Back-office)**: manage asset master (funds, bonds, gold), prices/NAV, users; view audit log. Primary goal: keep master data and prices correct.

## Key Features
- **Account**: registration (email verification), login/logout (JWT), forgot/reset password, profile + risk-suitability assessment (Risk Level 1–5).
- **Market (Asset Catalog)**: browse/search assets grouped by type with filters; asset detail (NAV/price history, fund/bond/gold specifics).
- **Wallet**: cash deposit/withdraw, reserved vs available balance, wallet history.
- **Trading**: Buy/Sell Mutual Fund (NAV, cut-off, T+n), Buy/Sell Bond (lot, offering, coupon schedule), Buy/Sell Gold (BID/ASK, 30s price lock).
- **DCA**: create/manage recurring plans for Mutual Fund and Gold (weekly/monthly), auto-execution, pause/resume/cancel.
- **History**: transaction list grouped by date, filter/search/sort, transaction detail with status timeline, cancel pending order, CSV export + monthly statement.
- **Portfolio**: holdings with qty, avg cost, market value, unrealized P/L, next events.
- **Dashboard**: total value & P/L, allocation pie, value trend chart, top movers, DCA summary, upcoming events.
- **Notification**: in-app + email for order/DCA/coupon/maturity events.
- **Admin / Mock External Systems**: asset & price management, mock REST APIs backed by the WeBond data.

## Domain Language

| Term | Definition | Example |
|------|-----------|---------|
| MF | Mutual Fund | K-GLOBAL |
| NAV | Net Asset Value per unit | 15.2341 |
| DCA | Dollar-Cost Averaging — invest a fixed amount on a fixed schedule | THB 5,000 monthly |
| Unit | MF holding unit | 328.2143 u |
| Face Value | Bond par value | THB 1,000 |
| Coupon | Periodic bond interest | THB 1,250 |
| YTM | Yield to Maturity | 3.2% |
| Cut-off time | Latest order time to get today's NAV | 15:30 |
| T+n | Settlement n business days after trade date | T+3 |
| P/L | Profit / Loss | +THB 84,200 |
| Baht-weight | Thai gold unit | 1 baht-weight = 15.244 g |
| Price auto-generator | Background job that simulates moving prices/NAVs; toggle on/off | gold tick ~1 min |
| Reserved balance | Cash held for pending orders | THB 5,000 |
| Txn ref | Unique transaction reference | WB2610270000123 |

## Success Criteria
- All 9 UI pages are built and usable end-to-end against the Go API + MySQL (runnable full-stack app; frontend-first on mock JSON as an interim step).
- All in-scope features per requirements work and pass tests before hand-off.
- UX is clean, modern and friendly: **red and white theme** (white surfaces, red brand/primary actions, soft red tints, black text; Add-on 2026-11-12c replaces the original navy theme); responsive.
- Financial display rules honored (THB 2 decimals, units/grams 4 decimals, +/- sign by wallet view).

## Constraints & Assumptions
- **Constraints**: Currency THB only; timezone Asia/Bangkok; money/units as DECIMAL (never float); PDPA compliant; responsive desktop + mobile; default language Thai (Thai/English supported). Full stack React + Go + MySQL; external systems mocked but backed by MySQL data. No real payment, no real external provider network calls.
- **Assumptions**: Several business values (cut-off 15:30, min purchase 500 THB, T+3, DCA minimums) are stated as assumptions in the spec and used as defaults. All fees are 0 in phase 1 (fee column retained for future).

## Add-on (Change Requests)
- Bond min lot = **100,000 THB** (multiples of 100,000) — user override of source BR-04 (1,000 THB).
- Market: return filter periods 3M/6M/1Y/3Y/5Y + sort by performance.
- Dashboard value-trend range: 1M/6M/1Y/3Y/5Y/All.
- History export format: **Excel (.xlsx)** (not CSV); monthly PDF statement unchanged.
- Mock data: **auto price-data generator with on/off toggle** (simulated realtime) replaces manual admin price updates.
- E-mail: real SMTP delivery (free: Mailpit locally, Gmail app password for real inboxes); verification and password reset use a **6-digit OTP**.
- UI theme: **red and white**, modern and friendly (replaces white/dark-navy/navy-blue).

## Project Type
- **Type**: Greenfield
- **Scope**: New product
