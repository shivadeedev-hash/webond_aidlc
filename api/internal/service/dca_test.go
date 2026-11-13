package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
)

func setupDCA(t *testing.T) (*DCAService, *TradingService, *FixedClock) {
	t.Helper()
	store := openTestStore(t)
	clock := testClock()
	notify := NewNotificationService(store, clock)
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	ds := NewDCAService(store, ts, notify, clock)
	return ds, ts, clock
}

func TestDCA_CreateRejectsBondAndLowAmount(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	ds, _, clock := setupDCA(t)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "DC"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	// Bond rejected (BR-06).
	_, err := ds.CreatePlan(ctx, CreatePlanInput{
		UserID: userID, AssetType: "bond", AssetID: 1, AmountPerRound: decimal.RequireFromString("1000"),
		Frequency: "monthly", RunDay: 1,
	})
	if err == nil {
		t.Fatal("expected bond DCA to be rejected")
	}

	// Below MF min 500 (BR-07).
	_, err = ds.CreatePlan(ctx, CreatePlanInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, AmountPerRound: decimal.RequireFromString("100"),
		Frequency: "monthly", RunDay: 1,
	})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR for low amount, got %v", err)
	}
}

func TestDCA_RunDueIdempotent(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	ds, _, clock := setupDCA(t)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "DR"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	plan, err := ds.CreatePlan(ctx, CreatePlanInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, AmountPerRound: decimal.RequireFromString("5000"),
		Frequency: "monthly", RunDay: clock.Now().Day(), StartDate: clock.Now().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}

	if _, err := ds.RunDue(ctx); err != nil {
		t.Fatalf("run due #1: %v", err)
	}
	// Second run on the same date must not create a duplicate round (FR-15).
	if _, err := ds.RunDue(ctx); err != nil {
		t.Fatalf("run due #2: %v", err)
	}
	rounds, _ := ds.ListRounds(ctx, userID, plan.ID)
	successCount := 0
	for _, r := range rounds {
		if sameDate(r.RunDate, ds.today()) && r.Status == db.DcaRoundsStatusSuccess {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 success round for today, got %d", successCount)
	}
}

func TestDCA_PauseAfterThreeFailures(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	ds, _, clock := setupDCA(t)
	ctx := context.Background()

	// Zero balance → every DCA buy fails (insufficient balance).
	userID := seedActiveUser(t, store, clock, decimal.Zero, 5)
	fundID := seedFund(t, store, clock, "DF"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	plan, err := ds.CreatePlan(ctx, CreatePlanInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, AmountPerRound: decimal.RequireFromString("5000"),
		Frequency: "monthly", RunDay: clock.Now().Day(), StartDate: clock.Now().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}

	// Force three failed rounds on three distinct dates by advancing the clock.
	for i := 0; i < maxConsecutiveFailures; i++ {
		if _, err := ds.RunDue(ctx); err != nil {
			t.Fatalf("run due %d: %v", i, err)
		}
		clock.Advance(32 * 24 * 60 * 60 * 1e9) // ~1 month
		// Reactivate next_run to be due again (the plan may have advanced it).
	}

	p, _ := store.Dca.GetPlanByID(ctx, plan.ID)
	if p.Status != db.DcaPlansStatusPaused {
		t.Fatalf("expected plan paused after %d failures, got status=%s failures=%d",
			maxConsecutiveFailures, p.Status, p.ConsecutiveFailures)
	}
}
