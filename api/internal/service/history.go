package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// History limits (BR-12, EX-14, US-025/028).
const (
	maxHistoryMonths = 12
	maxExportRows    = 10000
)

// HistoryService implements the transaction history feed: list/filter/search/
// sort/paginate, detail with ownership check (EX-15), cancel pending (BR-13),
// and Excel export (components.md § HistoryService; FR-16..20, Add-on xlsx).
type HistoryService struct {
	store  *repo.Store
	wallet *WalletService
	clock  Clock
}

// NewHistoryService builds a HistoryService.
func NewHistoryService(store *repo.Store, wallet *WalletService, clock Clock) *HistoryService {
	return &HistoryService{store: store, wallet: wallet, clock: clock}
}

// HistoryFilter captures the list query params (US-024/025).
type HistoryFilter struct {
	Tab       string // all|orders|dca|cash|income
	AssetType string
	Type      string
	Status    string
	From      string // yyyy-MM-dd
	To        string
	Query     string
	Sort      string // date|amount
	Order     string // asc|desc
	Page      int
	PageSize  int
}

// TxnView is one history row (US-024).
type TxnView struct {
	TxnRef     string `json:"txnRef"`
	Type       string `json:"type"`
	AssetType  string `json:"assetType,omitempty"`
	AssetID    int64  `json:"assetId,omitempty"`
	Amount     string `json:"amount"`
	Quantity   string `json:"quantity,omitempty"`
	Price      string `json:"price,omitempty"`
	NetAmount  string `json:"netAmount"`
	Status     string `json:"status"`
	OccurredAt string `json:"occurredAt"`
	RealizedPL string `json:"realizedPl,omitempty"`
	TradeDate  string `json:"tradeDate,omitempty"`
	SettleDate string `json:"settlementDate,omitempty"`
}

