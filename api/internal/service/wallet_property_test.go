package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/repo"
)

// op is one generated wallet operation for the P2 property.
type op struct {
	kind   int // 0 deposit, 1 withdraw, 2 reserve, 3 release
	amount int // THB (integer to keep generators simple; stored as DECIMAL)
}

// TestProperty_CashConservation validates correctness.md P2: after any sequence
// of deposit/withdraw/reserve/release, the invariants hold always
// (reserved≥0, balance≥0, reserved≤balance, available≥0), and total cash only
// moves by net deposits−withdrawals (reserve/release never create or destroy
// cash). Operations that would violate an invariant must be rejected, leaving
// state unchanged.
func TestProperty_CashConservation(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ctx := context.Background()

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 60 // DB-backed; keep bounded but meaningful
	properties := gopter.NewProperties(params)

	opGen := gopter.CombineGens(
		gen.IntRange(0, 3),
		gen.IntRange(1, 5000),
	).Map(func(vals []interface{}) op {
		return op{kind: vals[0].(int), amount: vals[1].(int)}
	})

	properties.Property("cash conserved & invariants hold", prop.ForAll(
		func(ops []op) bool {
			userID := seedActiveUserQuiet(store)
			defer rawCleanup(pool, userID)

			// Track expected total cash (balance) from accepted deposits/withdrawals.
			expectedBalance := decimal.Zero
			expectedReserved := decimal.Zero

			for _, o := range ops {
				amt := decimal.NewFromInt(int64(o.amount))
				switch o.kind {
				case 0: // deposit (always accepted for positive amount)
					if _, err := ws.Deposit(ctx, userID, amt); err != nil {
						return false
					}
					expectedBalance = expectedBalance.Add(amt)
				case 1: // withdraw (accepted iff ≤ available)
					available := expectedBalance.Sub(expectedReserved)
					_, err := ws.Withdraw(ctx, userID, amt)
					if amt.LessThanOrEqual(available) {
						if err != nil {
							return false
						}
						expectedBalance = expectedBalance.Sub(amt)
					} else if err == nil {
						return false // should have been rejected
					}
				case 2: // reserve (accepted iff ≤ available)
					available := expectedBalance.Sub(expectedReserved)
					err := store.WithinTx(ctx, func(tx *repo.Store) error {
						return ws.ReserveTx(ctx, tx, userID, amt, 0)
					})
					if amt.LessThanOrEqual(available) {
						if err != nil {
							return false
						}
						expectedReserved = expectedReserved.Add(amt)
					} else if err == nil {
						return false
					}
				case 3: // release (accepted iff ≤ reserved)
					err := store.WithinTx(ctx, func(tx *repo.Store) error {
						return ws.ReleaseTx(ctx, tx, userID, amt, 0)
					})
					if amt.LessThanOrEqual(expectedReserved) {
						if err != nil {
							return false
						}
						expectedReserved = expectedReserved.Sub(amt)
					} else if err == nil {
						return false
					}
				}

				// Invariants after every op.
				b, err := ws.Get(ctx, userID)
				if err != nil {
					return false
				}
				if b.Reserved.IsNegative() || b.Balance.IsNegative() || b.Available.IsNegative() {
					return false
				}
				if b.Reserved.GreaterThan(b.Balance) {
					return false
				}
				if !b.Balance.Equal(expectedBalance.Round(4)) || !b.Reserved.Equal(expectedReserved.Round(4)) {
					return false
				}
			}
			return true
		},
		gen.SliceOf(opGen),
	))

	properties.TestingRun(t)
}

// seedActiveUserQuiet mirrors seedActiveUser without *testing.T for use inside
// a gopter property (which runs many iterations).
func seedActiveUserQuiet(store *repo.Store) uint64 {
	ctx := context.Background()
	var userID uint64
	_ = store.WithinTx(ctx, func(tx *repo.Store) error {
		res, err := tx.Users.Create(ctx, db.CreateUserParams{
			Email:        uniqueEmail(),
			PasswordHash: "x",
			FullName:     "Property User",
			Role:         db.UsersRoleInvestor,
			Status:       db.UsersStatusActive,
		})
		id, err := lastInsertID(res, err)
		if err != nil {
			return err
		}
		userID = id
		_, err = tx.Wallets.Create(ctx, userID)
		return err
	})
	return userID
}

// rawCleanup removes a user and its wallet/ledger rows via the raw pool.
func rawCleanup(pool *sql.DB, userID uint64) {
	ctx := context.Background()
	_, _ = pool.ExecContext(ctx, "DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id = ?)", userID)
	_, _ = pool.ExecContext(ctx, "DELETE FROM wallets WHERE user_id = ?", userID)
	_, _ = pool.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
}
