package money

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/shopspring/decimal"
)

// Property-based tests for the core money math (correctness.md P1/P4/P5/P6).
// Framework: gopter. 300 runs per property, shrinking enabled.

func testParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 300
	return params
}

// dec4 builds a 4-dp decimal from scaled integer units (value = n / 10000).
func dec4(n int64) decimal.Decimal {
	return decimal.New(n, -Scale)
}

// dec2 builds a 2-dp decimal from scaled integer units (value = n / 100).
func dec2(n int64) decimal.Decimal {
	return decimal.New(n, -2)
}

// quantum is the 4-dp quantum (0.0001) used as the rounding tolerance.
var quantum = dec4(1)

// --- P1: MF unit allotment is exact to 4 dp (US-013) ---

func TestProperty_P1_AllotUnits(t *testing.T) {
	properties := gopter.NewProperties(testParameters())

	properties.Property("units = round(amount/NAV, 4); units*NAV ≈ amount within one quantum; units ≥ 0", prop.ForAll(
		func(amountCents int64, navUnits int64) bool {
			amount := dec2(amountCents) // amount in [1, 10,000,000], 2dp
			nav := dec4(navUnits)       // nav in [0.0001, 10,000], 4dp

			units, err := AllotUnits(amount, nav)
			if err != nil {
				return false
			}
			if units.IsNegative() {
				return false
			}
			// units must be exactly 4 dp.
			if units.Exponent() < -Scale {
				return false
			}
			// Reconstruction: units*NAV should be within one 4-dp quantum of amount,
			// scaled by NAV (the rounding error on units is < 0.5 quantum, so the
			// amount error is < NAV * quantum). Use a generous bound of NAV*quantum.
			reconstructed := units.Mul(nav)
			diff := reconstructed.Sub(amount).Abs()
			tolerance := nav.Mul(quantum)
			return diff.LessThanOrEqual(tolerance)
		},
		gen.Int64Range(100, 1_000_000_000), // amountCents: 1 .. 10,000,000 (2dp)
		gen.Int64Range(1, 100_000_000),     // navUnits: 0.0001 .. 10,000 (4dp)
	))

	properties.Property("zero or negative NAV is rejected", prop.ForAll(
		func(amountCents int64) bool {
			amount := dec2(amountCents)
			_, errZero := AllotUnits(amount, decimal.Zero)
			_, errNeg := AllotUnits(amount, dec4(-1))
			return errZero == ErrNonPositiveNAV && errNeg == ErrNonPositiveNAV
		},
		gen.Int64Range(100, 1_000_000_000),
	))

	properties.TestingRun(t)
}

// --- P4: Gold grams ↔ baht-weight & BID/ASK side (US-017) ---

func TestProperty_P4_GoldConversionAndSide(t *testing.T) {
	properties := gopter.NewProperties(testParameters())

	properties.Property("grams ↔ baht-weight round-trips within 4-dp tolerance; grams ≥ 0", prop.ForAll(
		func(gramUnits int64) bool {
			grams := dec4(gramUnits) // weight in [0.0001, 1000] g, 4dp
			bahtWeight := BahtWeightFromGrams(grams)
			back := GramsFromBahtWeight(bahtWeight)
			if back.IsNegative() {
				return false
			}
			// Round-trip error is bounded by the conversion constant times the
			// quantum (one rounding step each way).
			diff := back.Sub(grams).Abs()
			tolerance := gramsPerBahtWeight.Mul(quantum)
			return diff.LessThanOrEqual(tolerance)
		},
		gen.Int64Range(1, 10_000_000), // 0.0001 .. 1000 g (4dp)
	))

	properties.Property("Buy uses ASK, Sell uses BID (bid ≤ ask)", prop.ForAll(
		func(bidUnits int64, spreadUnits int64, gramUnits int64) bool {
			bid := dec4(bidUnits)
			ask := dec4(bidUnits + spreadUnits) // ensure ask ≥ bid
			grams := dec4(gramUnits)

			buyVal := GoldTradeValue(GoldBuy, grams, bid, ask)
			sellVal := GoldTradeValue(GoldSell, grams, bid, ask)

			wantBuy := Quantize(grams.Mul(ask))
			wantSell := Quantize(grams.Mul(bid))
			if !buyVal.Equal(wantBuy) || !sellVal.Equal(wantSell) {
				return false
			}
			// Buy value ≥ Sell value since ask ≥ bid and grams ≥ 0.
			return buyVal.GreaterThanOrEqual(sellVal)
		},
		gen.Int64Range(1, 500_000_000), // bid
		gen.Int64Range(0, 10_000_000),  // spread (ask-bid) ≥ 0
		gen.Int64Range(1, 10_000_000),  // grams
	))

	properties.TestingRun(t)
}

// --- P5: Bond coupon math & lot rule (US-015, US-023) ---

