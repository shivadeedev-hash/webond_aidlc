package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// TestPortfolio_ValuationAndUnrealizedPL buys a fund and gold, then checks the
// portfolio valuation + unrealized P/L after a price move (portfolio.go Get/
// valueHolding; US-022, BR-10).
func TestPortfolio_ValuationAndUnrealizedPL(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	ps := NewPortfolioService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "PF"+sfx, decimal.RequireFromString("10"), 3)
	goldID := seedGold(t, store, clock, "PG"+sfx, decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	t.Cleanup(func() {
		cleanupUser(t, pool, userID)
		cleanupAsset(t, pool, "fund", fundID)
		cleanupAsset(t, pool, "gold", goldID)
	})

	// Buy 5000 THB of the fund at NAV 10 → 500 units, cost 5000.
	if _, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	}); err != nil {
		t.Fatalf("buy fund: %v", err)
	}
	// Buy gold for 41250 THB at ASK 41250 → 1 gram, cost 41250.
	gq, err := ts.Quote(ctx, userID, goldID, "buy", decimal.RequireFromString("41250"), decimal.Zero)
	if err != nil {
		t.Fatalf("gold quote: %v", err)
	}
	if _, err := ts.ConfirmQuote(ctx, userID, gq.QuoteID, ""); err != nil {
		t.Fatalf("gold confirm: %v", err)
	}

	// Move fund NAV up to 12 → MV 6000, unrealized +1000 on the fund.
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
		FundID: fundID, NavDate: clock.Now(), Nav: decimal.RequireFromString("12"),
	}); err != nil {
		t.Fatalf("bump nav: %v", err)
	}

	pf, err := ps.Get(ctx, userID)
	if err != nil {
		t.Fatalf("portfolio get: %v", err)
	}
	if len(pf.Holdings) < 2 {
		t.Fatalf("expected >=2 holdings, got %d", len(pf.Holdings))
	}
	for _, h := range pf.Holdings {
		switch h.AssetType {
		case string(db.HoldingsAssetTypeFund):
			if h.Code != "PF"+sfx {
				continue
			}
			if h.MarketPrice != "12.0000" {
				t.Fatalf("fund market price=%s want 12.0000", h.MarketPrice)
			}
			// unrealized % = (6000-5000)/5000*100 = 20.00
			if h.UnrealizedPct != "20.00" {
				t.Fatalf("fund unrealized pct=%s want 20.00", h.UnrealizedPct)
			}
		case string(db.HoldingsAssetTypeGold):
			if h.Code != "PG"+sfx {
				continue
			}
			// Gold valued at BID 41000 → MV 41000, cost 41250 → loss.
			if h.MarketPrice != "41000.0000" {
				t.Fatalf("gold market price=%s want BID 41000.0000", h.MarketPrice)
			}
		}
	}
	// Summary unrealized P/L should be non-empty and reflect the combined move.
	if pf.Summary.UnrealizedPL == "" || pf.Summary.MarketValue == "" {
		t.Fatalf("empty summary: %+v", pf.Summary)
	}
}

// TestPortfolio_EmptyHoldings returns a zeroed portfolio for a user with no
// holdings (portfolio.go Get; US-022).
func TestPortfolio_EmptyHoldings(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ps := NewPortfolioService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	pf, err := ps.Get(ctx, userID)
	if err != nil {
		t.Fatalf("empty portfolio: %v", err)
	}
	if len(pf.Holdings) != 0 {
		t.Fatalf("expected no holdings, got %d", len(pf.Holdings))
	}
}

// TestDashboard_AggregatesAndTrend builds the dashboard for a user with a fund
// holding and NAV history, checking totals, allocation, top movers and the
// value-trend series (dashboard.go Get/allocation/topMovers/valueTrend; US-023).
func TestDashboard_AggregatesAndTrend(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	ps := NewPortfolioService(store, clock)
	dash := NewDashboardService(store, ps, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "DB"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	if _, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("10000"),
	}); err != nil {
		t.Fatalf("buy fund: %v", err)
	}
	// Historical NAV points inside the 6M trend window.
	for _, p := range []struct {
		monthsAgo int
		nav       string
	}{{5, "9"}, {3, "9.5"}, {1, "10.5"}} {
		if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
			FundID: fundID, NavDate: clock.Now().AddDate(0, -p.monthsAgo, 0), Nav: decimal.RequireFromString(p.nav),
		}); err != nil {
			t.Fatalf("seed trend nav: %v", err)
		}
	}

	d, err := dash.Get(ctx, userID, "6M")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if d.TotalValue == "" || d.CashAvailable == "" || d.LastUpdated == "" {
		t.Fatalf("dashboard missing scalar fields: %+v", d)
	}
	// Allocation must include the fund slice at 100% (only one asset type held).
	foundFund := false
	for _, a := range d.Allocation {
		if a.AssetType == "fund" {
			foundFund = true
			if a.Pct != "100.00" {
				t.Fatalf("fund allocation pct=%s want 100.00", a.Pct)
			}
		}
	}
	if !foundFund {
		t.Fatalf("allocation missing fund slice: %+v", d.Allocation)
	}
	// Top movers should include our fund code.
	moverSeen := false
	for _, m := range d.TopMovers {
		if m.Code == "DB"+sfx {
			moverSeen = true
		}
	}
	if !moverSeen {
		t.Fatalf("top movers missing fund: %+v", d.TopMovers)
	}
	// Trend has multiple points (the seeded history + latest).
	if len(d.Trend) < 3 {
		t.Fatalf("expected >=3 trend points, got %d", len(d.Trend))
	}
	if d.DCASummary.ActivePlans != 0 {
		t.Fatalf("no DCA plan expected, got %d active", d.DCASummary.ActivePlans)
	}
}

// TestDashboard_TrendUsesLastGoldTickPerDay is a regression test: with the live
// price generator gold has many ticks per day, and the trend used to add all of
// them up (today's point became ~150x the real value).
func TestDashboard_TrendUsesLastGoldTickPerDay(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ps := NewPortfolioService(store, clock)
	dash := NewDashboardService(store, ps, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	goldID := seedGold(t, store, clock, "TG"+shortID(), decimal.RequireFromString("100"), decimal.RequireFromString("101"))
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "gold", goldID) })

	if err := store.Holdings.Upsert(ctx, db.UpsertHoldingParams{
		UserID: userID, AssetType: db.HoldingsAssetTypeGold, AssetID: goldID,
		Quantity: decimal.RequireFromString("2"), LockedQuantity: decimal.Zero,
		AvgCost: decimal.RequireFromString("100"), TotalCost: decimal.RequireFromString("200"),
	}); err != nil {
		t.Fatalf("seed holding: %v", err)
	}
	// Three more ticks later the same day; the last one (bid 130) must win.
	for i, bid := range []string{"110", "120", "130"} {
		if _, err := store.Gold.CreatePrice(ctx, db.CreateGoldPriceParams{
			GoldProductID: goldID, PricedAt: clock.Now().Add(time.Duration(i+1) * time.Minute),
			BidPerGram: decimal.RequireFromString(bid), AskPerGram: decimal.RequireFromString(bid),
		}); err != nil {
			t.Fatalf("seed tick: %v", err)
		}
	}

	d, err := dash.Get(ctx, userID, "1M")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(d.Trend) != 1 {
		t.Fatalf("expected one trend day, got %d: %+v", len(d.Trend), d.Trend)
	}
	if got := d.Trend[0].Value; got != "260.00" {
		t.Fatalf("trend value = %s, want 260.00 (2 g x last bid 130), not the sum of all ticks", got)
	}
}
