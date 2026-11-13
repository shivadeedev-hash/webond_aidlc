// Package money holds the core monetary math for WeBond using
// shopspring/decimal with DECIMAL(18,4) semantics — never float (NFR-05,
// correctness.md). All money/unit quantities are quantized to 4 decimal places;
// display formatting applies THB 2dp for amounts and 4dp for units/grams.
package money

import (
	"errors"

	"github.com/shopspring/decimal"
)

// Scale is the stored decimal scale for money and unit columns: DECIMAL(18,4).
const Scale int32 = 4

// gramsPerBahtWeight is the Thai gold conversion constant: 1 baht-weight =
// 15.244 g (correctness.md P4).
var gramsPerBahtWeight = decimal.RequireFromString("15.244")

// Sentinel errors returned by the helpers. Services map these to typed API
// errors (domain.AppError) at their boundary.
var (
	// ErrNonPositiveNAV is returned when a NAV/price divisor is zero or negative.
	ErrNonPositiveNAV = errors.New("nav must be positive")
	// ErrNonPositiveAmount is returned when an amount must be > 0 but is not.
	ErrNonPositiveAmount = errors.New("amount must be positive")
	// ErrNonPositiveFrequency is returned when coupon frequency is not > 0.
	ErrNonPositiveFrequency = errors.New("coupon frequency must be positive")
	// ErrBondLotInvalid is returned when a bond amount is not a positive exact
	// multiple of the minimum lot (Add-on: 100,000).
	ErrBondLotInvalid = errors.New("bond amount must be a positive multiple of the minimum lot")
	// ErrSellExceedsHolding is returned when a sell quantity exceeds the holding.
	ErrSellExceedsHolding = errors.New("sell quantity exceeds holding")
	// ErrNegativeQuantity is returned when a quantity would become negative.
	ErrNegativeQuantity = errors.New("quantity must not be negative")
)

// Quantize rounds a decimal to the stored 4-dp quantum using banker-free
// half-up rounding (decimal.Round), matching DECIMAL(18,4) storage.
func Quantize(d decimal.Decimal) decimal.Decimal {
	return d.Round(Scale)
}

// AllotUnits computes mutual-fund units for a purchase amount at a given NAV:
// units = round(amount / NAV, 4) (correctness.md P1, US-013). NAV must be > 0
// and amount must be > 0; units are never negative.
func AllotUnits(amount, nav decimal.Decimal) (decimal.Decimal, error) {
	if nav.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrNonPositiveNAV
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrNonPositiveAmount
	}
	// DivisionPrecision is global on shopspring/decimal; divide at higher
	// precision then quantize to the 4-dp storage scale.
	units := amount.DivRound(nav, Scale+8)
	return Quantize(units), nil
}

// GramsFromBahtWeight converts a baht-weight gold quantity to grams:
// grams = bahtWeight * 15.244, quantized to 4 dp (correctness.md P4).
func GramsFromBahtWeight(bahtWeight decimal.Decimal) decimal.Decimal {
	return Quantize(bahtWeight.Mul(gramsPerBahtWeight))
}

// BahtWeightFromGrams converts grams to baht-weight:
// bahtWeight = grams / 15.244, quantized to 4 dp (correctness.md P4).
func BahtWeightFromGrams(grams decimal.Decimal) decimal.Decimal {
	return Quantize(grams.DivRound(gramsPerBahtWeight, Scale+8))
}

// GoldSide selects the price leg for a gold trade: Buy pays ASK, Sell receives
// BID (BR-09, correctness.md P4).
type GoldSide int

const (
	// GoldBuy uses the ASK price.
	GoldBuy GoldSide = iota
	// GoldSell uses the BID price.
	GoldSell
)

// GoldUnitPrice returns the per-gram price to use for a gold trade side:
// ASK for a buy, BID for a sell (BR-09).
func GoldUnitPrice(side GoldSide, bid, ask decimal.Decimal) decimal.Decimal {
	if side == GoldBuy {
		return ask
	}
	return bid
}

// GoldTradeValue computes the quantized cash value of a gold trade for a given
// weight in grams, using the side-appropriate per-gram price (BR-09).
func GoldTradeValue(side GoldSide, grams, bid, ask decimal.Decimal) decimal.Decimal {
	return Quantize(grams.Mul(GoldUnitPrice(side, bid, ask)))
}

// CouponPerUnit computes a single bond coupon per unit (per face value):
// coupon = faceValue * couponRate / frequency, quantized to 4 dp
// (correctness.md P5, US-015/US-023). couponRate is the annual rate as a
// fraction (e.g. 0.03 for 3%); frequency is payments per year.
func CouponPerUnit(faceValue, couponRate decimal.Decimal, frequency int) (decimal.Decimal, error) {
	if frequency <= 0 {
		return decimal.Zero, ErrNonPositiveFrequency
	}
	annual := faceValue.Mul(couponRate)
	perPeriod := annual.DivRound(decimal.NewFromInt(int64(frequency)), Scale+8)
	return Quantize(perPeriod), nil
}

// BondParPercent is the bond clean price at par: bond prices are quoted as a
// percentage of face value, so 100 means "trading at par".
var BondParPercent = decimal.NewFromInt(100)

// BondUnitPrice converts a bond clean price quoted as a percentage of par
// (e.g. 100.5) into THB per unit: faceValue * cleanPricePct / 100, quantized to
// 4 dp. accrued_interest in bond_prices is already THB per unit, so a dirty
// (cash) price per unit is BondUnitPrice(...) + accrued. Holdings are counted
// in face-value units (units = amount / faceValue), so any value computed as
// quantity * price MUST use this THB-per-unit figure, not the raw percentage.
func BondUnitPrice(faceValue, cleanPricePct decimal.Decimal) decimal.Decimal {
	return Quantize(faceValue.Mul(cleanPricePct).DivRound(BondParPercent, Scale+8))
}

