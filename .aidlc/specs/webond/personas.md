# Personas

## Summary
<!-- Compact digest for downstream phases. -->
- **User Types**: 3 personas
- **Key Roles**: Investor (registered), Guest, Admin (back-office)
- **Design Implications**: RBAC (Investor vs Admin), guest-accessible public catalog, KYC/suitability gate before trading, responsive UI, PDPA data handling

## Overview
WeBond serves 3 distinct user types. The Investor is the primary persona and owns the 9 wireframe pages. Guest is a pre-signup, read-only visitor. Admin maintains master data and prices (simplified in phase 1 per D1-1).

---

## Investor (Registered User)

**Role**: Individual retail investor managing Bond, Mutual Fund and Gold in one place

**Goals**:
- Grow and monitor investments across asset classes from one dashboard
- Automate recurring investing with DCA and track income events (coupons, maturities)

**Pain Points**:
- Separate apps/accounts per asset class; no consolidated portfolio or P/L view
- Manual, easy-to-forget recurring purchases; hard to track settlement and cut-off timing

**User Journey**: Register + verify email → complete profile + risk assessment → deposit cash → browse Market → Buy MF/Bond/Gold (or set up DCA) → view Portfolio & Dashboard → review History

**Implications**: Must enforce KYC/suitability before trading (BR-01), risk-mismatch acknowledgement (BR-08/EX-12), cash reservation (BR-11), accurate DECIMAL money/units (NFR-05), clear financial formatting (BR-15). Needs responsive, friendly UI.

---

## Guest

**Role**: Unauthenticated visitor evaluating the platform

**Goals**:
- See what WeBond offers and browse the public asset catalog/prices before signing up

**Pain Points**:
- Wants to assess assets and prices without committing to registration first

**User Journey**: Land on login/landing → view public Market catalog + prices → decide to Create account

**Implications**: Market catalog (FR-05) viewable without login; all trading/portfolio/wallet actions require auth and redirect to login. No PII exposure.

---

## Admin (Back-office)

**Role**: Operations user maintaining asset master data and prices

**Goals**:
- Keep funds/bonds/gold master and NAV/prices correct; trigger scheduled events; review audit log

**Pain Points**:
- Needs a controlled way to update master data and publish prices/NAV without touching the DB directly

**User Journey**: Log in (admin role) → manage asset/price master → trigger DCA/coupon/maturity runs (dev "run now") → view audit log

**Implications**: Separate admin role (RBAC), all changes audited (NFR-02). Phase 1 (D1-1): admin surface simplified — a minimal admin/dev area for seeding, price publish, and manual scheduler triggers rather than a full back-office UI.

---

## Design Implications

- **Architecture**: Role-based access control (Investor / Admin); guest-readable public catalog endpoints; auth guard on all account/trading/portfolio/wallet routes.
- **UI/UX**: Modern, friendly red-and-white theme (white surfaces, red primary actions, soft red tints, black text; Add-on 2026-11-12c, was navy), responsive desktop + mobile, Thai default with i18n-ready structure, proper THB/unit/date formatting.
- **Data & Privacy**: PDPA consent capture at registration, encrypt sensitive PII at rest, mask sensitive data on screen, audit log for consequential actions; transactions immutable (BR-14).
