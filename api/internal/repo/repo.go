// Package repo provides repository interfaces and sqlc-backed implementations
// over MySQL (components.md § Repositories). Money values are shopspring/decimal
// (mapped from DECIMAL(18,4) by sqlc). All multi-statement mutations run inside
// a single DB transaction via WithinTx (NFR-05).
package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/webond/api/internal/db"
)

// Store is the top-level data access entry point. It owns the connection pool,
// exposes per-domain repositories bound to the pool, and runs units of work in
// a transaction where all repositories share the same *db.Queries (and thus the
// same tx).
type Store struct {
	pool    *sql.DB
	queries *db.Queries

	Users         UserRepo
	Risk          RiskRepo
	Wallets       WalletRepo
	Funds         FundRepo
	Bonds         BondRepo
	Gold          GoldRepo
	Orders        OrderRepo
	Holdings      HoldingRepo
	Dca           DcaRepo
	Transactions  TransactionRepo
	Notifications NotificationRepo
	Audit         AuditRepo
	PriceGen      PriceGeneratorRepo
}

// NewStore builds a Store from an open *sql.DB pool. Repositories are bound to
// the pool for non-transactional reads/writes.
func NewStore(pool *sql.DB) *Store {
	q := db.New(pool)
	return newStoreFromQueries(pool, q)
}

func newStoreFromQueries(pool *sql.DB, q *db.Queries) *Store {
	return &Store{
		pool:          pool,
		queries:       q,
		Users:         &userRepo{q: q},
		Risk:          &riskRepo{q: q},
		Wallets:       &walletRepo{q: q},
		Funds:         &fundRepo{q: q},
		Bonds:         &bondRepo{q: q},
		Gold:          &goldRepo{q: q},
		Orders:        &orderRepo{q: q},
		Holdings:      &holdingRepo{q: q},
		Dca:           &dcaRepo{q: q},
		Transactions:  &transactionRepo{q: q},
		Notifications: &notificationRepo{q: q},
		Audit:         &auditRepo{q: q},
		PriceGen:      &priceGeneratorRepo{q: q},
	}
}

// WithinTx runs fn inside a single transaction. The Store passed to fn has all
// repositories bound to the transaction's *db.Queries, so every write in fn is
// atomic: commit on success, rollback on error or panic. This is the canonical
// way services perform multi-statement mutations (e.g. reserve cash + create
// order + write transaction — components.md § TradingService).
func (s *Store) WithinTx(ctx context.Context, fn func(txStore *Store) error) (err error) {
	return s.withinTx(ctx, nil, fn)
}

// WithinTxReadCommitted is WithinTx at READ COMMITTED. InnoDB then takes no gap
// locks on searches, so short write transactions that range-scan a secondary
// index (one-time codes in notification_outbox) do not deadlock with each other
// across users. Callers serialise per user with an explicit row lock.
func (s *Store) WithinTxReadCommitted(ctx context.Context, fn func(txStore *Store) error) error {
	return s.withinTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted}, fn)
}

func (s *Store) withinTx(ctx context.Context, opts *sql.TxOptions, fn func(txStore *Store) error) (err error) {
	tx, err := s.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-panic after cleanup so the recovery middleware sees it
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				err = fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
			}
			return
		}
		if cmErr := tx.Commit(); cmErr != nil {
			err = fmt.Errorf("commit tx: %w", cmErr)
		}
	}()

	txStore := newStoreFromQueries(s.pool, s.queries.WithTx(tx))
	err = fn(txStore)
	return err
}

// HoldersByAsset returns all current holders of an asset (used by the bond
// lifecycle for coupon/maturity payouts — FR-23).
func (s *Store) HoldersByAsset(ctx context.Context, assetType db.HoldingsAssetType, assetID uint64) ([]db.Holding, error) {
	return s.Holdings.ListByAsset(ctx, db.ListHoldersByAssetParams{
		AssetType: assetType, AssetID: assetID,
	})
}