// List returns a filtered, paginated page of the user's transactions (US-024/025).
// A requested range > 12 months is rejected with RANGE_TOO_LARGE (BR-12).
func (s *HistoryService) List(ctx context.Context, userID uint64, f HistoryFilter) ([]TxnView, int64, error) {
	if err := s.validateRange(f); err != nil {
		return nil, 0, err
	}
	// Fetch a generous window then filter/sort in memory (phase-1 scope; the
	// dataset per user is bounded and the base index is user+occurred_at).
	rows, err := s.store.Transactions.ListByUser(ctx, db.ListTransactionsByUserParams{
		UserID: userID, Limit: 5000, Offset: 0,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	filtered := s.applyFilter(rows, f)
	sortTxns(filtered, f)

	total := int64(len(filtered))
	limit, offset := pageBounds(f.Page, f.PageSize)
	end := offset + limit
	if offset > len(filtered) {
		offset = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	page := filtered[offset:end]
	views := make([]TxnView, 0, len(page))
	for _, r := range page {
		views = append(views, toTxnView(r))
	}
	return views, total, nil
}

func (s *HistoryService) validateRange(f HistoryFilter) error {
	if f.From == "" || f.To == "" {
		return nil
	}
	from, err1 := time.ParseInLocation("2006-01-02", f.From, Bangkok)
	to, err2 := time.ParseInLocation("2006-01-02", f.To, Bangkok)
	if err1 != nil || err2 != nil {
		return domain.NewError(domain.CodeValidationError, "Invalid date range")
	}
	if to.Before(from) {
		return domain.NewError(domain.CodeValidationError, "End date is before start date")
	}
	if to.After(from.AddDate(0, maxHistoryMonths, 1)) {
		return domain.NewError(domain.CodeRangeTooLarge, "Please narrow the date range to 12 months or less")
	}
	return nil
}

// applyFilter applies tab/assetType/type/status/date/search filters in memory.
func (s *HistoryService) applyFilter(rows []db.Transaction, f HistoryFilter) []db.Transaction {
	q := strings.ToLower(strings.TrimSpace(f.Query))
	var from, to time.Time
	if f.From != "" {
		from, _ = time.ParseInLocation("2006-01-02", f.From, Bangkok)
	}
	if f.To != "" {
		to, _ = time.ParseInLocation("2006-01-02", f.To, Bangkok)
		to = to.AddDate(0, 0, 1)
	}
	out := make([]db.Transaction, 0, len(rows))
	for _, r := range rows {
		if !tabMatches(f.Tab, r.Type) {
			continue
		}
		if f.AssetType != "" && !(r.AssetType.Valid && string(r.AssetType.TransactionsAssetType) == f.AssetType) {
			continue
		}
		if f.Type != "" && string(r.Type) != f.Type {
			continue
		}
		if f.Status != "" && string(r.Status) != f.Status {
			continue
		}
		if !from.IsZero() && r.OccurredAt.Before(from) {
			continue
		}
		if !to.IsZero() && !r.OccurredAt.Before(to) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(r.TxnRef), q) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// tabMatches maps a tab to the transaction types it includes (US-025).
func tabMatches(tab string, t db.TransactionsType) bool {
	switch tab {
	case "", "all":
		return true
	case "orders":
		return t == db.TransactionsTypeBuy || t == db.TransactionsTypeSell
	case "dca":
		return t == db.TransactionsTypeDcaBuy
	case "cash":
		return t == db.TransactionsTypeDeposit || t == db.TransactionsTypeWithdraw
	case "income":
		return t == db.TransactionsTypeCoupon || t == db.TransactionsTypeMaturity
	default:
		return true
	}
}

func sortTxns(rows []db.Transaction, f HistoryFilter) {
	desc := !strings.EqualFold(f.Order, "asc") // default newest/largest first
	byAmount := strings.EqualFold(f.Sort, "amount")
	sort.SliceStable(rows, func(i, j int) bool {
		if byAmount {
			c := rows[i].Amount.Cmp(rows[j].Amount)
			if desc {
				return c > 0
			}
			return c < 0
		}
		if desc {
			return rows[i].OccurredAt.After(rows[j].OccurredAt)
		}
		return rows[i].OccurredAt.Before(rows[j].OccurredAt)
	})
}

// TxnDetail is the detail view with a status timeline (US-026).
type TxnDetail struct {
	Transaction TxnView            `json:"transaction"`
	Timeline    []StatusTimelinePt `json:"timeline"`
	Fee         string             `json:"fee"`
	AccruedInt  string             `json:"accruedInterest,omitempty"`
	DcaPlanID   int64              `json:"dcaPlanId,omitempty"`
	DcaRound    int32              `json:"dcaRound,omitempty"`
}

// StatusTimelinePt is one status-history entry (US-026).
type StatusTimelinePt struct {
	Status    string `json:"status"`
	Note      string `json:"note,omitempty"`
	ChangedAt string `json:"changedAt"`
}

// Detail returns a transaction with its status timeline, enforcing ownership —
// another user's ref returns TXN_NOT_FOUND (EX-15).
func (s *HistoryService) Detail(ctx context.Context, userID uint64, txnRef string) (TxnDetail, error) {
	r, err := s.store.Transactions.GetByTxnRef(ctx, txnRef)
	if err != nil || r.UserID != userID {
		return TxnDetail{}, domain.NewError(domain.CodeTxnNotFound, "Transaction not found")
	}
	hist, _ := s.store.Transactions.ListStatusHistory(ctx, r.ID)
	tl := make([]StatusTimelinePt, 0, len(hist))
	for _, h := range hist {
		tl = append(tl, StatusTimelinePt{
			Status: h.Status, Note: h.Note.String, ChangedAt: h.ChangedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}
	d := TxnDetail{Transaction: toTxnView(r), Timeline: tl, Fee: money.FormatTHB(r.Fee)}
	if r.AccruedInterest.Valid {
		d.AccruedInt = money.String(r.AccruedInterest.Decimal)
	}
	if r.DcaPlanID.Valid {
		d.DcaPlanID = r.DcaPlanID.Int64
	}
	if r.DcaRound.Valid {
		d.DcaRound = r.DcaRound.Int32
	}
	return d, nil
}

// Cancel cancels an eligible pending order (BR-13, US-027): only a pending MF
// Buy/Sell (or DCA buy) before cut-off. Releases reserved cash and appends to
// the status timeline. Other cases → CANCEL_NOT_ALLOWED.
func (s *HistoryService) Cancel(ctx context.Context, userID uint64, txnRef string) error {
	return s.store.WithinTx(ctx, func(tx *repo.Store) error {
		t, err := tx.Transactions.GetByTxnRef(ctx, txnRef)
		if err != nil || t.UserID != userID {
			return domain.NewError(domain.CodeTxnNotFound, "Transaction not found")
		}
		// Only fund buy/sell and dca_buy are cancellable, and only while pending.
		cancellable := (t.Type == db.TransactionsTypeBuy || t.Type == db.TransactionsTypeSell || t.Type == db.TransactionsTypeDcaBuy) &&
			t.AssetType.Valid && t.AssetType.TransactionsAssetType == db.TransactionsAssetTypeFund
		if !cancellable || t.Status != db.TransactionsStatusSubmitted && t.Status != db.TransactionsStatusProcessing {
			return domain.NewError(domain.CodeCancelNotAllowed, "This transaction cannot be cancelled")
		}
		if !t.OrderID.Valid {
			return domain.NewError(domain.CodeCancelNotAllowed, "This transaction cannot be cancelled")
		}
		order, err := tx.Orders.GetByID(ctx, uint64(t.OrderID.Int64))
		if err != nil {
			return domain.NewError(domain.CodeCancelNotAllowed, "This transaction cannot be cancelled")
		}
		if order.Status != db.OrdersStatusPending {
			return domain.NewError(domain.CodeCancelNotAllowed, "Order is no longer pending")
		}
		// Release reserved cash for a buy.
		if order.Side == db.OrdersSideBuy && order.Amount.Valid {
			if err := s.wallet.ReleaseTx(ctx, tx, userID, order.Amount.Decimal, int64(t.ID)); err != nil {
				return err
			}
		}
		if err := tx.Orders.UpdateStatus(ctx, db.UpdateOrderStatusParams{Status: db.OrdersStatusCancelled, ID: order.ID}); err != nil {
			return fmt.Errorf("cancel order: %w", err)
		}
		if err := tx.Transactions.UpdateStatus(ctx, db.UpdateTransactionStatusParams{Status: db.TransactionsStatusCancelled, ID: t.ID}); err != nil {
			return fmt.Errorf("cancel transaction: %w", err)
		}
		_, err = tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
			TransactionID: t.ID, Status: "cancelled", Note: nullString("cancelled by user"),
			ChangedBy: nullString("user"), ChangedAt: s.clock.Now(),
		})
		return err
	})
}

// toTxnView maps a db.Transaction to the API view.
func toTxnView(r db.Transaction) TxnView {
	v := TxnView{
		TxnRef: r.TxnRef, Type: string(r.Type),
		Amount: money.FormatTHB(r.Amount), NetAmount: money.FormatTHB(r.NetAmount),
		Status: string(r.Status), OccurredAt: r.OccurredAt.Format("2006-01-02T15:04:05-07:00"),
	}
	if r.AssetType.Valid {
		v.AssetType = string(r.AssetType.TransactionsAssetType)
	}
	if r.AssetID.Valid {
		v.AssetID = r.AssetID.Int64
	}
	if r.Quantity.Valid {
		v.Quantity = money.FormatQuantity(r.Quantity.Decimal)
	}
	if r.Price.Valid {
		v.Price = money.String(r.Price.Decimal)
	}
	if r.RealizedPl.Valid {
		v.RealizedPL = money.String(r.RealizedPl.Decimal)
	}
	if r.TradeDate.Valid {
		v.TradeDate = r.TradeDate.Time.Format("2006-01-02")
	}
	if r.SettlementDate.Valid {
		v.SettleDate = r.SettlementDate.Time.Format("2006-01-02")
	}
	return v
}
