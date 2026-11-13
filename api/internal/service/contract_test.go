package service

import (
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// Response details the web app relies on (Task 10.7 web<->API contract).

func TestCatalog_BondPricesAreThbPerUnit(t *testing.T) {
	ctx, _, bondID, _, _, _, clock := bondFixture(t) // quote moved to clean 100.5%
	store := openTestStore(t)
	cs := NewCatalogService(store, clock)

	items, err := cs.List(ctx, ListParams{Type: "bond", Page: 1, PageSize: 500})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found bool
	for _, it := range items {
		if it.ID == bondID {
			found = true
			if it.LatestPrice != "1005.0000" {
				t.Errorf("list latestPrice = %s, want 1005.0000 (THB per unit, not 100.5%%)", it.LatestPrice)
			}
		}
	}
	if !found {
		t.Fatal("bond not in list")
	}

	d, err := cs.Detail(ctx, "bond", bondID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if d.Fields["price"] != "1005.0000" || d.Fields["cleanPrice"] != "100.5000" {
		t.Errorf("detail price=%v cleanPrice=%v, want 1005.0000 and 100.5000", d.Fields["price"], d.Fields["cleanPrice"])
	}

	pts, err := cs.PriceHistory(ctx, "bond", bondID, "1M")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(pts) == 0 || pts[len(pts)-1].Value != "1005.0000" {
		t.Errorf("latest history point = %+v, want value 1005.0000", pts)
	}
}

func TestDashboard_MoversEventsAndMonthlyCommitment(t *testing.T) {
	ctx, userID, bondID, _, ps, ws, clock := bondFixture(t)
	store := openTestStore(t)
	dash := NewDashboardService(store, ps, ws, clock)
	now := clock.Now()

	// A coupon due in 10 days on the held 100 units at 15 THB/unit = 1,500.
	if _, err := store.Bonds.CreateCouponSchedule(ctx, db.CreateBondCouponScheduleParams{
		BondID: bondID, CouponDate: now.AddDate(0, 0, 10), AmountPerUnit: decimal.RequireFromString("15"),
	}); err != nil {
		t.Fatalf("coupon schedule: %v", err)
	}
	// Two active DCA plans: monthly 5,000 and weekly 1,200 (=5,200/month).
	fundID := seedFund(t, store, clock, "MC"+shortID(), decimal.RequireFromString("10"), 1)
	t.Cleanup(func() { cleanupAsset(t, rawPool(t), "fund", fundID) })
	for _, p := range []struct {
		freq   db.DcaPlansFrequency
		amount string
		day    int8
	}{{db.DcaPlansFrequencyMonthly, "5000", 5}, {db.DcaPlansFrequencyWeekly, "1200", 2}} {
		if _, err := store.Dca.CreatePlan(ctx, db.CreateDcaPlanParams{
			UserID: userID, AssetType: db.DcaPlansAssetTypeFund, AssetID: fundID,
			AmountPerRound: decimal.RequireFromString(p.amount), Frequency: p.freq, RunDay: p.day,
			StartDate: now, NextRunDate: sql.NullTime{Time: now.AddDate(0, 0, 3), Valid: true},
		}); err != nil {
			t.Fatalf("create plan: %v", err)
		}
	}

	d, err := dash.Get(ctx, userID, "1M")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(d.TopMovers) == 0 || d.TopMovers[0].AssetType != "bond" || d.TopMovers[0].AssetID != bondID || d.TopMovers[0].Name == "" {
		t.Errorf("top movers must identify the asset, got %+v", d.TopMovers)
	}
	if d.DCASummary.ActivePlans != 2 || d.DCASummary.MonthlyCommitment != "10200.00" {
		t.Errorf("dca summary = %+v, want 2 active and monthlyCommitment 10200.00", d.DCASummary)
	}
	var coupon, dca bool
	for _, e := range d.UpcomingEvents {
		switch e.Type {
		case "coupon":
			coupon = e.Ref != "" && e.Amount == "1500.00"
		case "dca_run":
			if e.Ref != "" && (e.Amount == "5000.00" || e.Amount == "1200.00") {
				dca = true
			}
		}
	}
	if !coupon || !dca {
		t.Errorf("events must carry ref + amount (coupon=%v dca=%v): %+v", coupon, dca, d.UpcomingEvents)
	}

	pf, err := ps.Get(ctx, userID)
	if err != nil {
		t.Fatalf("portfolio: %v", err)
	}
	for _, h := range pf.Holdings {
		if h.AssetID == bondID && h.TotalCost != "100000.00" {
			t.Errorf("holding totalCost = %q, want 100000.00", h.TotalCost)
		}
	}
}
