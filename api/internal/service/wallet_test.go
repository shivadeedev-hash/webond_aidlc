package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

func TestWallet_DepositWithdraw(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// Deposit increases balance (US-008).
	b, err := ws.Deposit(ctx, userID, decimal.RequireFromString("10000"))
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if !b.Balance.Equal(decimal.RequireFromString("10000.0000")) {
		t.Fatalf("balance after deposit: %s", b.Balance)
	}

	// Withdraw within available succeeds (US-009).
	b, err = ws.Withdraw(ctx, userID, decimal.RequireFromString("4000"))
	if err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if !b.Available.Equal(decimal.RequireFromString("6000.0000")) {
		t.Fatalf("available after withdraw: %s", b.Available)
	}

	// Over-withdraw is rejected (INSUFFICIENT_BALANCE).
	_, err = ws.Withdraw(ctx, userID, decimal.RequireFromString("999999"))
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeInsufficientBalance {
		t.Fatalf("expected INSUFFICIENT_BALANCE, got %v", err)
	}
}

func TestWallet_ReserveReleaseSettle(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("10000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// Reserve moves available→reserved, balance unchanged.
	err := store.WithinTx(ctx, func(tx *repo.Store) error {
		return ws.ReserveTx(ctx, tx, userID, decimal.RequireFromString("3000"), 0)
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	b, _ := ws.Get(ctx, userID)
	if !b.Reserved.Equal(decimal.RequireFromString("3000.0000")) || !b.Available.Equal(decimal.RequireFromString("7000.0000")) {
		t.Fatalf("after reserve: reserved=%s available=%s", b.Reserved, b.Available)
	}

	// Release returns reserved→available.
	err = store.WithinTx(ctx, func(tx *repo.Store) error {
		return ws.ReleaseTx(ctx, tx, userID, decimal.RequireFromString("1000"), 0)
	})
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	b, _ = ws.Get(ctx, userID)
	if !b.Reserved.Equal(decimal.RequireFromString("2000.0000")) {
		t.Fatalf("after release reserved=%s", b.Reserved)
	}

	// Settle debits the remaining reserved from both balance and reserved.
	err = store.WithinTx(ctx, func(tx *repo.Store) error {
		return ws.SettleBuyTx(ctx, tx, userID, decimal.RequireFromString("2000"), 0)
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	b, _ = ws.Get(ctx, userID)
	if !b.Balance.Equal(decimal.RequireFromString("8000.0000")) || !b.Reserved.Equal(decimal.Zero.Round(4)) {
		t.Fatalf("after settle balance=%s reserved=%s", b.Balance, b.Reserved)
	}
	if !b.Available.Equal(decimal.RequireFromString("8000.0000")) {
		t.Fatalf("after settle available=%s", b.Available)
	}
}
