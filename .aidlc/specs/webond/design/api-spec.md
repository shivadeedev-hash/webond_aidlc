# API Specification

## Overview
- **Style**: REST/JSON. **Base URL**: `/api`.
- **Auth**: JWT Bearer (access token, 15 min) in `Authorization` header; refresh token in httpOnly secure cookie; `POST /api/auth/refresh` rotates.
- **RBAC**: `investor` default; `admin` for `/api/admin/*` and generator controls.

## Conventions
- **Envelope**: success `{ "data": <payload>, "meta": {...}? }`; error `{ "error": { "code": "STRING", "message": "...", "details": {...}? }, "meta": { "requestId": "uuid" } }`.
- **Pagination**: `?page=1&pageSize=20`; `meta: { page, pageSize, total, hasMore }`. History uses infinite scroll (cursor or page).
- **Filtering/sorting**: query params; History keeps filters in URL (FR-17).
- **Idempotency**: `Idempotency-Key` header on order POSTs (EX-11, NFR-05).
- **Money**: strings in JSON for DECIMAL to avoid float loss (e.g. `"1250400.0000"`).
- **Error codes**: `INSUFFICIENT_BALANCE`(422), `INSUFFICIENT_HOLDING`(422), `BOND_OFFER_INVALID`(422), `QUOTE_EXPIRED`(409), `CANCEL_NOT_ALLOWED`(409), `RANGE_TOO_LARGE`(400), `EXPORT_TOO_LARGE`(400), `TXN_NOT_FOUND`(404), `DUPLICATE_EMAIL`(409), `RISK_ACK_REQUIRED`(422), `SUITABILITY_REQUIRED`(403), `UNAUTHORIZED`(401), `FORBIDDEN`(403), `VALIDATION_ERROR`(400), `PRICE_UNAVAILABLE`(409).

## Endpoints

### Auth
- `POST /api/auth/register` — body {email,password,fullName,mobile,dateOfBirth,consent}. 201 or 409 DUPLICATE_EMAIL. (US-001)
- `POST /api/auth/verify` — body {email,code} (6-digit OTP from the e-mail). 200 → status active; 422 VALIDATION_ERROR for wrong/expired/used code (5 wrong attempts invalidate it). (US-002, Add-on 2026-11-12b)
- `POST /api/auth/resend-verification` — body {email}. Always 202 (no enumeration); sends a new OTP to pending accounts, at most once per 60 s. (US-002, Add-on 2026-11-12b)
- `POST /api/auth/login` — body {email,password}. 200 {accessToken,user} + sets refresh cookie; 401 generic; lockout after 5 (EX-01). (US-003)
- `POST /api/auth/refresh` — uses refresh cookie. 200 {accessToken} or 401. (EX-03)
- `POST /api/auth/logout` — revokes refresh. 204. (US-003)
- `POST /api/auth/forgot-password` — body {email}. Always 202 (no enumeration). (US-004)
- `POST /api/auth/reset-password` — body {email,code,newPassword} (6-digit OTP, 15 min, one use, 5 attempts). 200, invalidates sessions. (US-004, Add-on 2026-11-12b)

### Profile & Risk
- `GET /api/profile` — 200 profile. (US-006)
- `PUT /api/profile` — body {fullName,mobile,bankAccount}. 200. (US-006)
- `GET /api/risk-assessment` — latest result or null. (US-007)
- `POST /api/risk-assessment` — body {answers}. 200 {riskLevel,assessedAt,expiresAt}. (US-007)

### Wallet
- `GET /api/wallet` — 200 {balance,reserved,available}. (US-010)
- `GET /api/wallet/transactions?page=&pageSize=` — 200 list. (US-008/009/010)
- `POST /api/wallet/deposit` — body {amount}. 200. (US-008)
- `POST /api/wallet/withdraw` — body {amount}. 200 or 422 INSUFFICIENT_BALANCE. (US-009)

### Market / Assets
- `GET /api/assets?type=fund|bond|gold&q=&riskLevel=&category=&maturity=&period=3M|6M|1Y|3Y|5Y&sort=performance|name&order=asc|desc&page=&pageSize=` — 200 list with latest price/NAV + period return. Public (guest). (US-011 Add-on)
- `GET /api/assets/{type}/{id}` — 200 type-specific detail. Public. (US-012)
- `GET /api/assets/{type}/{id}/prices?period=1M|6M|1Y|3Y|5Y|all` — 200 history series for charts. (US-012)