// ValidateBondLot verifies a bond buy amount is a positive exact multiple of
// minLot (Add-on: 100,000) (correctness.md P5, BR-04 override). Returns
// ErrBondLotInvalid otherwise.
func ValidateBondLot(amount, minLot decimal.Decimal) error {
	if minLot.LessThanOrEqual(decimal.Zero) {
		return ErrBondLotInvalid
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrBondLotInvalid
	}
	// amount must be an exact integer multiple of minLot.
	if !amount.Mod(minLot).Equal(decimal.Zero) {
		return ErrBondLotInvalid
	}
	return nil
}

// BondFaceValue computes the total face value held for a number of bond units:
// units * faceValue (correctness.md P5).
func BondFaceValue(units, faceValue decimal.Decimal) decimal.Decimal {
	return Quantize(units.Mul(faceValue))
}

// AvgCostResult is the outcome of applying a buy/sell to a position (BR-10).
type AvgCostResult struct {
	Quantity  decimal.Decimal // resulting quantity (4 dp)
	TotalCost decimal.Decimal // resulting total cost (4 dp)
	AvgCost   decimal.Decimal // total_cost / quantity (4 dp; zero when flat)
	RealizUS  decimal.Decimal // realized P/L on a sell (0 on a buy)
}

// ApplyBuy adds a purchase of qty units at unitPrice to a position and returns
// the new weighted average cost (BR-10, correctness.md P3):
// total_cost += qty*unitPrice; avg_cost = total_cost / quantity.
func ApplyBuy(quantity, totalCost, qty, unitPrice decimal.Decimal) (AvgCostResult, error) {
	if qty.LessThanOrEqual(decimal.Zero) {
		return AvgCostResult{}, ErrNonPositiveAmount
	}
	newQty := Quantize(quantity.Add(qty))
	newTotal := Quantize(totalCost.Add(qty.Mul(unitPrice)))
	return AvgCostResult{
		Quantity:  newQty,
		TotalCost: newTotal,
		AvgCost:   avgCost(newTotal, newQty),
		RealizUS:  decimal.Zero,
	}, nil
}

// ApplySell removes qty units from a position at avg cost, reducing total_cost
// proportionally, and computes realized P/L = proceeds − avgCost*qty (BR-10,
// correctness.md P3). Selling more than held is rejected.
func ApplySell(quantity, totalCost, qty, unitPrice decimal.Decimal) (AvgCostResult, error) {
	if qty.LessThanOrEqual(decimal.Zero) {
		return AvgCostResult{}, ErrNonPositiveAmount
	}
	if qty.GreaterThan(quantity) {
		return AvgCostResult{}, ErrSellExceedsHolding
	}
	avg := avgCost(totalCost, quantity)
	newQty := Quantize(quantity.Sub(qty))
	// Reduce total cost by the average cost of the sold quantity; on full
	// liquidation the remaining total_cost is exactly zero.
	var newTotal decimal.Decimal
	if newQty.Equal(decimal.Zero) {
		newTotal = decimal.Zero
	} else {
		newTotal = Quantize(totalCost.Sub(avg.Mul(qty)))
	}
	proceeds := qty.Mul(unitPrice)
	realized := Quantize(proceeds.Sub(avg.Mul(qty)))
	return AvgCostResult{
		Quantity:  newQty,
		TotalCost: newTotal,
		AvgCost:   avgCost(newTotal, newQty),
		RealizUS:  realized,
	}, nil
}

// avgCost returns total/quantity quantized to 4 dp, or zero when the position
// is flat (avoids a divide-by-zero on full liquidation).
func avgCost(total, quantity decimal.Decimal) decimal.Decimal {
	if quantity.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return Quantize(total.DivRound(quantity, Scale+8))
}

// TxnType enumerates transaction types for the signed-amount rule (BR-15).
type TxnType string

const (
	TxnBuy      TxnType = "buy"
	TxnSell     TxnType = "sell"
	TxnDcaBuy   TxnType = "dca_buy"
	TxnCoupon   TxnType = "coupon"
	TxnMaturity TxnType = "maturity"
	TxnDeposit  TxnType = "deposit"
	TxnWithdraw TxnType = "withdraw"
)

// SignedAmount applies the BR-15 sign convention to a non-negative magnitude:
// Sell / Coupon / Maturity / Deposit are credits (+); Buy / DCA Buy / Withdraw
// are debits (−) (correctness.md P6, US-024). The magnitude's own sign is
// ignored (its absolute value is used) so callers can pass raw amounts.
func SignedAmount(t TxnType, magnitude decimal.Decimal) decimal.Decimal {
	mag := magnitude.Abs()
	switch t {
	case TxnSell, TxnCoupon, TxnMaturity, TxnDeposit:
		return mag
	case TxnBuy, TxnDcaBuy, TxnWithdraw:
		return mag.Neg()
	default:
		return mag
	}
}

// FormatTHB formats a monetary amount for display with 2 decimal places
// (BR-15 — THB 2dp). The returned string is a plain decimal (no thousands
// separators; the UI applies locale grouping).
func FormatTHB(amount decimal.Decimal) string {
	return amount.StringFixed(2)
}

// FormatQuantity formats a unit/gram quantity for display with 4 decimal
// places (BR-15 — units/grams 4dp).
func FormatQuantity(q decimal.Decimal) string {
	return q.StringFixed(Scale)
}

// String is a convenience for serializing a DECIMAL to the API as a 4-dp string
// (api-spec.md § Money — strings in JSON to avoid float loss).
func String(d decimal.Decimal) string {
	return d.StringFixed(Scale)
}
