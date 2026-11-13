package service

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// uncompressedPDF turns PDF stream compression off for a test so the drawn
// text can be asserted on, and restores it afterwards.
func uncompressedPDF(t *testing.T) {
	t.Helper()
	prev := pdfCompress
	pdfCompress = false
	t.Cleanup(func() { pdfCompress = prev })
}

func TestStatement_PDFContentTotalsAndIsolation(t *testing.T) {
	uncompressedPDF(t)
	store := openTestStore(t)
	pool := rawPool(t)
	// Wallet-ledger rows are stamped with the database clock (created_at
	// default), so this test uses the real current time to keep the transaction
	// feed and the ledger in the same month, exactly as in production.
	clock := NewFixedClock(time.Now().In(Bangkok))
	month := clock.Now().Format("200601")
	ws := NewWalletService(store, clock)
	notify := NewNotificationService(store, clock)
	ts := NewTradingService(store, ws, notify, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	otherID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	sfx := shortID()
	fundCode := "ST" + sfx
	fundID := seedFund(t, store, clock, fundCode, decimal.RequireFromString("10"), 3)
	t.Cleanup(func() {
		cleanupUser(t, pool, userID)
		cleanupUser(t, pool, otherID)
		cleanupAsset(t, pool, "fund", fundID)
	})

	if _, err := ws.Deposit(ctx, userID, decimal.RequireFromString("20000")); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	buy, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	})
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	// Another user's activity must never leak into this user's statement.
	otherBuy, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: otherID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("1000"),
	})
	if err != nil {
		t.Fatalf("other buy: %v", err)
	}

	data, name, err := hs.MonthlyStatement(ctx, userID, month)
	if err != nil {
		t.Fatalf("statement: %v", err)
	}
	if name != "webond-statement-"+month+".pdf" {
		t.Errorf("filename = %q, want webond-statement-%s.pdf", name, month)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || !bytes.Contains(data[len(data)-32:], []byte("%%EOF")) {
		t.Fatalf("output is not a well-formed PDF (len=%d)", len(data))
	}

	text := string(data)
	wallet, err := ws.Get(ctx, userID)
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	for _, want := range []string{
		"Monthly Statement - " + clock.Now().Format("January 2006"),
		`Transactions \(2\)`, // the deposit + the buy (PDF escapes parentheses)
		buy.TxnRef,
		fundCode,
		"20,000.00",               // deposits total
		"5,000.00",                // purchases total
		moneyText(wallet.Balance), // closing cash balance == current balance (all activity is in October)
	} {
		if !strings.Contains(text, want) {
			t.Errorf("statement is missing %q", want)
		}
	}
	if strings.Contains(text, otherBuy.TxnRef) {
		t.Errorf("statement leaked another user's transaction %s", otherBuy.TxnRef)
	}
}

func TestStatement_EmptyMonthAndValidation(t *testing.T) {
	uncompressedPDF(t)
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock() // 2026-10-28
	ws := NewWalletService(store, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()
	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// A past month with no activity still yields a valid PDF saying so.
	data, _, err := hs.MonthlyStatement(ctx, userID, "202601")
	if err != nil {
		t.Fatalf("empty month: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || !strings.Contains(string(data), "No transactions in this month.") {
		t.Errorf("empty-month statement should be a PDF with the empty notice")
	}

	for name, month := range map[string]string{
		"future month":    "202611",
		"month 13":        "202613",
		"month 00":        "202600",
		"dash format":     "2026-10",
		"too short":       "2026",
		"letters":         "abcdef",
		"empty":           "",
		"far future year": "209901",
	} {
		if _, _, err := hs.MonthlyStatement(ctx, userID, month); err == nil {
			t.Errorf("%s (%q): expected a validation error", name, month)
		}
	}
}

func TestMoneyText(t *testing.T) {
	cases := map[string]string{
		"0":            "0.00",
		"5":            "5.00",
		"999.999":      "1,000.00",
		"1234":         "1,234.00",
		"1234567.8":    "1,234,567.80",
		"-1234567.8":   "-1,234,567.80",
		"-0.5":         "-0.50",
		"100000":       "100,000.00",
		"123456789.12": "123,456,789.12",
	}
	for in, want := range cases {
		if got := moneyText(decimal.RequireFromString(in)); got != want {
			t.Errorf("moneyText(%s) = %q, want %q", in, got, want)
		}
	}
}

func TestLatin(t *testing.T) {
	if got := latin("abc 123-@"); got != "abc 123-@" {
		t.Errorf("printable ASCII must pass through, got %q", got)
	}
	if got := latin("สวัสดี ok"); got != "?????? ok" {
		t.Errorf("non-ASCII must become '?', got %q", got)
	}
	if got := latin("line\nbreak"); got != "line?break" {
		t.Errorf("control characters must become '?', got %q", got)
	}
}

// US-024: History must list Deposit and Withdraw. They used to be written to
// the wallet ledger only, so a user's own deposits never showed up under the
// "cash" tab. Each movement now also writes a feed row with a status timeline,
// and the ledger row points back at it.
func TestWallet_DepositWithdrawAppearInHistory(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	hs := NewHistoryService(store, ws, clock)
	ctx := context.Background()
	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	if _, err := ws.Deposit(ctx, userID, decimal.RequireFromString("2500")); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if _, err := ws.Withdraw(ctx, userID, decimal.RequireFromString("400")); err != nil {
		t.Fatalf("withdraw: %v", err)
	}

	views, total, err := hs.List(ctx, userID, HistoryFilter{Tab: "cash", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list cash: %v", err)
	}
	if total != 2 || len(views) != 2 {
		t.Fatalf("cash tab has %d rows (total %d), want 2", len(views), total)
	}
	byType := map[string]TxnView{}
	for _, v := range views {
		byType[v.Type] = v
	}
	dep, wd := byType["deposit"], byType["withdraw"]
	if dep.Amount != "2500.00" || dep.Status != "completed" || dep.TxnRef == "" {
		t.Errorf("deposit row = %+v, want +2500.00 completed with a reference", dep)
	}
	if wd.Amount != "-400.00" || wd.Status != "completed" {
		t.Errorf("withdraw row = %+v, want -400.00 completed", wd)
	}

	// The detail view works and has a timeline entry.
	detail, err := hs.Detail(ctx, userID, dep.TxnRef)
	if err != nil {
		t.Fatalf("deposit detail: %v", err)
	}
	if len(detail.Timeline) == 0 || detail.Timeline[len(detail.Timeline)-1].Status != "completed" {
		t.Errorf("deposit timeline = %+v, want a completed entry", detail.Timeline)
	}

	// The ledger row links back to the feed row.
	var ref sql.NullInt64
	if err := pool.QueryRow(
		"SELECT wt.ref_txn_id FROM wallet_transactions wt JOIN wallets w ON w.id = wt.wallet_id WHERE w.user_id = ? AND wt.type = 'deposit' ORDER BY wt.id DESC LIMIT 1",
		userID).Scan(&ref); err != nil {
		t.Fatalf("read ledger ref: %v", err)
	}
	if !ref.Valid || ref.Int64 == 0 {
		t.Errorf("deposit ledger row must reference its history transaction, got %+v", ref)
	}
}
