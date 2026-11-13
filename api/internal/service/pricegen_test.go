package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

func quietLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// TestPriceGen_GetSetStatePersisted toggles the persisted state and restores it
// (pricegen.go GetState/SetState; US-030 Add-on, D3-7). The ticker is stopped
// so no background goroutine leaks into other tests.
func TestPriceGen_GetSetStatePersisted(t *testing.T) {
	store := openTestStore(t)
	clock := testClock()
	pg := NewPriceGeneratorService(store, quietLogger(), clock)
	ctx := context.Background()

	// Capture the original persisted state to restore after the test.
	orig, err := pg.GetState(ctx)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	t.Cleanup(func() {
		pg.Stop()
		vol, _ := decimal.NewFromString(orig.VolatilityPct)
		_, _ = pg.SetState(context.Background(), orig.Enabled, int(orig.IntervalSeconds), vol)
		pg.Stop()
	})

	// Enable with explicit interval/volatility.
	st, err := pg.SetState(ctx, true, 120, decimal.RequireFromString("0.75"))
	if err != nil {
		t.Fatalf("set state enable: %v", err)
	}
	if !st.Enabled || st.IntervalSeconds != 120 || st.VolatilityPct != "0.7500" {
		t.Fatalf("unexpected state after enable: %+v", st)
	}
	pg.Stop() // halt the ticker started by enable

	// Re-read from the DB to confirm persistence.
	got, err := pg.GetState(ctx)
	if err != nil {
		t.Fatalf("re-get state: %v", err)
	}
	if !got.Enabled || got.IntervalSeconds != 120 {
		t.Fatalf("state not persisted: %+v", got)
	}

	// Disable, and confirm defaults are applied for non-positive inputs.
	off, err := pg.SetState(ctx, false, 0, decimal.Zero)
	if err != nil {
		t.Fatalf("set state disable: %v", err)
	}
	if off.Enabled {
		t.Fatalf("expected disabled, got %+v", off)
	}
	if off.IntervalSeconds != 60 || off.VolatilityPct != "0.5000" {
		t.Fatalf("expected defaults (60s, 0.5), got %+v", off)
	}
}

// TestPriceGen_TickAppendsPrices runs a manual tick and confirms it appends a
// fresh gold price and a fund NAV (pricegen.go Tick/TickGold/PublishNav; US-030).
func TestPriceGen_TickAppendsPrices(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	pg := NewPriceGeneratorService(store, quietLogger(), clock)
	ctx := context.Background()

	sfx := shortID()
	goldID := seedGold(t, store, clock, "TK"+sfx, decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	fundID := seedFund(t, store, clock, "TKF"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupAsset(t, pool, "gold", goldID); cleanupAsset(t, pool, "fund", fundID) })

	// Advance the clock a day so PublishNav writes a NAV on a new date rather
	// than upserting onto the seeded one.
	clock.Advance(26 * 60 * 60 * 1e9)

	// TickGold appends a new bounded-walk price for every gold product.
	before, _ := store.Gold.GetLatestPrice(ctx, goldID)
	if err := pg.TickGold(ctx, decimal.RequireFromString("0.5")); err != nil {
		t.Fatalf("tick gold: %v", err)
	}
	after, _ := store.Gold.GetLatestPrice(ctx, goldID)
	if !after.PricedAt.After(before.PricedAt) {
		t.Fatalf("TickGold did not append a newer price (before=%s after=%s)", before.PricedAt, after.PricedAt)
	}
	if after.AskPerGram.LessThan(after.BidPerGram) {
		t.Fatalf("ask < bid after walk: bid=%s ask=%s", after.BidPerGram, after.AskPerGram)
	}

	// PublishNav writes a NAV for every fund on today's date.
	if err := pg.PublishNav(ctx, decimal.RequireFromString("0.5")); err != nil {
		t.Fatalf("publish nav: %v", err)
	}
	navs, _ := store.Funds.ListNavs(ctx, db.ListFundNavsParams{FundID: fundID, NavDate: clock.Now().AddDate(-1, 0, 0)})
	if len(navs) < 2 {
		t.Fatalf("expected >=2 NAV points after publish, got %d", len(navs))
	}
}
