package repo

import (
	"context"
	"database/sql"

	"github.com/webond/api/internal/db"
)

// WalletRepo is cash balance + ledger persistence (FR-07, BR-11). Reserve /
// release / settle mutations run inside Store.WithinTx with GetByUserForUpdate.
type WalletRepo interface {
	GetByUser(ctx context.Context, userID uint64) (db.Wallet, error)
	GetByUserForUpdate(ctx context.Context, userID uint64) (db.Wallet, error)
	Create(ctx context.Context, userID uint64) (sql.Result, error)
	UpdateBalances(ctx context.Context, arg db.UpdateWalletBalancesParams) error
	CreateTransaction(ctx context.Context, arg db.CreateWalletTransactionParams) (sql.Result, error)
	ListTransactions(ctx context.Context, arg db.ListWalletTransactionsParams) ([]db.WalletTransaction, error)
	CountTransactions(ctx context.Context, walletID uint64) (int64, error)
}

type walletRepo struct{ q *db.Queries }

func (r *walletRepo) GetByUser(ctx context.Context, userID uint64) (db.Wallet, error) {
	return r.q.GetWalletByUser(ctx, userID)
}
func (r *walletRepo) GetByUserForUpdate(ctx context.Context, userID uint64) (db.Wallet, error) {
	return r.q.GetWalletByUserForUpdate(ctx, userID)
}
func (r *walletRepo) Create(ctx context.Context, userID uint64) (sql.Result, error) {
	return r.q.CreateWallet(ctx, userID)
}
func (r *walletRepo) UpdateBalances(ctx context.Context, arg db.UpdateWalletBalancesParams) error {
	return r.q.UpdateWalletBalances(ctx, arg)
}
func (r *walletRepo) CreateTransaction(ctx context.Context, arg db.CreateWalletTransactionParams) (sql.Result, error) {
	return r.q.CreateWalletTransaction(ctx, arg)
}
func (r *walletRepo) ListTransactions(ctx context.Context, arg db.ListWalletTransactionsParams) ([]db.WalletTransaction, error) {
	return r.q.ListWalletTransactions(ctx, arg)
}
func (r *walletRepo) CountTransactions(ctx context.Context, walletID uint64) (int64, error) {
	return r.q.CountWalletTransactions(ctx, walletID)
}
