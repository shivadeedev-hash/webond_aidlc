package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/webond/api/internal/testdb"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/database"
	"github.com/webond/api/internal/db"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := testdb.DSN(t)
	pool, err := database.Open(context.Background(), dsn)
	if err != nil {
		t.Skipf("cannot reach test DB (%v); skipping", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return NewStore(pool)
}

// uniqueEmail avoids UNIQUE(email) clashes across runs.
func uniqueEmail() string {
	return fmt.Sprintf("it-%d@webond.test", time.Now().UnixNano())
}

func TestIntegration_CreateUserAndWallet_InTx(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	email := uniqueEmail()

	var userID uint64
	err := store.WithinTx(ctx, func(tx *Store) error {
		res, err := tx.Users.Create(ctx, db.CreateUserParams{
			Email:        email,
			PasswordHash: "x",
			FullName:     "Integration Test",
			Role:         db.UsersRoleInvestor,
			Status:       db.UsersStatusActive,
		})
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		userID = uint64(id)
		if _, err := tx.Wallets.Create(ctx, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
	t.Cleanup(func() { cleanupUser(context.Background(), store, userID) })

	// Committed data is visible on the pool.
	u, err := store.Users.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if u.ID != userID {
		t.Fatalf("user id mismatch: got %d want %d", u.ID, userID)
	}
	w, err := store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetWalletByUser: %v", err)
	}
	if !w.Balance.Equal(decimal.Zero) || !w.Reserved.Equal(decimal.Zero) {
		t.Fatalf("new wallet not zeroed: balance=%s reserved=%s", w.Balance, w.Reserved)
	}
}

func TestIntegration_TxRollback(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	email := uniqueEmail()

	sentinel := errors.New("force rollback")
	err := store.WithinTx(ctx, func(tx *Store) error {
		if _, err := tx.Users.Create(ctx, db.CreateUserParams{
			Email:        email,
			PasswordHash: "x",
			FullName:     "Rollback Test",
			Role:         db.UsersRoleInvestor,
			Status:       db.UsersStatusActive,
		}); err != nil {
			return err
		}
		return sentinel // triggers rollback
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	// The user must NOT exist after rollback.
	if _, err := store.Users.GetByEmail(ctx, email); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected no rows after rollback, got %v", err)
	}
}

func TestIntegration_DecimalRoundTrip(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	email := uniqueEmail()

	res, err := store.Users.Create(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: "x",
		FullName:     "Decimal Test",
		Role:         db.UsersRoleInvestor,
		Status:       db.UsersStatusActive,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	id, _ := res.LastInsertId()
	userID := uint64(id)
	t.Cleanup(func() { cleanupUser(context.Background(), store, userID) })

	if _, err := store.Wallets.Create(ctx, userID); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	w, err := store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}

	// Store a precise DECIMAL(18,4) value and read it back exactly.
	want := decimal.RequireFromString("1234567.8901")
	if err := store.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
		Balance:  want,
		Reserved: decimal.RequireFromString("0.5000"),
		ID:       w.ID,
	}); err != nil {
		t.Fatalf("update balances: %v", err)
	}
	got, err := store.Wallets.GetByUser(ctx, userID)
	if err != nil {
		t.Fatalf("re-get wallet: %v", err)
	}
	if !got.Balance.Equal(want) {
		t.Fatalf("decimal round-trip: got %s want %s", got.Balance, want)
	}
	if !got.Reserved.Equal(decimal.RequireFromString("0.5000")) {
		t.Fatalf("reserved round-trip: got %s", got.Reserved)
	}
}

// cleanupUser removes a test user (wallets cascade via FK ON DELETE CASCADE).
func cleanupUser(ctx context.Context, store *Store, userID uint64) {
	// Best-effort: use a short-lived raw exec through the pool.
	_ = store.WithinTx(ctx, func(tx *Store) error {
		return tx.deleteUser(ctx, userID)
	})
}

// deleteUser is a test-only helper to clean up fixtures.
func (s *Store) deleteUser(ctx context.Context, userID uint64) error {
	_, err := s.pool.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	return err
}
