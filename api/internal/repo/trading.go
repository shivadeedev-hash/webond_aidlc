package repo

import (
	"context"
	"database/sql"

	"github.com/webond/api/internal/db"
)

// OrderRepo is order persistence (FR-08..12). Order creation + execution run
// inside Store.WithinTx together with wallet/holding/transaction writes.
type OrderRepo interface {
	GetByID(ctx context.Context, id uint64) (db.Order, error)
	GetByTxnRef(ctx context.Context, txnRef string) (db.Order, error)
	GetByIdempotencyKey(ctx context.Context, key sql.NullString) (db.Order, error)
	Create(ctx context.Context, arg db.CreateOrderParams) (sql.Result, error)
	UpdateStatus(ctx context.Context, arg db.UpdateOrderStatusParams) error
	SetExecution(ctx context.Context, arg db.SetOrderExecutionParams) error
	ListByUser(ctx context.Context, arg db.ListOrdersByUserParams) ([]db.Order, error)
}

type orderRepo struct{ q *db.Queries }

func (r *orderRepo) GetByID(ctx context.Context, id uint64) (db.Order, error) {
	return r.q.GetOrderByID(ctx, id)
}
func (r *orderRepo) GetByTxnRef(ctx context.Context, txnRef string) (db.Order, error) {
	return r.q.GetOrderByTxnRef(ctx, txnRef)
}
func (r *orderRepo) GetByIdempotencyKey(ctx context.Context, key sql.NullString) (db.Order, error) {
	return r.q.GetOrderByIdempotencyKey(ctx, key)
}
func (r *orderRepo) Create(ctx context.Context, arg db.CreateOrderParams) (sql.Result, error) {
	return r.q.CreateOrder(ctx, arg)
}
func (r *orderRepo) UpdateStatus(ctx context.Context, arg db.UpdateOrderStatusParams) error {
	return r.q.UpdateOrderStatus(ctx, arg)
}
func (r *orderRepo) SetExecution(ctx context.Context, arg db.SetOrderExecutionParams) error {
	return r.q.SetOrderExecution(ctx, arg)
}
func (r *orderRepo) ListByUser(ctx context.Context, arg db.ListOrdersByUserParams) ([]db.Order, error) {
	return r.q.ListOrdersByUser(ctx, arg)
}

// HoldingRepo is position persistence (FR-21, BR-10 avg cost).
type HoldingRepo interface {
	Get(ctx context.Context, arg db.GetHoldingParams) (db.Holding, error)
	GetForUpdate(ctx context.Context, arg db.GetHoldingForUpdateParams) (db.Holding, error)
	ListByUser(ctx context.Context, userID uint64) ([]db.Holding, error)
	ListByAsset(ctx context.Context, arg db.ListHoldersByAssetParams) ([]db.Holding, error)
	Upsert(ctx context.Context, arg db.UpsertHoldingParams) error
}

type holdingRepo struct{ q *db.Queries }

func (r *holdingRepo) Get(ctx context.Context, arg db.GetHoldingParams) (db.Holding, error) {
	return r.q.GetHolding(ctx, arg)
}
func (r *holdingRepo) GetForUpdate(ctx context.Context, arg db.GetHoldingForUpdateParams) (db.Holding, error) {
	return r.q.GetHoldingForUpdate(ctx, arg)
}
func (r *holdingRepo) ListByUser(ctx context.Context, userID uint64) ([]db.Holding, error) {
	return r.q.ListHoldingsByUser(ctx, userID)
}
func (r *holdingRepo) ListByAsset(ctx context.Context, arg db.ListHoldersByAssetParams) ([]db.Holding, error) {
	return r.q.ListHoldersByAsset(ctx, arg)
}
func (r *holdingRepo) Upsert(ctx context.Context, arg db.UpsertHoldingParams) error {
	return r.q.UpsertHolding(ctx, arg)
}

// DcaRepo is DCA plan + round persistence (FR-13/14/15).
type DcaRepo interface {
	GetPlanByID(ctx context.Context, id uint64) (db.DcaPlan, error)
	ListPlansByUser(ctx context.Context, userID uint64) ([]db.DcaPlan, error)
	ListDuePlans(ctx context.Context, nextRunDate sql.NullTime) ([]db.DcaPlan, error)
	CreatePlan(ctx context.Context, arg db.CreateDcaPlanParams) (sql.Result, error)
	UpdatePlan(ctx context.Context, arg db.UpdateDcaPlanParams) error
	SetPlanStatus(ctx context.Context, arg db.SetDcaPlanStatusParams) error
	RecordRun(ctx context.Context, arg db.RecordDcaRunParams) error
	CreateRound(ctx context.Context, arg db.CreateDcaRoundParams) (sql.Result, error)
	ListRounds(ctx context.Context, dcaPlanID uint64) ([]db.DcaRound, error)
}

type dcaRepo struct{ q *db.Queries }

func (r *dcaRepo) GetPlanByID(ctx context.Context, id uint64) (db.DcaPlan, error) {
	return r.q.GetDcaPlanByID(ctx, id)
}
func (r *dcaRepo) ListPlansByUser(ctx context.Context, userID uint64) ([]db.DcaPlan, error) {
	return r.q.ListDcaPlansByUser(ctx, userID)
}
func (r *dcaRepo) ListDuePlans(ctx context.Context, nextRunDate sql.NullTime) ([]db.DcaPlan, error) {
	return r.q.ListDuePlans(ctx, nextRunDate)
}
func (r *dcaRepo) CreatePlan(ctx context.Context, arg db.CreateDcaPlanParams) (sql.Result, error) {
	return r.q.CreateDcaPlan(ctx, arg)
}
func (r *dcaRepo) UpdatePlan(ctx context.Context, arg db.UpdateDcaPlanParams) error {
	return r.q.UpdateDcaPlan(ctx, arg)
}
func (r *dcaRepo) SetPlanStatus(ctx context.Context, arg db.SetDcaPlanStatusParams) error {
	return r.q.SetDcaPlanStatus(ctx, arg)
}
func (r *dcaRepo) RecordRun(ctx context.Context, arg db.RecordDcaRunParams) error {
	return r.q.RecordDcaRun(ctx, arg)
}
func (r *dcaRepo) CreateRound(ctx context.Context, arg db.CreateDcaRoundParams) (sql.Result, error) {
	return r.q.CreateDcaRound(ctx, arg)
}
func (r *dcaRepo) ListRounds(ctx context.Context, dcaPlanID uint64) ([]db.DcaRound, error) {
	return r.q.ListDcaRounds(ctx, dcaPlanID)
}
