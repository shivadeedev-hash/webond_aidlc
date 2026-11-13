package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// BondLifecycleService credits bond coupons and closes out matured bonds
// (components.md § BondLifecycleService; FR-23, BR-05). Manual-trigger in
// phase 1 (D1-3) via the admin run endpoints.
type BondLifecycleService struct {
	store  *repo.Store
	wallet *WalletService
	notify *NotificationService
	clock  Clock
}

// NewBondLifecycleService builds a BondLifecycleService.
func NewBondLifecycleService(store *repo.Store, wallet *WalletService, notify *NotificationService, clock Clock) *BondLifecycleService {
	return &BondLifecycleService{store: store, wallet: wallet, notify: notify, clock: clock}
}

// RunCoupons credits all unpaid coupons due on/before "now" to holders, writing
// a coupon transaction + wallet credit per holder and marking the schedule paid
// (FR-23). Returns the number of coupon schedules processed.
func (s *BondLifecycleService) RunCoupons(ctx context.Context) (int, error) {
	due, err := s.store.Bonds.ListUnpaidCouponsDue(ctx, s.clock.Now())
	if err != nil {
		return 0, fmt.Errorf("list due coupons: %w", err)
	}
	processed := 0
	for _, c := range due {
		if err := s.payCoupon(ctx, c); err != nil {
			// Log-and-continue: a single failing coupon must not abort the batch.
			continue
		}
		processed++
	}
	return processed, nil
}

// payCoupon credits a single coupon schedule to every current holder of the
// bond, inside one transaction (atomic per coupon schedule).
func (s *BondLifecycleService) payCoupon(ctx context.Context, c db.BondCouponSchedule) error {
	return s.store.WithinTx(ctx, func(tx *repo.Store) error {
		holders, err := s.holdersOf(ctx, tx, db.HoldingsAssetTypeBond, c.BondID)
		if err != nil {
			return err
		}
		for _, h := range holders {
			coupon := money.Quantize(c.AmountPerUnit.Mul(h.Quantity))
			if coupon.LessThanOrEqual(decimal.Zero) {
				continue
			}
			if err := s.wallet.CreditTx(ctx, tx, h.UserID, coupon, db.WalletTransactionsTypeCoupon, 0); err != nil {
				return err
			}
			txnRef := NewTxnRef(s.clock.Now())
			txnID, err := lastInsertID(tx.Transactions.Create(ctx, db.CreateTransactionParams{
				TxnRef: txnRef, UserID: h.UserID, Type: db.TransactionsTypeCoupon,
				AssetType: db.NullTransactionsAssetType{TransactionsAssetType: db.TransactionsAssetTypeBond, Valid: true},
				AssetID:   nullInt64(int64(c.BondID)),
				Amount:    money.SignedAmount(money.TxnCoupon, coupon), NetAmount: money.Quantize(coupon),
				Fee: decimal.Zero, Status: db.TransactionsStatusCompleted, OccurredAt: s.clock.Now(),
			}))
			if err != nil {
				return fmt.Errorf("coupon txn: %w", err)
			}
			_, _ = tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
				TransactionID: txnID, Status: "completed", ChangedBy: nullString("system"), ChangedAt: s.clock.Now(),
			})
			_ = s.notify.EmitTx(ctx, tx, Notice{
				UserID: h.UserID, Category: db.NotificationOutboxCategoryCoupon,
				Subject: "Coupon received", Body: fmt.Sprintf("You received a coupon of %s THB", money.FormatTHB(coupon)),
				EmailEnabled: s.emailEnabled(ctx, tx, h.UserID),
			})
		}
		return tx.Bonds.MarkCouponPaid(ctx, c.ID)
	})
}

