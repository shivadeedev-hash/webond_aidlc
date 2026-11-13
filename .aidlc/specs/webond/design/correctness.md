# Correctness Properties

## Overview
**PBT Framework**: gopter (Go). Scope per D3-11: core money math invariants. Money uses shopspring/decimal with DECIMAL(18,4) semantics.

---

## Properties

### P1 — MF unit allotment is exact to 4 dp
**Validates**: US-013
**Property**: For amount > 0 and NAV > 0, `units = round(amount / NAV, 4)` and `units * NAV ≈ amount` within one unit of the 4-dp quantum. No negative units; zero NAV rejected.
**Generators**: amount ∈ [1, 10,000,000] (2dp), NAV ∈ [0.0001, 10,000] (4dp).
**Edge cases**: tiny NAV, amount not divisible, rounding half-up.

### P2 — Cash conservation on reserve/release
**Validates**: US-010, US-013, US-027
**Property**: After reserve(x) then release(x) (or settle), `balance` and `reserved` return to consistent values; invariants hold always: `reserved ≥ 0`, `balance ≥ 0`, `reserved ≤ balance`, `available = balance − reserved ≥ 0`.
**Generators**: sequences of deposit/withdraw/reserve/release ops.
**Edge cases**: release more than reserved (must reject), concurrent-like interleavings in a single transaction.

### P3 — Average-cost & P/L correctness
**Validates**: US-022
**Property**: After a sequence of buys at various prices, `avg_cost = total_cost / quantity`; a subsequent sell reduces quantity and `total_cost` proportionally by avg_cost (realized P/L = proceeds − avg_cost*soldQty); unrealized P/L = marketValue − total_cost. Quantities never negative; selling more than held rejected (US-014/016/017).
**Generators**: buy/sell sequences with positive quantities/prices.
**Edge cases**: full liquidation (quantity→0, total_cost→0), sell-before-buy.

### P4 — Gold grams ↔ baht-weight & BID/ASK side
**Validates**: US-017
**Property**: `grams = bahtWeight * 15.244` and inverse round-trips within 4-dp tolerance; Buy total uses ASK, Sell proceeds use BID; grams stored 4dp, never negative.
**Generators**: weight ∈ [0.0001, 1000] g; bid ≤ ask prices.

### P5 — Bond coupon math & lot rule
**Validates**: US-015, US-023 (coupon events)
**Property**: `coupon = faceValue * couponRate / frequency`; bond buy amount must be ≥ 100,000 and an exact multiple of 100,000 (Add-on) else rejected; holding face value = units * faceValue.
**Generators**: faceValue=100000; couponRate ∈ [0, 0.1]; frequency ∈ {1,2,4}; amounts both valid multiples and invalid.

### P6 — Signed amount & display rule
**Validates**: US-024
**Property**: For any transaction type, the signed `amount` matches BR-15 (Sell/Coupon/Maturity/Deposit = +, Buy/DCA Buy/Withdraw = −); formatting yields THB 2dp, units/grams 4dp.
**Generators**: all txn types × random magnitudes.

## Test Configuration
- **Tests per property**: 300 (gopter default range), shrinking enabled (log minimal counterexample).
- **Timeout**: 10s per property.
- **Run**: `go test ./internal/money/... ./internal/service/... -run Property`
- **Organization**:
```
api/internal/money/money_property_test.go
api/internal/service/trading_property_test.go
api/internal/service/portfolio_property_test.go
```
