package money

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestAllotUnits_Example(t *testing.T) {
	// 10,000 / 15.2341 = 656.42... → 4dp.
	units, err := AllotUnits(d("10000"), d("15.2341"))
	require.NoError(t, err)
	assert.Equal(t, "656.4221", units.StringFixed(4))
}

func TestGoldConversion_ExactConstant(t *testing.T) {
	// 1 baht-weight = 15.244 g exactly.
	assert.Equal(t, "15.2440", GramsFromBahtWeight(d("1")).StringFixed(4))
	// 15.244 g → 1 baht-weight.
	assert.Equal(t, "1.0000", BahtWeightFromGrams(d("15.244")).StringFixed(4))
}

func TestGoldSide(t *testing.T) {
	bid, ask := d("41000"), d("41250")
	assert.Equal(t, ask, GoldUnitPrice(GoldBuy, bid, ask))
	assert.Equal(t, bid, GoldUnitPrice(GoldSell, bid, ask))
}

func TestCouponPerUnit(t *testing.T) {
	// face 1000, rate 3% (0.03), semiannual (freq 2) → 15.0000 per period.
	c, err := CouponPerUnit(d("1000"), d("0.03"), 2)
	require.NoError(t, err)
	assert.Equal(t, "15.0000", c.StringFixed(4))

	_, err = CouponPerUnit(d("1000"), d("0.03"), 0)
	assert.ErrorIs(t, err, ErrNonPositiveFrequency)
}

func TestValidateBondLot(t *testing.T) {
	minLot := d("100000")
	assert.NoError(t, ValidateBondLot(d("100000"), minLot))
	assert.NoError(t, ValidateBondLot(d("300000"), minLot))
	assert.ErrorIs(t, ValidateBondLot(d("150000"), minLot), ErrBondLotInvalid)
	assert.ErrorIs(t, ValidateBondLot(d("0"), minLot), ErrBondLotInvalid)
	assert.ErrorIs(t, ValidateBondLot(d("-100000"), minLot), ErrBondLotInvalid)
}

func TestAvgCost_BuyThenSell(t *testing.T) {
	// Buy 100 @ 10 → qty 100, total 1000, avg 10.
	r, err := ApplyBuy(decimal.Zero, decimal.Zero, d("100"), d("10"))
	require.NoError(t, err)
	assert.Equal(t, "100.0000", r.Quantity.StringFixed(4))
	assert.Equal(t, "1000.0000", r.TotalCost.StringFixed(4))
	assert.Equal(t, "10.0000", r.AvgCost.StringFixed(4))

	// Buy 100 @ 20 → qty 200, total 3000, avg 15.
	r, err = ApplyBuy(r.Quantity, r.TotalCost, d("100"), d("20"))
	require.NoError(t, err)
	assert.Equal(t, "200.0000", r.Quantity.StringFixed(4))
	assert.Equal(t, "3000.0000", r.TotalCost.StringFixed(4))
	assert.Equal(t, "15.0000", r.AvgCost.StringFixed(4))

	// Sell 50 @ 25 → realized = 50*(25-15) = 500; qty 150; total 2250.
	r, err = ApplySell(r.Quantity, r.TotalCost, d("50"), d("25"))
	require.NoError(t, err)
	assert.Equal(t, "150.0000", r.Quantity.StringFixed(4))
	assert.Equal(t, "2250.0000", r.TotalCost.StringFixed(4))
	assert.Equal(t, "500.0000", r.RealizUS.StringFixed(4))
}

func TestAvgCost_FullLiquidation(t *testing.T) {
	r, err := ApplyBuy(decimal.Zero, decimal.Zero, d("10"), d("50"))
	require.NoError(t, err)
	// Sell everything → qty 0, total 0, avg 0.
	r, err = ApplySell(r.Quantity, r.TotalCost, d("10"), d("55"))
	require.NoError(t, err)
	assert.True(t, r.Quantity.IsZero())
	assert.True(t, r.TotalCost.IsZero())
	assert.True(t, r.AvgCost.IsZero())
	assert.Equal(t, "50.0000", r.RealizUS.StringFixed(4)) // 10*(55-50)
}

func TestApplySell_ExceedsHolding(t *testing.T) {
	_, err := ApplySell(d("10"), d("100"), d("11"), d("10"))
	assert.ErrorIs(t, err, ErrSellExceedsHolding)
}

func TestSignedAmount(t *testing.T) {
	mag := d("1000")
	assert.True(t, SignedAmount(TxnDeposit, mag).IsPositive())
	assert.True(t, SignedAmount(TxnSell, mag).IsPositive())
	assert.True(t, SignedAmount(TxnCoupon, mag).IsPositive())
	assert.True(t, SignedAmount(TxnMaturity, mag).IsPositive())
	assert.True(t, SignedAmount(TxnBuy, mag).IsNegative())
	assert.True(t, SignedAmount(TxnDcaBuy, mag).IsNegative())
	assert.True(t, SignedAmount(TxnWithdraw, mag).IsNegative())
}

func TestBondUnitPrice(t *testing.T) {
	face := decimal.RequireFromString("1000")
	cases := map[string]string{
		"100":    "1000.0000", // at par
		"100.5":  "1005.0000", // demo LB28DA
		"99.2":   "992.0000",  // demo CP234A
		"98.765": "987.6500",
	}
	for pct, want := range cases {
		if got := BondUnitPrice(face, decimal.RequireFromString(pct)); got.StringFixed(4) != want {
			t.Errorf("BondUnitPrice(1000, %s) = %s, want %s", pct, got.StringFixed(4), want)
		}
	}
	// A different face value scales linearly.
	if got := BondUnitPrice(decimal.RequireFromString("100"), decimal.RequireFromString("101")); got.StringFixed(4) != "101.0000" {
		t.Errorf("BondUnitPrice(100, 101) = %s, want 101.0000", got.StringFixed(4))
	}
}
