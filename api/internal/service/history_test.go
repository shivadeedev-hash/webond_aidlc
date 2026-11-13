package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

// seedPendingFundOrder creates a pending MF buy order with a matching submitted
// transaction and reserves the cash, mirroring a pre-cut-off order awaiting
// settlement. Returns the txn_ref so a Cancel test can target it (BR-13).
func seedPendingFundOrder(t *testing.T, store *repo.Store, ws *WalletService, clock Clock, userID, fundID uint64, amount decimal.Decimal) string {
	t.Helper()
	ctx := context.Background()
	txnRef := NewTxnRef(clock.Now())
	err := store.WithinTx(ctx, func(tx *repo.Store) error {
		if err := ws.ReserveTx(ctx, tx, userID, amount, 0); err != nil {
			return err
		}
		orderID, err := lastInsertID(tx.Orders.Create(ctx, db.CreateOrderParams{
			TxnRef: txnRef, UserID: userID, AssetType: db.OrdersAssetTypeFund, AssetID: fundID,
			Side: db.OrdersSideBuy, OrderMode: db.OrdersOrderModeAmount,
			Amount:    decimal.NullDecimal{Decimal: amount, Valid: true},
			NetAmount: decimal.NullDecimal{Decimal: amount, Valid: true},
			Fee:       decimal.Zero, Status: db.OrdersStatusPending, Channel: "web",
		}))
		if err != nil {
			return err
		}
		txnID, err := lastInsertID(tx.Transactions.Create(ctx, db.CreateTransactionParams{
			TxnRef: txnRef, UserID: userID, Type: db.TransactionsTypeBuy,
			AssetType: db.NullTransactionsAssetType{TransactionsAssetType: db.TransactionsAssetTypeFund, Valid: true},
			AssetID:   sql.NullInt64{Int64: int64(fundID), Valid: true},
			Amount:    amount.Neg(), NetAmount: amount, Fee: decimal.Zero,
			Status:     db.TransactionsStatusSubmitted,
			OrderID:    sql.NullInt64{Int64: int64(orderID), Valid: true},
			OccurredAt: clock.Now(),
		}))
		if err != nil {
			return err
		}
		_, err = tx.Transactions.CreateStatusHistory(ctx, db.CreateTransactionStatusHistoryParams{
			TransactionID: txnID, Status: "submitted", ChangedBy: nullString("system"), ChangedAt: clock.Now(),
		})
		return err
	})
	if err != nil {
		t.Fatalf("seedPendingFundOrder: %v", err)
	}
	return txnRef
}