// RunMaturities closes out bonds that matured on/before "now": credits face
// value to each holder, zeroes the holding, and records a maturity transaction
// (FR-23, BR-05). Returns the number of (bond,holder) closeouts performed.
func (s *BondLifecycleService) RunMaturities(ctx context.Context) (int, error) {
	bonds, err := s.store.Bonds.List(ctx, db.ListBondsParams{Limit: 1000, Offset: 0})
	if err != nil {
		return 0, fmt.Errorf("list bonds: %w", err)
	}
	now := s.clock.Now()
	count := 0
	for _, b := range bonds {
		if !b.MaturityDate.Valid || b.MaturityDate.Time.After(now) {
			continue
		}
		n, _ := s.matureBond(ctx, b)
		count += n
	}
	return count, nil
}

func (s *BondLifecycleService) matureBond(ctx context.Context, b db.Bond) (int, error) {
	var count int
	err := s.store.WithinTx(ctx, func(tx *repo.Store) error {
		holders, err := s.holdersOf(ctx, tx, db.HoldingsAssetTypeBond, b.ID)
		if err != nil {
			return err
		}
		for _, h := range holders {
			if h.Quantity.LessThanOrEqual(decimal.Zero) {
				continue
			}
			faceTotal := money.BondFaceValue(h.Quantity, b.FaceValue)
			if err := s.wallet.CreditTx(ctx, tx, h.UserID, faceTotal, db.WalletTransactionsTypeMaturity, 0); err != nil {
				return err
			}
			// Zero out the holding.
			if err := tx.Holdings.Upsert(ctx, db.UpsertHoldingParams{
				UserID: h.UserID, AssetType: db.HoldingsAssetTypeBond, AssetID: b.ID,
				Quantity: decimal.Zero, LockedQuantity: decimal.Zero, AvgCost: decimal.Zero, TotalCost: decimal.Zero,
			}); err != nil {
				return fmt.Errorf("zero holding: %w", err)
			}
			txnRef := NewTxnRef(s.clock.Now())
			txnID, err := lastInsertID(tx.Transactions.Create(ctx, db.CreateTransactionParams{
				TxnRef: txnRef, UserID: h.UserID, Type: db.TransactionsTypeMaturity,
				AssetType: db.NullTransactionsAssetType{TransactionsAssetType: db.TransactionsAssetTypeBond, Valid: true},
				AssetID:   nullInt64(int64(b.ID)),
				Amount:    money.SignedAmount(money.TxnMaturity, faceTotal), NetAmount: faceTotal,
				Fee: decimal.Zero, Status: db.TransactionsStatusCompleted, OccurredAt: s.clock.Now(),
			}))
			if err != nil {
				return fmt.Errorf("maturity txn: %w", err)
			}
			_, _ = tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
				TransactionID: txnID, Status: "completed", ChangedBy: nullString("system"), ChangedAt: s.clock.Now(),
			})
			_ = s.notify.EmitTx(ctx, tx, Notice{
				UserID: h.UserID, Category: db.NotificationOutboxCategoryMaturity,
				Subject: "Bond matured", Body: fmt.Sprintf("Bond %s matured; %s THB credited", b.Code, money.FormatTHB(faceTotal)),
				EmailEnabled: s.emailEnabled(ctx, tx, h.UserID),
			})
			count++
		}
		return nil
	})
	return count, err
}

// holdersOf returns all users currently holding a given asset by scanning
// holdings. Phase-1: a full scan via a direct query is acceptable for the
// seeded dataset; there is no dedicated "holders by asset" sqlc query, so this
// uses the pool-level query through the tx store's underlying connection.
func (s *BondLifecycleService) holdersOf(ctx context.Context, tx *repo.Store, assetType db.HoldingsAssetType, assetID uint64) ([]db.Holding, error) {
	return tx.HoldersByAsset(ctx, assetType, assetID)
}

func (s *BondLifecycleService) emailEnabled(ctx context.Context, tx *repo.Store, userID uint64) bool {
	u, err := tx.Users.GetByID(ctx, userID)
	if err != nil {
		return false
	}
	return u.EmailNotificationsEnabled
}
