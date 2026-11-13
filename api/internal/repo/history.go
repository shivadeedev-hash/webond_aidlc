package repo

import (
	"context"
	"database/sql"

	"github.com/webond/api/internal/db"
)

// TransactionRepo is the immutable history feed persistence (FR-16..20, BR-14).
// Rows are never edited/deleted; corrections are new reversal rows.
type TransactionRepo interface {
	GetByID(ctx context.Context, id uint64) (db.Transaction, error)
	GetByTxnRef(ctx context.Context, txnRef string) (db.Transaction, error)
	Create(ctx context.Context, arg db.CreateTransactionParams) (sql.Result, error)
	UpdateStatus(ctx context.Context, arg db.UpdateTransactionStatusParams) error
	ListByUser(ctx context.Context, arg db.ListTransactionsByUserParams) ([]db.Transaction, error)
	CountByUser(ctx context.Context, userID uint64) (int64, error)
	CreateStatusHistory(ctx context.Context, arg db.CreateTransactionStatusHistoryParams) (sql.Result, error)
	ListStatusHistory(ctx context.Context, transactionID uint64) ([]db.TransactionStatusHistory, error)
}

type transactionRepo struct{ q *db.Queries }

func (r *transactionRepo) GetByID(ctx context.Context, id uint64) (db.Transaction, error) {
	return r.q.GetTransactionByID(ctx, id)
}
func (r *transactionRepo) GetByTxnRef(ctx context.Context, txnRef string) (db.Transaction, error) {
	return r.q.GetTransactionByTxnRef(ctx, txnRef)
}
func (r *transactionRepo) Create(ctx context.Context, arg db.CreateTransactionParams) (sql.Result, error) {
	return r.q.CreateTransaction(ctx, arg)
}
func (r *transactionRepo) UpdateStatus(ctx context.Context, arg db.UpdateTransactionStatusParams) error {
	return r.q.UpdateTransactionStatus(ctx, arg)
}
func (r *transactionRepo) ListByUser(ctx context.Context, arg db.ListTransactionsByUserParams) ([]db.Transaction, error) {
	return r.q.ListTransactionsByUser(ctx, arg)
}
func (r *transactionRepo) CountByUser(ctx context.Context, userID uint64) (int64, error) {
	return r.q.CountTransactionsByUser(ctx, userID)
}
func (r *transactionRepo) CreateStatusHistory(ctx context.Context, arg db.CreateTransactionStatusHistoryParams) (sql.Result, error) {
	return r.q.CreateTransactionStatusHistory(ctx, arg)
}
func (r *transactionRepo) ListStatusHistory(ctx context.Context, transactionID uint64) ([]db.TransactionStatusHistory, error) {
	return r.q.ListTransactionStatusHistory(ctx, transactionID)
}