### Orders / Trading
- `POST /api/orders` — Idempotency-Key header; body {assetType,assetId,side,orderMode,amount|units|weightGrams,riskAck?}. 201 order; errors per codes above (US-013..018). Gold returns a 30s quote first via:
- `POST /api/orders/quote` — body {assetType:'gold',assetId,side,amount|weightGrams}. 200 {quoteId,price,weight,total,expiresAt}. (US-017)
- `POST /api/orders/{quoteId}/confirm` — confirm gold within 30s; 409 QUOTE_EXPIRED if late. (US-017)
- `GET /api/orders/{txnRef}` — 200 or 404 TXN_NOT_FOUND (ownership). (US-026)

### DCA
- `GET /api/dca-plans` — 200 list with status/invested/next run. (US-020)
- `POST /api/dca-plans` — body {assetType:'fund'|'gold',assetId,amountPerRound,frequency,runDay,startDate,endDate?,maxRounds?}. 201; 422 if bond or below min. (US-019)
- `PUT /api/dca-plans/{id}` — edit amount/day. (US-020)
- `POST /api/dca-plans/{id}/pause` | `/resume` | `DELETE /api/dca-plans/{id}` (cancel). (US-020)
- `GET /api/dca-plans/{id}/rounds` — round history. (US-020)

### Portfolio & Dashboard
- `GET /api/portfolio` — 200 {summary, holdings[]}. (US-022)
- `GET /api/dashboard?trend=1M|6M|1Y|3Y|5Y|all` — 200 {totals, allocation, trend[], topMovers, dcaSummary, upcomingEvents}. (US-023 Add-on)

### History
- `GET /api/transactions?tab=all|orders|dca|cash|income&assetType=&type=&status=&from=&to=&q=&sort=date|amount&order=&page=&pageSize=20` — 200 grouped list; 400 RANGE_TOO_LARGE if >12mo. (US-024/025)
- `GET /api/transactions/{txnRef}` — 200 detail + status timeline; 404 TXN_NOT_FOUND. (US-026)
- `POST /api/transactions/{txnRef}/cancel` — 200 cancelled or 409 CANCEL_NOT_ALLOWED. (US-027)
- `GET /api/transactions/export?<same filters>` — 200 `.xlsx` stream; 400 EXPORT_TOO_LARGE if >10k rows. (US-028 Add-on)
- `GET /api/statements/{yyyyMM}` — 200 PDF (monthly). (US-028)

### Notifications
- `GET /api/notifications?page=` — in-app list. `POST /api/notifications/{id}/read`. (US-029)
- `GET /api/inbox` — simulated email inbox incl. OTP/verify/reset (dev/reviewer convenience, FR-26). (US-002/004)
- `PUT /api/profile/notification-settings` — body {emailEnabled}. (US-029)

### Admin & Mock controls (RBAC admin / dev)
- `POST /api/admin/price-generator` — body {enabled,intervalSeconds?,volatilityPct?} toggle on/off. `GET /api/admin/price-generator` state. (US-030 Add-on)
- `POST /api/admin/run/dca` | `/run/coupons` | `/run/maturities` — manual scheduler triggers (D1-3). (US-021/US-023)
- `POST /api/admin/publish-nav` | `/tick-gold` — manual mock events (FR-26).
- `GET|POST|PUT /api/admin/funds|bonds|gold` — master data CRUD (FR-25). `GET /api/admin/audit-logs`. (US-031)

### Health
- `GET /health` — liveness. `GET /health/ready` — readiness (DB check).

## Example — place gold buy
Request `POST /api/orders/quote` → `{ "data": { "quoteId":"q_123","price":"41250.0000","weight":"0.2424","total":"10000.0000","expiresAt":"2026-10-28T10:00:30+07:00" } }`
Confirm `POST /api/orders/q_123/confirm` (within 30s) → `201 { "data": { "txnRef":"WB2610280000045","status":"completed", ... } }`; if late → `409 { "error": { "code":"QUOTE_EXPIRED","message":"Price quote expired, please re-quote" } }`.
