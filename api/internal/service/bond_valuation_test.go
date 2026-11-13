package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// Regression for the bond pricing-unit bug (Task 10.2). Real demo data: face
// value 1,000 THB/unit, clean price quoted as % of par (100.5), accrued
// interest in THB per unit (8.125 = half of the 16.25 semi-annual coupon).
// Holdings are counted in face-value units (units = amount / face), so a
// 100,000 THB purchase is 100 units and must be worth 100 x 1,005 = 100,500.
// The old code multiplied the units by the raw percentage (100 x 100.5 =
// 10,050) and also under-paid sell proceeds ~10x.

// bondFixture buys 100,000 THB of a fresh bond at par and then moves its quote
// to clean 100.5 / accrued 8.125 (the demo LB28DA numbers).
func bondFixture(t *testing.T) (ctx context.Context, userID, bondID uint64, ts *TradingService, ps *PortfolioService, ws *WalletService, clock Clock) {
	t.Helper()
	store := openTestStore(t)
	pool := rawPool(t)
	fixed := testClock()
	ws = NewWalletService(store, fixed)
	notify := NewNotificationService(store, fixed)
	ts = NewTradingService(store, ws, notify, fixed)
	ps = NewPortfolioService(store, fixed)
	ctx = context.Background()

	userID = seedActiveUser(t, store, fixed, decimal.RequireFromString("1000000"), 5)
	bondID = seedBond(t, store, fixed, "BV"+shortID(), 3)
	t.Cleanup(func() {
		cleanupUser(t, pool, userID)
		cleanupAsset(t, pool, "bond", bondID)
	})

	if _, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "bond", AssetID: bondID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("100000"),
	}); err != nil {
		t.Fatalf("buy bond: %v", err)
	}
	if err := store.Bonds.UpsertPrice(ctx, db.UpsertBondPriceParams{
		BondID: bondID, PriceDate: fixed.Now(),
		CleanPrice:      decimal.RequireFromString("100.5"),
		AccruedInterest: decimal.RequireFromString("8.125"),
	}); err != nil {
		t.Fatalf("move bond quote: %v", err)
	}
	return ctx, userID, bondID, ts, ps, ws, fixed
}

func findHolding(t *testing.T, ps *PortfolioService, ctx context.Context, userID, assetID uint64) HoldingView {
	t.Helper()
	pf, err := ps.Get(ctx, userID)
	if err != nil {
		t.Fatalf("portfolio: %v", err)
	}
	for _, h := range pf.Holdings {
		if h.AssetType == "bond" && h.AssetID == assetID {
			return h
		}
	}
	t.Fatalf("bond holding %d not found in %+v", assetID, pf.Holdings)
	return HoldingView{}
}

func TestBondValuation_PortfolioUsesThbPerUnit(t *testing.T) {
	ctx, userID, bondID, _, ps, _, _ := bondFixture(t)

	h := findHolding(t, ps, ctx, userID, bondID)
	if h.Quantity != "100.0000" {
		t.Fatalf("quantity = %s, want 100.0000 units (100,000 / 1,000 face)", h.Quantity)
	}
	if h.MarketPrice != "1005.0000" {
		t.Errorf("market price per unit = %s, want 1005.0000 THB (1,000 x 100.5%%)", h.MarketPrice)
	}
	if h.MarketValue != "100500.00" {
		t.Errorf("market value = %s, want 100500.00", h.MarketValue)
	}
	if h.UnrealizedPL != "500.00" {
		t.Errorf("unrealized P/L = %s, want 500.00 (+0.5%% over par cost)", h.UnrealizedPL)
	}
	if h.UnrealizedPct != "0.50" {
		t.Errorf("unrealized pct = %s, want 0.50", h.UnrealizedPct)
	}
	// Accrued interest is shown separately: 100 units x 8.125 THB/unit.
	if h.AccruedInt != "812.5000" {
		t.Errorf("accrued interest = %s, want 812.5000", h.AccruedInt)
	}
}

func TestBondValuation_SellProceedsUseCleanPlusAccruedPerUnit(t *testing.T) {
	ctx, userID, bondID, ts, ps, ws, _ := bondFixture(t)

	before, err := ws.Get(ctx, userID)
	if err != nil {
		t.Fatalf("wallet before: %v", err)
	}

	// Sell 10 units. Cash price per unit = 1,005 (clean in THB) + 8.125 accrued
	// = 1,013.125, so proceeds = 10,131.25 and, against a par cost of 10,000,
	// the realized P/L is 131.25.
	res, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "bond", AssetID: bondID, Side: "sell",
		OrderMode: "units", Units: decimal.RequireFromString("10"),
	})
	if err != nil {
		t.Fatalf("sell bond: %v", err)
	}
	if res.RealizedPL != "131.2500" {
		t.Errorf("realized P/L = %s, want 131.2500", res.RealizedPL)
	}
	if res.NetAmount != "10131.25" {
		t.Errorf("net amount = %s, want 10131.25", res.NetAmount)
	}

	after, err := ws.Get(ctx, userID)
	if err != nil {
		t.Fatalf("wallet after: %v", err)
	}
	if got := after.Balance.Sub(before.Balance); !got.Equal(decimal.RequireFromString("10131.25")) {
		t.Errorf("wallet credit = %s, want 10131.25", got)
	}

	// 90 units remain, still valued at 1,005 each.
	h := findHolding(t, ps, ctx, userID, bondID)
	if h.Quantity != "90.0000" || h.MarketValue != "90450.00" {
		t.Errorf("after sell: qty=%s mv=%s, want 90.0000 / 90450.00", h.Quantity, h.MarketValue)
	}
}

func TestBondValuation_DashboardTotalsAndTrendUseThbPerUnit(t *testing.T) {
	ctx, userID, _, _, ps, ws, clock := bondFixture(t)
	store := openTestStore(t)
	dash := NewDashboardService(store, ps, ws, clock)

	d, err := dash.Get(ctx, userID, "1M")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	// Allocation: the only holding is the bond.
	if len(d.Allocation) != 1 || d.Allocation[0].AssetType != "bond" || d.Allocation[0].Pct != "100.00" {
		t.Fatalf("allocation = %+v, want a single bond slice at 100.00", d.Allocation)
	}
	// The trend's latest bond point must be 100 units x 1,005 THB, not 100 x 100.5.
	var last TrendPoint
	for _, p := range d.Trend {
		last = p
	}
	if last.Value == "" {
		t.Fatalf("trend is empty: %+v", d.Trend)
	}
	got, err := decimal.NewFromString(last.Value)
	if err != nil {
		t.Fatalf("trend value %q: %v", last.Value, err)
	}
	if got.LessThan(decimal.RequireFromString("100000")) {
		t.Errorf("latest trend value = %s, want about 100,500 (bond valued in THB per unit)", last.Value)
	}
}
