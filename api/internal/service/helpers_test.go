package service

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/webond/api/internal/testdb"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/database"
	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/repo"
)

// openTestStore opens a Store against the test DB, skipping if unreachable.
func openTestStore(t *testing.T) *repo.Store {
	t.Helper()
	dsn := testdb.DSN(t)
	pool, err := database.Open(context.Background(), dsn)
	if err != nil {
		t.Skipf("cannot reach test DB (%v); skipping", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return repo.NewStore(pool)
}

// rawPool opens a second connection for direct cleanup/fixture SQL.
func rawPool(t *testing.T) *sql.DB {
	t.Helper()
	pool, err := database.Open(context.Background(), testdb.DSN(t))
	if err != nil {
		t.Fatalf("open raw pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}

// testClock returns a FixedClock pinned to a deterministic business day/time in
// Asia/Bangkok (a Wednesday at 10:00, before the 15:30 MF cut-off).
func testClock() *FixedClock {
	return NewFixedClock(time.Date(2026, 10, 28, 10, 0, 0, 0, Bangkok))
}

// testTokens builds a JWT manager with dev secrets for tests.
func testTokens() *auth.Manager {
	return auth.NewManager("test-access-secret", "test-refresh-secret")
}

// emailSeq makes uniqueEmail unique even when called from parallel goroutines.
var emailSeq atomic.Int64

// uniqueEmail avoids UNIQUE(email) clashes across runs and goroutines.
func uniqueEmail() string {
	return "svc-" + decimal.NewFromInt(time.Now().UnixNano()).String() + "-" +
		decimal.NewFromInt(emailSeq.Add(1)).String() + "@webond.test"
}

// cleanupUser removes a test user and all its dependent rows.
func cleanupUser(t *testing.T, pool *sql.DB, userID uint64) {
	t.Helper()
	ctx := context.Background()
	// Delete dependents first (FKs may be RESTRICT for some tables).
	stmts := []string{
		"DELETE FROM transaction_status_history WHERE transaction_id IN (SELECT id FROM transactions WHERE user_id = ?)",
		"DELETE FROM transactions WHERE user_id = ?",
		"DELETE FROM dca_rounds WHERE dca_plan_id IN (SELECT id FROM dca_plans WHERE user_id = ?)",
		"DELETE FROM dca_plans WHERE user_id = ?",
		"DELETE FROM orders WHERE user_id = ?",
		"DELETE FROM holdings WHERE user_id = ?",
		"DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id = ?)",
		"DELETE FROM wallets WHERE user_id = ?",
		"DELETE FROM risk_assessments WHERE user_id = ?",
		"DELETE FROM notification_outbox WHERE user_id = ?",
		"DELETE FROM audit_logs WHERE user_id = ?",
		"DELETE FROM users WHERE id = ?",
	}
	for _, q := range stmts {
		if _, err := pool.ExecContext(ctx, q, userID); err != nil {
			t.Logf("cleanup (%s): %v", q, err)
		}
	}
}

// seedActiveUser creates an active investor with a funded wallet and a valid
// risk assessment at the given level, returning the user id. Used by trading/
// DCA/wallet tests that need a tradeable account.
func seedActiveUser(t *testing.T, store *repo.Store, clock Clock, balance decimal.Decimal, riskLevel int8) uint64 {
	t.Helper()
	ctx := context.Background()
	var userID uint64
	err := store.WithinTx(ctx, func(tx *repo.Store) error {
		res, err := tx.Users.Create(ctx, db.CreateUserParams{
			Email:        uniqueEmail(),
			PasswordHash: "x",
			FullName:     "Seeded Investor",
			Role:         db.UsersRoleInvestor,
			Status:       db.UsersStatusActive,
		})
		id, err := lastInsertID(res, err)
		if err != nil {
			return err
		}
		userID = id
		if _, err := tx.Wallets.Create(ctx, userID); err != nil {
			return err
		}
		w, err := tx.Wallets.GetByUser(ctx, userID)
		if err != nil {
			return err
		}
		if err := tx.Wallets.UpdateBalances(ctx, db.UpdateWalletBalancesParams{
			Balance:  balance,
			Reserved: decimal.Zero,
			ID:       w.ID,
		}); err != nil {
			return err
		}
		if riskLevel > 0 {
			if _, err := tx.Risk.Create(ctx, db.CreateRiskAssessmentParams{
				UserID:      userID,
				RiskLevel:   riskLevel,
				AnswersJson: []byte(`{}`),
				AssessedAt:  clock.Now(),
				ExpiresAt:   clock.Now().Add(riskExpiryTTL),
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seedActiveUser: %v", err)
	}
	return userID
}

// --- asset fixture helpers (trading/catalog/portfolio tests) ---

// seedFund inserts a fund with a latest NAV and returns its id. riskLevel 0 → NULL.
func seedFund(t *testing.T, store *repo.Store, clock Clock, code string, nav decimal.Decimal, riskLevel int16) uint64 {
	t.Helper()
	ctx := context.Background()
	res, err := store.Funds.Create(ctx, db.CreateFundParams{
		Code: code, Name: code + " Fund",
		Category:    sql.NullString{String: "Equity", Valid: true},
		RiskLevel:   sql.NullInt16{Int16: riskLevel, Valid: riskLevel > 0},
		MinPurchase: decimal.RequireFromString("500"), SettlementDays: 3,
		CutoffTime: "15:30:00", IsActive: true,
	})
	id, err := lastInsertID(res, err)
	if err != nil {
		t.Fatalf("seedFund: %v", err)
	}
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
		FundID: id, NavDate: clock.Now(), Nav: nav,
	}); err != nil {
		t.Fatalf("seedFund nav: %v", err)
	}
	return id
}

// seedBond inserts an open-offering bond with a price, returns its id.
func seedBond(t *testing.T, store *repo.Store, clock Clock, code string, riskLevel int16) uint64 {
	t.Helper()
	ctx := context.Background()
	now := clock.Now()
	res, err := store.Bonds.Create(ctx, db.CreateBondParams{
		Code: code, Name: code + " Bond",
		Issuer:          sql.NullString{String: "Gov", Valid: true},
		CouponRate:      decimal.RequireFromString("0.0300"),
		CouponFrequency: 2, FaceValue: decimal.RequireFromString("1000"),
		MaturityDate:  sql.NullTime{Time: now.AddDate(3, 0, 0), Valid: true},
		Ytm:           decimal.NullDecimal{Decimal: decimal.RequireFromString("3.2000"), Valid: true},
		OfferingStart: sql.NullTime{Time: now.AddDate(0, 0, -5), Valid: true},
		OfferingEnd:   sql.NullTime{Time: now.AddDate(0, 0, 30), Valid: true},
		MinLot:        decimal.RequireFromString("100000"),
		RiskLevel:     sql.NullInt16{Int16: riskLevel, Valid: riskLevel > 0}, IsActive: true,
	})
	id, err := lastInsertID(res, err)
	if err != nil {
		t.Fatalf("seedBond: %v", err)
	}
	if err := store.Bonds.UpsertPrice(ctx, db.UpsertBondPriceParams{
		// clean_price is quoted as % of par (100 = at par); accrued_interest is THB per unit.
		BondID: id, PriceDate: now, CleanPrice: decimal.RequireFromString("100"),
		AccruedInterest: decimal.RequireFromString("5"),
	}); err != nil {
		t.Fatalf("seedBond price: %v", err)
	}
	return id
}

// seedGold inserts a gold product with a fresh BID/ASK price, returns its id.
func seedGold(t *testing.T, store *repo.Store, clock Clock, code string, bid, ask decimal.Decimal) uint64 {
	t.Helper()
	ctx := context.Background()
	res, err := store.Gold.Create(ctx, db.CreateGoldProductParams{
		Code: code, Name: code + " Gold",
		Purity: sql.NullString{String: "96.5%", Valid: true}, IsActive: true,
	})
	id, err := lastInsertID(res, err)
	if err != nil {
		t.Fatalf("seedGold: %v", err)
	}
	if _, err := store.Gold.CreatePrice(ctx, db.CreateGoldPriceParams{
		GoldProductID: id, PricedAt: clock.Now(), BidPerGram: bid, AskPerGram: ask,
	}); err != nil {
		t.Fatalf("seedGold price: %v", err)
	}
	return id
}

// cleanupFund / cleanupBond / cleanupGold remove asset fixtures + price rows.
func cleanupAsset(t *testing.T, pool *sql.DB, kind string, id uint64) {
	t.Helper()
	ctx := context.Background()
	switch kind {
	case "fund":
		_, _ = pool.ExecContext(ctx, "DELETE FROM fund_navs WHERE fund_id = ?", id)
		_, _ = pool.ExecContext(ctx, "DELETE FROM funds WHERE id = ?", id)
	case "bond":
		_, _ = pool.ExecContext(ctx, "DELETE FROM bond_prices WHERE bond_id = ?", id)
		_, _ = pool.ExecContext(ctx, "DELETE FROM bond_coupon_schedules WHERE bond_id = ?", id)
		_, _ = pool.ExecContext(ctx, "DELETE FROM bonds WHERE id = ?", id)
	case "gold":
		_, _ = pool.ExecContext(ctx, "DELETE FROM gold_prices WHERE gold_product_id = ?", id)
		_, _ = pool.ExecContext(ctx, "DELETE FROM gold_products WHERE id = ?", id)
	}
}
