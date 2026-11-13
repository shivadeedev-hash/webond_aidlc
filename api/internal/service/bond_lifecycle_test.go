package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/repo"
)

// seedBondHolding gives a user a bond position directly (bypassing trading) so
// lifecycle payouts have a holder to credit.
func seedBondHolding(t *testing.T, store *repo.Store, userID, bondID uint64, qty decimal.Decimal, cost decimal.Decimal) {
	t.Helper()
	ctx := context.Background()
	if err := store.Holdings.Upsert(ctx, db.UpsertHoldingParams{
		UserID: userID, AssetType: db.HoldingsAssetTypeBond, AssetID: bondID,
		Quantity: qty, LockedQuantity: decimal.Zero, AvgCost: cost.Div(qty).Round(4), TotalCost: cost,
	}); err != nil {
		t.Fatalf("seedBondHolding: %v", err)
	}
}

// seedMaturedBond creates a bond whose maturity date is in the past.
func seedMaturedBond(t *testing.T, store *repo.Store, clock Clock, code string) uint64 {
	t.Helper()
	ctx := context.Background()
	now := clock.Now()
	res, err := store.Bonds.Create(ctx, db.CreateBondParams{
		Code: code, Name: code + " Matured",
		Issuer:          sql.NullString{String: "Gov", Valid: true},
		CouponRate:      decimal.RequireFromString("0.0300"),
		CouponFrequency: 2, FaceValue: decimal.RequireFromString("1000"),
		MaturityDate: sql.NullTime{Time: now.AddDate(0, 0, -1), Valid: true},
		MinLot:       decimal.RequireFromString("100000"),
		IsActive:     true,
	})
	id, err := lastInsertID(res, err)
	if err != nil {
		t.Fatalf("seedMaturedBond: %v", err)
	}
	return id
}

// TestBondLifecycle_RunCoupons credits a due coupon to a bond holder and marks
// the schedule paid (bond_lifecycle.go RunCoupons/payCoupon; FR-23).
func TestBondLifecycle_RunCoupons(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	bl := NewBondLifecycleService(store, ws, notify, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("0"), 5)
	sfx := shortID()
	bondID := seedBond(t, store, clock, "CP"+sfx, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "bond", bondID) })

	// Hold 100 units; a coupon of 5 THB/unit due yesterday → 500 THB credit.
	seedBondHolding(t, store, userID, bondID, decimal.RequireFromString("100"), decimal.RequireFromString("100000"))
	if _, err := store.Bonds.CreateCouponSchedule(ctx, db.CreateBondCouponScheduleParams{
		BondID: bondID, CouponDate: clock.Now().AddDate(0, 0, -1),
		AmountPerUnit: decimal.RequireFromString("5"), Paid: false,
	}); err != nil {
		t.Fatalf("seed coupon schedule: %v", err)
	}

	n, err := bl.RunCoupons(ctx)
	if err != nil {
		t.Fatalf("run coupons: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected >=1 coupon processed, got %d", n)
	}
	bal, _ := ws.Get(ctx, userID)
	if !bal.Balance.Equal(decimal.RequireFromString("500.0000")) {
		t.Fatalf("expected 500 coupon credited, got %s", bal.Balance)
	}

	// Running again must not re-pay (schedule now marked paid).
	bal0 := bal.Balance
	if _, err := bl.RunCoupons(ctx); err != nil {
		t.Fatalf("run coupons #2: %v", err)
	}
	bal2, _ := ws.Get(ctx, userID)
	if !bal2.Balance.Equal(bal0) {
		t.Fatalf("coupon double-paid: %s -> %s", bal0, bal2.Balance)
	}
}

// TestBondLifecycle_RunMaturities credits face value, zeroes the holding and
// records the maturity (bond_lifecycle.go RunMaturities/matureBond; FR-23, BR-05).
func TestBondLifecycle_RunMaturities(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	bl := NewBondLifecycleService(store, ws, notify, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("0"), 5)
	sfx := shortID()
	bondID := seedMaturedBond(t, store, clock, "MT"+sfx)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "bond", bondID) })

	// Hold 200 units of face value 1000 → 200,000 THB credited on maturity.
	seedBondHolding(t, store, userID, bondID, decimal.RequireFromString("200"), decimal.RequireFromString("200000"))

	n, err := bl.RunMaturities(ctx)
	if err != nil {
		t.Fatalf("run maturities: %v", err)
	}
	if n < 1 {
		t.Fatalf("expected >=1 closeout, got %d", n)
	}
	bal, _ := ws.Get(ctx, userID)
	if bal.Balance.LessThan(decimal.RequireFromString("200000")) {
		t.Fatalf("expected >=200000 face credited, got %s", bal.Balance)
	}
	// Holding zeroed.
	holdings, _ := store.Holdings.ListByUser(ctx, userID)
	for _, h := range holdings {
		if h.AssetID == bondID && h.AssetType == db.HoldingsAssetTypeBond {
			if !h.Quantity.Equal(decimal.Zero.Round(4)) {
				t.Fatalf("matured holding not zeroed: qty=%s", h.Quantity)
			}
		}
	}
}