func TestProperty_P5_BondCouponAndLot(t *testing.T) {
	properties := gopter.NewProperties(testParameters())

	const minLotInt = 100_000
	minLot := decimal.NewFromInt(minLotInt)
	faceValue := decimal.NewFromInt(100_000)

	properties.Property("coupon = faceValue*couponRate/frequency, 4dp; frequency>0", prop.ForAll(
		func(rateBps int64, freqIdx int) bool {
			// couponRate in [0, 0.1] as fraction; rateBps in [0, 1000] bps → /10000.
			couponRate := decimal.New(rateBps, -4)
			freqs := []int{1, 2, 4}
			frequency := freqs[freqIdx%len(freqs)]

			got, err := CouponPerUnit(faceValue, couponRate, frequency)
			if err != nil {
				return false
			}
			want := Quantize(faceValue.Mul(couponRate).DivRound(decimal.NewFromInt(int64(frequency)), Scale+8))
			return got.Equal(want) && got.Exponent() >= -Scale
		},
		gen.Int64Range(0, 1000), // 0 .. 0.1000
		gen.IntRange(0, 2),      // index into {1,2,4}
	))

	properties.Property("valid lot = positive exact multiple of 100,000 accepted; others rejected", prop.ForAll(
		func(multiple int64, remainder int64) bool {
			// Valid amount: multiple * 100,000 (multiple ≥ 1).
			validAmount := minLot.Mul(decimal.NewFromInt(multiple))
			if err := ValidateBondLot(validAmount, minLot); err != nil {
				return false
			}
			// Invalid amount: add a non-zero remainder in (0, 100000).
			if remainder == 0 {
				remainder = 1
			}
			invalidAmount := validAmount.Add(decimal.NewFromInt(remainder))
			return ValidateBondLot(invalidAmount, minLot) == ErrBondLotInvalid
		},
		gen.Int64Range(1, 100),         // multiple ≥ 1
		gen.Int64Range(1, minLotInt-1), // remainder in (0, 100000)
	))

	properties.Property("zero and negative amounts are rejected", prop.ForAll(
		func(neg int64) bool {
			zeroErr := ValidateBondLot(decimal.Zero, minLot)
			negErr := ValidateBondLot(decimal.NewFromInt(-neg), minLot)
			return zeroErr == ErrBondLotInvalid && negErr == ErrBondLotInvalid
		},
		gen.Int64Range(1, minLotInt),
	))

	properties.Property("holding face value = units * faceValue (4dp)", prop.ForAll(
		func(unitUnits int64) bool {
			units := dec4(unitUnits)
			got := BondFaceValue(units, faceValue)
			want := Quantize(units.Mul(faceValue))
			return got.Equal(want)
		},
		gen.Int64Range(0, 10_000_000),
	))

	properties.TestingRun(t)
}

// --- P6: Signed amount & display rule (US-024) ---

func TestProperty_P6_SignedAmountAndDisplay(t *testing.T) {
	properties := gopter.NewProperties(testParameters())

	credits := map[TxnType]bool{TxnSell: true, TxnCoupon: true, TxnMaturity: true, TxnDeposit: true}
	allTypes := []TxnType{TxnBuy, TxnSell, TxnDcaBuy, TxnCoupon, TxnMaturity, TxnDeposit, TxnWithdraw}

	properties.Property("sign matches BR-15 for every txn type and magnitude", prop.ForAll(
		func(typeIdx int, magCents int64) bool {
			t := allTypes[typeIdx%len(allTypes)]
			mag := dec2(magCents) // non-negative magnitude
			signed := SignedAmount(t, mag)

			if mag.IsZero() {
				return signed.IsZero()
			}
			if credits[t] {
				return signed.Equal(mag) && signed.IsPositive()
			}
			return signed.Equal(mag.Neg()) && signed.IsNegative()
		},
		gen.IntRange(0, len(allTypes)-1),
		gen.Int64Range(0, 1_000_000_000),
	))

	properties.Property("sign rule ignores the input sign (uses absolute value)", prop.ForAll(
		func(typeIdx int, magCents int64) bool {
			t := allTypes[typeIdx%len(allTypes)]
			mag := dec2(magCents)
			return SignedAmount(t, mag).Equal(SignedAmount(t, mag.Neg()))
		},
		gen.IntRange(0, len(allTypes)-1),
		gen.Int64Range(1, 1_000_000_000),
	))

	properties.Property("THB formatting yields 2dp, quantity formatting 4dp", prop.ForAll(
		func(magCents int64) bool {
			amt := dec2(magCents)
			thb := FormatTHB(amt)
			q := FormatQuantity(amt)
			// 2dp has exactly 2 chars after the dot; 4dp has exactly 4.
			return decimalPlaces(thb) == 2 && decimalPlaces(q) == 4
		},
		gen.Int64Range(0, 1_000_000_000),
	))

	properties.TestingRun(t)
}

// decimalPlaces counts the digits after the decimal point in a fixed string.
func decimalPlaces(s string) int {
	dot := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return 0
	}
	return len(s) - dot - 1
}
