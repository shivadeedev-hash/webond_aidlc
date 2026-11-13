package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
)

// TestCatalog_ListFilterSearch covers List with type filter, text search, risk
// filter and name sort (catalog.go List/filterAndSearch/sortAssets; US-011).
func TestCatalog_ListFilterSearch(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	cs := NewCatalogService(store, clock)
	ctx := context.Background()

	sfx := shortID()
	fundID := seedFund(t, store, clock, "CZ"+sfx, decimal.RequireFromString("12.3456"), 3)
	bondID := seedBond(t, store, clock, "CZB"+sfx, 4)
	goldID := seedGold(t, store, clock, "CZG"+sfx, decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	t.Cleanup(func() {
		cleanupAsset(t, pool, "fund", fundID)
		cleanupAsset(t, pool, "bond", bondID)
		cleanupAsset(t, pool, "gold", goldID)
	})

	// Type filter = fund only.
	funds, err := cs.List(ctx, ListParams{Type: "fund"})
	if err != nil {
		t.Fatalf("list funds: %v", err)
	}
	for _, it := range funds {
		if it.Type != "fund" {
			t.Fatalf("type filter leaked a %s row", it.Type)
		}
	}

	// Text search by our unique code should find exactly our fund.
	hits, err := cs.List(ctx, ListParams{Query: "CZ" + sfx})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	found := false
	for _, it := range hits {
		if it.Code == "CZ"+sfx {
			found = true
			if it.LatestPrice != "12.3456" {
				t.Fatalf("latest price=%s want 12.3456", it.LatestPrice)
			}
		}
	}
	if !found {
		t.Fatal("search did not return the seeded fund")
	}

	// Risk filter = 3 should not include our risk-4 bond.
	risk3, err := cs.List(ctx, ListParams{RiskLevel: 3})
	if err != nil {
		t.Fatalf("risk filter: %v", err)
	}
	for _, it := range risk3 {
		if it.Code == "CZB"+sfx {
			t.Fatal("risk filter returned a risk-4 bond when asking for risk 3")
		}
	}

	// Unknown type → validation error.
	if _, err := cs.List(ctx, ListParams{Type: "crypto"}); err == nil {
		var appErr *domain.AppError
		_ = appErr
		t.Fatal("expected error for unknown asset type")
	}
}

// TestCatalog_PeriodReturnAndSort covers multi-period returns and performance
// sort (catalog.go fundPeriodReturn/periodStart/sortAssets; US-011 Add-on).
func TestCatalog_PeriodReturnAndSort(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	cs := NewCatalogService(store, clock)
	ctx := context.Background()

	sfx := shortID()
	// Two funds with the same latest NAV date but different 1Y base.
	hiID := seedFund(t, store, clock, "PRH"+sfx, decimal.RequireFromString("20"), 3)
	loID := seedFund(t, store, clock, "PRL"+sfx, decimal.RequireFromString("11"), 3)
	t.Cleanup(func() { cleanupAsset(t, pool, "fund", hiID); cleanupAsset(t, pool, "fund", loID) })

	// Add an older NAV ~11 months back (inside the 1Y window) as the base.
	base := clock.Now().AddDate(0, -11, 0)
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{FundID: hiID, NavDate: base, Nav: decimal.RequireFromString("10")}); err != nil {
		t.Fatalf("seed hi base nav: %v", err)
	}
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{FundID: loID, NavDate: base, Nav: decimal.RequireFromString("10")}); err != nil {
		t.Fatalf("seed lo base nav: %v", err)
	}

	items, err := cs.List(ctx, ListParams{Type: "fund", Period: "1Y", Sort: "performance", Order: "desc"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// Filter down to just our two seeded funds while preserving order.
	var hiPos, loPos = -1, -1
	for i, it := range items {
		switch it.Code {
		case "PRH" + sfx:
			hiPos = i
			// (20-10)/10*100 = 100.00
			if it.PeriodReturn != "100.00" {
				t.Fatalf("hi 1Y return=%s want 100.00", it.PeriodReturn)
			}
		case "PRL" + sfx:
			loPos = i
			// (11-10)/10*100 = 10.00
			if it.PeriodReturn != "10.00" {
				t.Fatalf("lo 1Y return=%s want 10.00", it.PeriodReturn)
			}
		}
	}
	if hiPos == -1 || loPos == -1 {
		t.Fatalf("seeded funds missing (hi=%d lo=%d)", hiPos, loPos)
	}
	if hiPos > loPos {
		t.Fatal("performance desc sort should place the higher return first")
	}
}

// TestCatalog_DetailAndStaleness covers fund/bond/gold Detail incl. staleness
// flag and the not-found path (catalog.go Detail; US-012, EX-10).
func TestCatalog_DetailAndStaleness(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	cs := NewCatalogService(store, clock)
	ctx := context.Background()

	sfx := shortID()
	fundID := seedFund(t, store, clock, "DT"+sfx, decimal.RequireFromString("15"), 2)
	bondID := seedBond(t, store, clock, "DTB"+sfx, 3)
	goldID := seedGold(t, store, clock, "DTG"+sfx, decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	t.Cleanup(func() {
		cleanupAsset(t, pool, "fund", fundID)
		cleanupAsset(t, pool, "bond", bondID)
		cleanupAsset(t, pool, "gold", goldID)
	})

	fd, err := cs.Detail(ctx, "fund", fundID)
	if err != nil || fd.Code != "DT"+sfx || fd.Fields["nav"] != "15.0000" {
		t.Fatalf("fund detail=%+v err=%v", fd, err)
	}
	if fd.Unavailable {
		t.Fatal("fresh fund NAV should not be flagged unavailable")
	}

	bd, err := cs.Detail(ctx, "bond", bondID)
	if err != nil || bd.Fields["faceValue"] != "1000.0000" {
		t.Fatalf("bond detail=%+v err=%v", bd, err)
	}

	// Gold detail now fresh → not stale.
	gd, err := cs.Detail(ctx, "gold", goldID)
	if err != nil || gd.Unavailable {
		t.Fatalf("gold detail=%+v err=%v", gd, err)
	}

	// Advance the clock beyond the 5-minute gold staleness threshold.
	clock.Advance(goldStaleAfter + time.Second)
	stale, err := cs.Detail(ctx, "gold", goldID)
	if err != nil {
		t.Fatalf("gold detail (stale): %v", err)
	}
	if !stale.Unavailable {
		t.Fatal("gold price older than 5 min should be flagged unavailable (EX-10)")
	}
	clock.Set(testClock().Now())

	// Not found.
	_, err = cs.Detail(ctx, "fund", 999999999)
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeTxnNotFound {
		t.Fatalf("expected TXN_NOT_FOUND for missing asset, got %v", err)
	}
	// Unknown type.
	if _, err := cs.Detail(ctx, "nope", fundID); err == nil {
		t.Fatal("expected error for unknown asset type")
	}
}

// TestCatalog_PriceHistory covers the history series for each asset type
// (catalog.go PriceHistory; US-012).
func TestCatalog_PriceHistory(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	cs := NewCatalogService(store, clock)
	ctx := context.Background()

	sfx := shortID()
	fundID := seedFund(t, store, clock, "PH"+sfx, decimal.RequireFromString("18"), 3)
	goldID := seedGold(t, store, clock, "PHG"+sfx, decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	t.Cleanup(func() { cleanupAsset(t, pool, "fund", fundID); cleanupAsset(t, pool, "gold", goldID) })

	// Older fund NAV within the 1Y window.
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
		FundID: fundID, NavDate: clock.Now().AddDate(0, -2, 0), Nav: decimal.RequireFromString("16"),
	}); err != nil {
		t.Fatalf("seed older nav: %v", err)
	}

	navs, err := cs.PriceHistory(ctx, "fund", fundID, "1Y")
	if err != nil {
		t.Fatalf("fund history: %v", err)
	}
	if len(navs) < 2 {
		t.Fatalf("expected >=2 NAV points, got %d", len(navs))
	}

	gold, err := cs.PriceHistory(ctx, "gold", goldID, "1Y")
	if err != nil {
		t.Fatalf("gold history: %v", err)
	}
	if len(gold) < 1 || gold[0].Bid == "" || gold[0].Ask == "" {
		t.Fatalf("gold history missing bid/ask: %+v", gold)
	}

	if _, err := cs.PriceHistory(ctx, "bogus", fundID, "1Y"); err == nil {
		t.Fatal("expected error for unknown asset type in history")
	}
}