// TestHistory_ListFilterSortPaginate covers list filtering (tab + asset type),
// sorting and pagination (history.go List/applyFilter/sortTxns; US-024/025).
func TestHistory_ListFilterSortPaginate(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "HF"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	// Two completed buys → two rows in the transaction feed (orders tab).
	for _, amt := range []string{"5000", "7000"} {
		if _, err := ts.PlaceOrder(ctx, OrderInput{
			UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
			OrderMode: "amount", Amount: decimal.RequireFromString(amt),
		}); err != nil {
			t.Fatalf("buy %s: %v", amt, err)
		}
	}

	// All tab: at least the two buys.
	all, total, err := hs.List(ctx, userID, HistoryFilter{Tab: "all", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if total < 2 || len(all) < 2 {
		t.Fatalf("expected >=2 txns, got total=%d page=%d", total, len(all))
	}

	// Orders tab: only buy/sell rows.
	orders, _, err := hs.List(ctx, userID, HistoryFilter{Tab: "orders", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	for _, v := range orders {
		if v.Type != "buy" && v.Type != "sell" {
			t.Fatalf("orders tab leaked type %s", v.Type)
		}
	}

	// Income tab filters to coupon/maturity only (none here → empty).
	income, _, err := hs.List(ctx, userID, HistoryFilter{Tab: "income", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list income: %v", err)
	}
	for _, v := range income {
		if v.Type != "coupon" && v.Type != "maturity" {
			t.Fatalf("income tab leaked type %s", v.Type)
		}
	}

	// Asset-type + search filters.
	bySearch, _, err := hs.List(ctx, userID, HistoryFilter{AssetType: "fund", Query: all[0].TxnRef, Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list by search: %v", err)
	}
	if len(bySearch) != 1 {
		t.Fatalf("expected exactly 1 row for an exact txnRef search, got %d", len(bySearch))
	}

	// Sort by amount desc: first row amount >= second.
	byAmt, _, err := hs.List(ctx, userID, HistoryFilter{Tab: "orders", Sort: "amount", Order: "desc", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("list sorted: %v", err)
	}
	if len(byAmt) >= 2 {
		a0 := decimalFromString(byAmt[0].Amount)
		a1 := decimalFromString(byAmt[1].Amount)
		if a0.LessThan(a1) {
			t.Fatalf("amount desc sort broken: %s < %s", byAmt[0].Amount, byAmt[1].Amount)
		}
	}

	// Pagination: pageSize 1 returns a single row but the real total.
	page1, totalP, err := hs.List(ctx, userID, HistoryFilter{Tab: "orders", Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("list paginated: %v", err)
	}
	if len(page1) != 1 || totalP < 2 {
		t.Fatalf("pagination: page len=%d total=%d", len(page1), totalP)
	}
}

// TestHistory_RangeTooLarge rejects a >12-month window (history.go validateRange;
// BR-12).
func TestHistory_RangeTooLarge(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	_, _, err := hs.List(ctx, userID, HistoryFilter{From: "2023-01-01", To: "2024-12-31"})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeRangeTooLarge {
		t.Fatalf("expected RANGE_TOO_LARGE, got %v", err)
	}

	// Invalid order (to before from) → validation error.
	_, _, err = hs.List(ctx, userID, HistoryFilter{From: "2024-06-01", To: "2024-01-01"})
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR for reversed range, got %v", err)
	}
}

// TestHistory_DetailOwnership returns detail for the owner and TXN_NOT_FOUND for
// another user (history.go Detail; EX-15).
func TestHistory_DetailOwnership(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	ownerID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	otherID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "HD"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() {
		cleanupUser(t, pool, ownerID)
		cleanupUser(t, pool, otherID)
		cleanupAsset(t, pool, "fund", fundID)
	})

	res, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: ownerID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	})
	if err != nil {
		t.Fatalf("buy: %v", err)
	}

	d, err := hs.Detail(ctx, ownerID, res.TxnRef)
	if err != nil {
		t.Fatalf("owner detail: %v", err)
	}
	if d.Transaction.TxnRef != res.TxnRef || len(d.Timeline) == 0 {
		t.Fatalf("detail missing data: %+v", d)
	}

	// Another user must not see it (EX-15).
	_, err = hs.Detail(ctx, otherID, res.TxnRef)
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeTxnNotFound {
		t.Fatalf("expected TXN_NOT_FOUND for non-owner, got %v", err)
	}
}

// TestHistory_CancelPending cancels a pending MF order, releasing the reserved
// cash; a completed order cannot be cancelled (history.go Cancel; BR-13, US-027).
func TestHistory_CancelPending(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "HC"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	// Pending order reserves 5000.
	ref := seedPendingFundOrder(t, store, ws, clock, userID, fundID, decimal.RequireFromString("5000"))
	before, _ := ws.Get(ctx, userID)
	if !before.Reserved.Equal(decimal.RequireFromString("5000.0000")) {
		t.Fatalf("expected 5000 reserved, got %s", before.Reserved)
	}

	if err := hs.Cancel(ctx, userID, ref); err != nil {
		t.Fatalf("cancel pending: %v", err)
	}
	after, _ := ws.Get(ctx, userID)
	if !after.Reserved.Equal(decimal.Zero.Round(4)) {
		t.Fatalf("reserved should be released, got %s", after.Reserved)
	}
	if !after.Available.Equal(before.Balance) {
		t.Fatalf("available should return to full balance, got %s", after.Available)
	}

	// Cancelling again → CANCEL_NOT_ALLOWED (no longer pending).
	err := hs.Cancel(ctx, userID, ref)
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeCancelNotAllowed {
		t.Fatalf("expected CANCEL_NOT_ALLOWED on re-cancel, got %v", err)
	}

	// A completed buy cannot be cancelled.
	done, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	})
	if err != nil {
		t.Fatalf("completed buy: %v", err)
	}
	err = hs.Cancel(ctx, userID, done.TxnRef)
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeCancelNotAllowed {
		t.Fatalf("expected CANCEL_NOT_ALLOWED for completed order, got %v", err)
	}
}

// TestHistory_ExportAndStatement covers the xlsx export happy path and the
// monthly statement PDF incl. its validation (history_export.go Export /
// statement.go MonthlyStatement; US-028 Add-on, EX-14).
func TestHistory_ExportAndStatement(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	sfx := shortID()
	fundID := seedFund(t, store, clock, "HE"+sfx, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	if _, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	}); err != nil {
		t.Fatalf("buy: %v", err)
	}

	data, filename, err := hs.Export(ctx, userID, HistoryFilter{Tab: "all"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	// .xlsx files are zip archives → start with "PK".
	if len(data) < 2 || data[0] != 'P' || data[1] != 'K' {
		t.Fatalf("export does not look like an xlsx (len=%d)", len(data))
	}
	if filename == "" {
		t.Fatal("export filename empty")
	}

	// Monthly statement (PDF).
	body, sName, err := hs.MonthlyStatement(ctx, userID, "202610")
	if err != nil {
		t.Fatalf("statement: %v", err)
	}
	if len(body) == 0 || sName == "" {
		t.Fatalf("statement empty: name=%s len=%d", sName, len(body))
	}
	// Bad format rejected.
	if _, _, err := hs.MonthlyStatement(ctx, userID, "2026-10"); err == nil {
		t.Fatal("expected validation error for bad month format")
	}
}
