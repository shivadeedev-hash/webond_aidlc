package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/domain"
)

// shortID returns a short unique suffix for asset codes.
func shortID() string { return fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000) }

func TestTrading_BuyFund_AllotsUnitsAndSettles(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "KT"+shortID(), decimal.RequireFromString("10.0000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	res, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	})
	if err != nil {
		t.Fatalf("buy fund: %v", err)
	}
	if res.Status != "completed" || res.Units != "500.0000" { // 5000/10
		t.Fatalf("got status=%s units=%s", res.Status, res.Units)
	}
	b, _ := ws.Get(ctx, userID)
	if !b.Balance.Equal(decimal.RequireFromString("95000.0000")) || !b.Reserved.Equal(decimal.Zero.Round(4)) {
		t.Fatalf("balance=%s reserved=%s", b.Balance, b.Reserved)
	}
}

func TestTrading_BuyFund_InsufficientBalance(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100"), 5)
	fundID := seedFund(t, store, clock, "LB"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	_, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeInsufficientBalance {
		t.Fatalf("expected INSUFFICIENT_BALANCE, got %v", err)
	}
}

func TestTrading_SellFund_InsufficientHolding(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "SF"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	_, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "sell",
		OrderMode: "units", Units: decimal.RequireFromString("100"),
	})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeInsufficientHolding {
		t.Fatalf("expected INSUFFICIENT_HOLDING, got %v", err)
	}
}

func TestTrading_Idempotency(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "ID"+shortID(), decimal.RequireFromString("10"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	key := "idem-" + shortID()
	in := OrderInput{UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"), IdempotencyKey: key}
	r1, err := ts.PlaceOrder(ctx, in)
	if err != nil {
		t.Fatalf("first order: %v", err)
	}
	r2, err := ts.PlaceOrder(ctx, in)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if r1.TxnRef != r2.TxnRef {
		t.Fatalf("idempotent replay returned different order: %s vs %s", r1.TxnRef, r2.TxnRef)
	}
	b, _ := ws.Get(ctx, userID)
	if !b.Balance.Equal(decimal.RequireFromString("95000.0000")) {
		t.Fatalf("balance after idempotent replay=%s want 95000", b.Balance)
	}
}

func TestTrading_RiskAckRequired(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 2)
	fundID := seedFund(t, store, clock, "RA"+shortID(), decimal.RequireFromString("10"), 5)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	base := OrderInput{UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000")}
	_, err := ts.PlaceOrder(ctx, base)
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeRiskAckRequired {
		t.Fatalf("expected RISK_ACK_REQUIRED, got %v", err)
	}
	base.RiskAck = true
	if _, err := ts.PlaceOrder(ctx, base); err != nil {
		t.Fatalf("buy with ack: %v", err)
	}
}

func TestTrading_BondLotRule(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("1000000"), 5)
	bondID := seedBond(t, store, clock, "BD"+shortID(), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "bond", bondID) })

	// Not a multiple of 100,000 → BOND_OFFER_INVALID (Add-on lot rule).
	_, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "bond", AssetID: bondID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("150000"),
	})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeBondOfferInvalid {
		t.Fatalf("expected BOND_OFFER_INVALID, got %v", err)
	}
	// Valid multiple → completed.
	res, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "bond", AssetID: bondID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("200000"),
	})
	if err != nil {
		t.Fatalf("valid bond buy: %v", err)
	}
	if res.Units != "200.0000" { // 200000 / 1000 face value
		t.Fatalf("bond units=%s want 200.0000", res.Units)
	}
}

func TestTrading_GoldQuoteConfirm(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	goldID := seedGold(t, store, clock, "GL"+shortID(), decimal.RequireFromString("41000"), decimal.RequireFromString("41250"))
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "gold", goldID) })

	q, err := ts.Quote(ctx, userID, goldID, "buy", decimal.RequireFromString("10000"), decimal.Zero)
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	res, err := ts.ConfirmQuote(ctx, userID, q.QuoteID, "")
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if res.Status != "completed" {
		t.Fatalf("gold confirm status=%s", res.Status)
	}

	// Expired quote → QUOTE_EXPIRED.
	q2, _ := ts.Quote(ctx, userID, goldID, "buy", decimal.RequireFromString("5000"), decimal.Zero)
	clock.Advance(quoteTTL + time.Second)
	_, err = ts.ConfirmQuote(ctx, userID, q2.QuoteID, "")
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeQuoteExpired {
		t.Fatalf("expected QUOTE_EXPIRED, got %v", err)
	}
}

// TestTrading_SellFund_ByAmountAndByUnits: US-014 allows selling "by units or
// amount". Selling by THB used to fail with "Sell quantity must be positive"
// because only units were read (Add-on 2026-11-12d).
func TestTrading_SellFund_ByAmountAndByUnits(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ws := NewWalletService(store, clock)
	ts := NewTradingService(store, ws, NewNotificationService(store, clock), clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.RequireFromString("100000"), 5)
	fundID := seedFund(t, store, clock, "SA"+shortID(), decimal.RequireFromString("10.0000"), 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID); cleanupAsset(t, pool, "fund", fundID) })

	if _, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "buy",
		OrderMode: "amount", Amount: decimal.RequireFromString("5000"),
	}); err != nil {
		t.Fatalf("buy: %v", err)
	}

	// Sell 1,000 THB at NAV 10 → 100 units.
	res, err := ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "sell",
		OrderMode: "amount", Amount: decimal.RequireFromString("1000"),
	})
	if err != nil {
		t.Fatalf("sell by amount: %v", err)
	}
	if res.Units != "100.0000" || res.Status != "completed" {
		t.Fatalf("sell by amount: units=%s status=%s", res.Units, res.Status)
	}

	// Sell by units: 50 of the remaining 400.
	if res, err = ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "sell",
		OrderMode: "units", Units: decimal.RequireFromString("50"),
	}); err != nil || res.Units != "50.0000" {
		t.Fatalf("sell by units: %+v %v", res, err)
	}

	// More than held: the message says how much can be sold.
	_, err = ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "sell",
		OrderMode: "amount", Amount: decimal.RequireFromString("999999"),
	})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeInsufficientHolding ||
		!strings.Contains(appErr.Message, "350.0000") {
		t.Fatalf("expected INSUFFICIENT_HOLDING with the 350-unit maximum, got %v", err)
	}

	// Zero / empty is a clear validation error, not "quantity must be positive".
	_, err = ts.PlaceOrder(ctx, OrderInput{
		UserID: userID, AssetType: "fund", AssetID: fundID, Side: "sell", OrderMode: "amount",
	})
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeValidationError ||
		!strings.Contains(appErr.Message, "Enter an amount") {
		t.Fatalf("expected a clear validation message, got %v", err)
	}
}
