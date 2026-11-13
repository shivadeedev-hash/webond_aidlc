package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/webond/api/internal/db"
)

// FundRepo is mutual-fund master + NAV persistence (FR-05/06).
type FundRepo interface {
	GetByID(ctx context.Context, id uint64) (db.Fund, error)
	GetByCode(ctx context.Context, code string) (db.Fund, error)
	List(ctx context.Context, arg db.ListFundsParams) ([]db.Fund, error)
	Create(ctx context.Context, arg db.CreateFundParams) (sql.Result, error)
	GetLatestNav(ctx context.Context, fundID uint64) (db.FundNav, error)
	ListNavs(ctx context.Context, arg db.ListFundNavsParams) ([]db.FundNav, error)
	UpsertNav(ctx context.Context, arg db.UpsertFundNavParams) error
}

type fundRepo struct{ q *db.Queries }

func (r *fundRepo) GetByID(ctx context.Context, id uint64) (db.Fund, error) {
	return r.q.GetFundByID(ctx, id)
}
func (r *fundRepo) GetByCode(ctx context.Context, code string) (db.Fund, error) {
	return r.q.GetFundByCode(ctx, code)
}
func (r *fundRepo) List(ctx context.Context, arg db.ListFundsParams) ([]db.Fund, error) {
	return r.q.ListFunds(ctx, arg)
}
func (r *fundRepo) Create(ctx context.Context, arg db.CreateFundParams) (sql.Result, error) {
	return r.q.CreateFund(ctx, arg)
}
func (r *fundRepo) GetLatestNav(ctx context.Context, fundID uint64) (db.FundNav, error) {
	return r.q.GetLatestFundNav(ctx, fundID)
}
func (r *fundRepo) ListNavs(ctx context.Context, arg db.ListFundNavsParams) ([]db.FundNav, error) {
	return r.q.ListFundNavs(ctx, arg)
}
func (r *fundRepo) UpsertNav(ctx context.Context, arg db.UpsertFundNavParams) error {
	return r.q.UpsertFundNav(ctx, arg)
}

// BondRepo is bond master + price + coupon schedule persistence (FR-05/06/23).
type BondRepo interface {
	GetByID(ctx context.Context, id uint64) (db.Bond, error)
	GetByCode(ctx context.Context, code string) (db.Bond, error)
	List(ctx context.Context, arg db.ListBondsParams) ([]db.Bond, error)
	Create(ctx context.Context, arg db.CreateBondParams) (sql.Result, error)
	GetLatestPrice(ctx context.Context, bondID uint64) (db.BondPrice, error)
	ListPrices(ctx context.Context, arg db.ListBondPricesParams) ([]db.BondPrice, error)
	UpsertPrice(ctx context.Context, arg db.UpsertBondPriceParams) error
	CreateCouponSchedule(ctx context.Context, arg db.CreateBondCouponScheduleParams) (sql.Result, error)
	ListUnpaidCouponsDue(ctx context.Context, couponDate time.Time) ([]db.BondCouponSchedule, error)
	MarkCouponPaid(ctx context.Context, id uint64) error
}

type bondRepo struct{ q *db.Queries }

func (r *bondRepo) GetByID(ctx context.Context, id uint64) (db.Bond, error) {
	return r.q.GetBondByID(ctx, id)
}
func (r *bondRepo) GetByCode(ctx context.Context, code string) (db.Bond, error) {
	return r.q.GetBondByCode(ctx, code)
}
func (r *bondRepo) List(ctx context.Context, arg db.ListBondsParams) ([]db.Bond, error) {
	return r.q.ListBonds(ctx, arg)
}
func (r *bondRepo) Create(ctx context.Context, arg db.CreateBondParams) (sql.Result, error) {
	return r.q.CreateBond(ctx, arg)
}
func (r *bondRepo) GetLatestPrice(ctx context.Context, bondID uint64) (db.BondPrice, error) {
	return r.q.GetLatestBondPrice(ctx, bondID)
}
func (r *bondRepo) ListPrices(ctx context.Context, arg db.ListBondPricesParams) ([]db.BondPrice, error) {
	return r.q.ListBondPrices(ctx, arg)
}
func (r *bondRepo) UpsertPrice(ctx context.Context, arg db.UpsertBondPriceParams) error {
	return r.q.UpsertBondPrice(ctx, arg)
}
func (r *bondRepo) CreateCouponSchedule(ctx context.Context, arg db.CreateBondCouponScheduleParams) (sql.Result, error) {
	return r.q.CreateBondCouponSchedule(ctx, arg)
}
func (r *bondRepo) ListUnpaidCouponsDue(ctx context.Context, couponDate time.Time) ([]db.BondCouponSchedule, error) {
	return r.q.ListUnpaidCouponsDue(ctx, couponDate)
}
func (r *bondRepo) MarkCouponPaid(ctx context.Context, id uint64) error {
	return r.q.MarkCouponPaid(ctx, id)
}

// GoldRepo is gold master + price persistence (FR-05/06, BR-09).
type GoldRepo interface {
	GetByID(ctx context.Context, id uint64) (db.GoldProduct, error)
	GetByCode(ctx context.Context, code string) (db.GoldProduct, error)
	List(ctx context.Context, arg db.ListGoldProductsParams) ([]db.GoldProduct, error)
	Create(ctx context.Context, arg db.CreateGoldProductParams) (sql.Result, error)
	GetLatestPrice(ctx context.Context, goldProductID uint64) (db.GoldPrice, error)
	ListPrices(ctx context.Context, arg db.ListGoldPricesParams) ([]db.GoldPrice, error)
	CreatePrice(ctx context.Context, arg db.CreateGoldPriceParams) (sql.Result, error)
}

type goldRepo struct{ q *db.Queries }

func (r *goldRepo) GetByID(ctx context.Context, id uint64) (db.GoldProduct, error) {
	return r.q.GetGoldProductByID(ctx, id)
}
func (r *goldRepo) GetByCode(ctx context.Context, code string) (db.GoldProduct, error) {
	return r.q.GetGoldProductByCode(ctx, code)
}
func (r *goldRepo) List(ctx context.Context, arg db.ListGoldProductsParams) ([]db.GoldProduct, error) {
	return r.q.ListGoldProducts(ctx, arg)
}
func (r *goldRepo) Create(ctx context.Context, arg db.CreateGoldProductParams) (sql.Result, error) {
	return r.q.CreateGoldProduct(ctx, arg)
}
func (r *goldRepo) GetLatestPrice(ctx context.Context, goldProductID uint64) (db.GoldPrice, error) {
	return r.q.GetLatestGoldPrice(ctx, goldProductID)
}
func (r *goldRepo) ListPrices(ctx context.Context, arg db.ListGoldPricesParams) ([]db.GoldPrice, error) {
	return r.q.ListGoldPrices(ctx, arg)
}
func (r *goldRepo) CreatePrice(ctx context.Context, arg db.CreateGoldPriceParams) (sql.Result, error) {
	return r.q.CreateGoldPrice(ctx, arg)
}
