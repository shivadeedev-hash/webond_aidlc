//go:build integration

// Package integration exercises the WeBond services + sqlc repos against a real
// MySQL database through the fully wired Gin router (testing-strategy.md §
// Backend integration; integration.md § Integration Testing).
//
// Run: cd api && go test -tags=integration ./test/integration/...
//
// DSN-gated: the suite reads WEBOND_TEST_DSN (preferred) or DATABASE_URL and
// skips entirely when neither is set, so a plain `go test ./...` without the
// build tag never requires a database. Each test creates its own fixtures with
// unique keys and cleans them up, so the suite is safe to run against the dev
// DB alongside the demo seed.
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/webond/api/internal/testdb"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/config"
	"github.com/webond/api/internal/database"
	"github.com/webond/api/internal/db"
	apihttp "github.com/webond/api/internal/http"
	"github.com/webond/api/internal/logging"
	"github.com/webond/api/internal/repo"
	"github.com/webond/api/internal/service"
)

// ---------------------------------------------------------------------------
// Harness
// ---------------------------------------------------------------------------

type harness struct {
	srv   *httptest.Server
	store *repo.Store
	pool  *sql.DB
}

func setup(t *testing.T) *harness {
	t.Helper()
	dsn := testdb.DSN(t)
	pool, err := database.Open(context.Background(), dsn)
	if err != nil {
		t.Skipf("cannot reach test DB (%v); skipping", err)
	}
	cfg := &config.Config{AppEnv: "development", Port: "0"}
	logger := logging.New(0, false)
	tokens := auth.NewManager("it-access", "it-refresh")
	store := repo.NewStore(pool)
	clock := service.NewSystemClock()

	notify := service.NewNotificationService(store, clock)
	ws := service.NewWalletService(store, clock)
	ts := service.NewTradingService(store, ws, notify, clock)
	pf := service.NewPortfolioService(store, clock)
	pg := service.NewPriceGeneratorService(store, logger, clock)
	api := apihttp.NewAPI(cfg, apihttp.APIServices{
		Auth:    service.NewAuthService(store, tokens, notify, clock),
		Profile: service.NewProfileService(store, clock), Wallet: ws,
		Catalog: service.NewCatalogService(store, clock), Trading: ts,
		DCA:       service.NewDCAService(store, ts, notify, clock),
		Portfolio: pf, Dashboard: service.NewDashboardService(store, pf, ws, clock),
		History: service.NewHistoryService(store, ws, clock), Notify: notify,
		Bond:  service.NewBondLifecycleService(store, ws, notify, clock),
		Admin: service.NewAdminService(store, clock), PriceGen: pg,
	})
	router := apihttp.NewRouter(apihttp.RouterDeps{
		Config: cfg, Logger: logger, Tokens: tokens,
		Health: apihttp.NewHealthHandler(pool), API: api, AllowedOrigins: []string{"*"},
	})
	srv := httptest.NewServer(router)
	h := &harness{srv: srv, store: store, pool: pool}
	t.Cleanup(func() { srv.Close(); _ = pool.Close() })
	return h
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

type reqOpts struct {
	token   string
	headers map[string]string
}

func (h *harness) do(t *testing.T, method, path, token string, body any, opts *reqOpts) (int, []byte) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, h.srv.URL+path, r)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if opts != nil {
		for k, v := range opts.headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func decodeData(t *testing.T, raw []byte, out any) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode envelope: %v (%s)", err, string(raw))
	}
	if out != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			t.Fatalf("decode data: %v (%s)", err, string(env.Data))
		}
	}
}

func errorCode(t *testing.T, raw []byte) string {
	t.Helper()
	var env struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &env)
	return env.Error.Code
}

// uniq returns a per-call unique suffix (keeps emails/codes collision-free).
func uniq() string { return fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000) }

// ---------------------------------------------------------------------------
// Fixture builders (direct store writes — assets are shared master data)
// ---------------------------------------------------------------------------

// activeUser registers → verifies → logs in and returns a usable access token
// plus the user id. The wallet is created by Register.
func (h *harness) activeUser(t *testing.T) (string, uint64) {
	t.Helper()
	email := "it-" + uniq() + "@webond.test"
	password := "Secret@123"

	var reg struct {
		UserID     uint64 `json:"userId"`
		VerifyCode string `json:"verifyCode"`
	}
	st, raw := h.do(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"email": email, "password": password, "fullName": "IT User", "consent": true,
	}, nil)
	if st != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &reg)
	t.Cleanup(func() { h.cleanupUser(reg.UserID) })

	st, _ = h.do(t, http.MethodPost, "/api/auth/verify", "", map[string]any{"email": email, "code": reg.VerifyCode}, nil)
	if st != http.StatusOK {
		t.Fatalf("verify status=%d", st)
	}

	var login struct {
		AccessToken string `json:"accessToken"`
	}
	st, raw = h.do(t, http.MethodPost, "/api/auth/login", "", map[string]any{"email": email, "password": password}, nil)
	if st != http.StatusOK {
		t.Fatalf("login status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &login)
	if login.AccessToken == "" {
		t.Fatal("empty access token")
	}
	return login.AccessToken, reg.UserID
}

// tradeableUser is an active user with cash + a valid risk assessment.
func (h *harness) tradeableUser(t *testing.T, deposit string) (string, uint64) {
	t.Helper()
	token, uid := h.activeUser(t)
	if st, raw := h.do(t, http.MethodPost, "/api/wallet/deposit", token, map[string]any{"amount": deposit}, nil); st != http.StatusOK {
		t.Fatalf("deposit status=%d body=%s", st, raw)
	}
	if st, _ := h.do(t, http.MethodPost, "/api/risk-assessment", token, map[string]any{"riskLevel": 5}, nil); st != http.StatusOK {
		t.Fatalf("risk status=%d", st)
	}
	return token, uid
}

// adminToken promotes a fresh active user to admin (direct status write) and
// returns an admin access token. Used to hit /api/admin/* (RBAC) endpoints.
func (h *harness) adminToken(t *testing.T) string {
	t.Helper()
	_, uid := h.activeUser(t)
	if _, err := h.pool.Exec("UPDATE users SET role='admin' WHERE id=?", uid); err != nil {
		t.Fatalf("promote admin: %v", err)
	}
	// Re-login so the JWT carries the admin role claim.
	email := h.emailOf(t, uid)
	var login struct {
		AccessToken string `json:"accessToken"`
	}
	st, raw := h.do(t, http.MethodPost, "/api/auth/login", "", map[string]any{"email": email, "password": "Secret@123"}, nil)
	if st != http.StatusOK {
		t.Fatalf("admin re-login status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &login)
	return login.AccessToken
}

func (h *harness) emailOf(t *testing.T, uid uint64) string {
	t.Helper()
	var email string
	if err := h.pool.QueryRow("SELECT email FROM users WHERE id=?", uid).Scan(&email); err != nil {
		t.Fatalf("email lookup: %v", err)
	}
	return email
}

func (h *harness) seedFund(t *testing.T, risk int16, nav string) uint64 {
	t.Helper()
	ctx := context.Background()
	code := "ITF" + uniq()
	res, err := h.store.Funds.Create(ctx, db.CreateFundParams{
		Code: code, Name: code, MinPurchase: decimal.RequireFromString("500"),
		SettlementDays: 3, CutoffTime: "15:30:00", IsActive: true,
		RiskLevel: sql.NullInt16{Int16: risk, Valid: true},
	})
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed fund: %v", err)
	}
	fundID := uint64(id)
	if err := h.store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
		FundID: fundID, NavDate: time.Now(), Nav: decimal.RequireFromString(nav),
	}); err != nil {
		t.Fatalf("seed nav: %v", err)
	}
	t.Cleanup(func() {
		h.exec("DELETE FROM fund_navs WHERE fund_id=?", fundID)
		h.exec("DELETE FROM funds WHERE id=?", fundID)
	})
	return fundID
}

func (h *harness) seedGold(t *testing.T, bid, ask string) uint64 {
	t.Helper()
	ctx := context.Background()
	code := "ITG" + uniq()
	res, err := h.store.Gold.Create(ctx, db.CreateGoldProductParams{Code: code, Name: code, Purity: sql.NullString{String: "96.5%", Valid: true}, IsActive: true})
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed gold: %v", err)
	}
	goldID := uint64(id)
	// Fresh tick (now) so quotes are not stale (EX-10).
	if _, err := h.store.Gold.CreatePrice(ctx, db.CreateGoldPriceParams{
		GoldProductID: goldID, PricedAt: time.Now(),
		BidPerGram: decimal.RequireFromString(bid), AskPerGram: decimal.RequireFromString(ask),
	}); err != nil {
		t.Fatalf("seed gold price: %v", err)
	}
	t.Cleanup(func() {
		h.exec("DELETE FROM gold_prices WHERE gold_product_id=?", goldID)
		h.exec("DELETE FROM gold_products WHERE id=?", goldID)
	})
	return goldID
}

func (h *harness) exec(q string, args ...any) { _, _ = h.pool.Exec(q, args...) }

func (h *harness) cleanupUser(userID uint64) {
	for _, q := range []string{
		"DELETE FROM dca_rounds WHERE dca_plan_id IN (SELECT id FROM dca_plans WHERE user_id=?)",
		"DELETE FROM dca_plans WHERE user_id=?",
		"DELETE FROM transaction_status_history WHERE transaction_id IN (SELECT id FROM transactions WHERE user_id=?)",
		"DELETE FROM transactions WHERE user_id=?",
		"DELETE FROM orders WHERE user_id=?",
		"DELETE FROM holdings WHERE user_id=?",
		"DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id=?)",
		"DELETE FROM wallets WHERE user_id=?",
		"DELETE FROM risk_assessments WHERE user_id=?",
		"DELETE FROM notification_outbox WHERE user_id=?",
		"DELETE FROM audit_logs WHERE user_id=?",
		"DELETE FROM users WHERE id=?",
	} {
		h.exec(q, userID)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Auth: register → verify → login → lockout after 5 bad passwords (EX-01).
func TestAuth_RegisterVerifyLogin_AndLockout(t *testing.T) {
	h := setup(t)
	email := "it-" + uniq() + "@webond.test"

	var reg struct {
		UserID     uint64 `json:"userId"`
		VerifyCode string `json:"verifyCode"`
	}
	st, raw := h.do(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"email": email, "password": "Secret@123", "fullName": "Lock User", "consent": true,
	}, nil)
	if st != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &reg)
	t.Cleanup(func() { h.cleanupUser(reg.UserID) })

	// Duplicate email → 409 DUPLICATE_EMAIL.
	if st, raw = h.do(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"email": email, "password": "Secret@123", "fullName": "Dup", "consent": true,
	}, nil); st != http.StatusConflict || errorCode(t, raw) != "DUPLICATE_EMAIL" {
		t.Fatalf("expected 409 DUPLICATE_EMAIL, got %d %s", st, raw)
	}

	// Verify then login succeeds.
	if st, _ = h.do(t, http.MethodPost, "/api/auth/verify", "", map[string]any{"email": email, "code": reg.VerifyCode}, nil); st != http.StatusOK {
		t.Fatalf("verify status=%d", st)
	}
	if st, raw = h.do(t, http.MethodPost, "/api/auth/login", "", map[string]any{"email": email, "password": "Secret@123"}, nil); st != http.StatusOK {
		t.Fatalf("login status=%d body=%s", st, raw)
	}

	// 5 bad attempts → account locked (EX-01). All return 401, but after the
	// threshold the DB status flips to 'locked'.
	for i := 0; i < 5; i++ {
		h.do(t, http.MethodPost, "/api/auth/login", "", map[string]any{"email": email, "password": "wrong-pass"}, nil)
	}
	var status string
	if err := h.pool.QueryRow("SELECT status FROM users WHERE id=?", reg.UserID).Scan(&status); err != nil {
		t.Fatalf("status lookup: %v", err)
	}
	if status != "locked" {
		t.Fatalf("expected account locked after 5 fails, got status=%q", status)
	}
	// Even the correct password is rejected while locked.
	if st, _ = h.do(t, http.MethodPost, "/api/auth/login", "", map[string]any{"email": email, "password": "Secret@123"}, nil); st != http.StatusUnauthorized {
		t.Fatalf("expected 401 while locked, got %d", st)
	}
}

// Orders: MF buy + idempotency replay; then sell part of the position.
func TestOrders_FundBuyIdempotentThenSell(t *testing.T) {
	h := setup(t)
	token, _ := h.tradeableUser(t, "100000")
	fundID := h.seedFund(t, 3, "10.0000")

	key := "it-" + uniq()
	buy := map[string]any{"assetType": "fund", "assetId": fundID, "side": "buy", "orderMode": "amount", "amount": "5000"}

	st, raw := h.do(t, http.MethodPost, "/api/orders", token, buy, &reqOpts{headers: map[string]string{"Idempotency-Key": key}})
	if st != http.StatusCreated {
		t.Fatalf("buy status=%d body=%s", st, raw)
	}
	var order1 struct {
		TxnRef string `json:"txnRef"`
		Units  string `json:"units"`
	}
	decodeData(t, raw, &order1)
	// The order can be fetched back by its reference (GET /api/orders/{ref}).
	if st, raw := h.do(t, http.MethodGet, "/api/orders/"+order1.TxnRef, token, nil, nil); st != http.StatusOK {
		t.Fatalf("get order by ref: status=%d body=%s", st, raw)
	} else {
		var got struct {
			TxnRef string `json:"txnRef"`
		}
		decodeData(t, raw, &got)
		if got.TxnRef != order1.TxnRef {
			t.Fatalf("get order returned %q, want %q", got.TxnRef, order1.TxnRef)
		}
	}

	// Replay with the SAME key. The in-process idempotency middleware is the
	// fast guard: a duplicate key is rejected with 409 (the DB unique constraint
	// on orders.idempotency_key is the durable guarantee). Either way, no second
	// order/holding is created.
	st, raw = h.do(t, http.MethodPost, "/api/orders", token, buy, &reqOpts{headers: map[string]string{"Idempotency-Key": key}})
	if st == http.StatusCreated || st == http.StatusOK {
		// If the middleware entry had expired, the service replay must return the
		// SAME order (same txnRef) — never a new one.
		var order2 struct {
			TxnRef string `json:"txnRef"`
		}
		decodeData(t, raw, &order2)
		if order2.TxnRef != order1.TxnRef {
			t.Fatalf("idempotency broken: replay produced a new order %q != %q", order2.TxnRef, order1.TxnRef)
		}
	} else if st != http.StatusConflict {
		t.Fatalf("idempotent replay: expected 409 (duplicate key) or 200/201 (same order), got %d body=%s", st, raw)
	}

	// Idempotency holds: portfolio shows exactly one holding regardless of the
	// replay path (one buy of 500 units at NAV 10).
	var pf struct {
		Holdings []map[string]any `json:"holdings"`
	}
	_, raw = h.do(t, http.MethodGet, "/api/portfolio", token, nil, nil)
	decodeData(t, raw, &pf)
	if len(pf.Holdings) != 1 {
		t.Fatalf("expected 1 holding after idempotent buys, got %d", len(pf.Holdings))
	}

	// Sell 200 units.
	st, raw = h.do(t, http.MethodPost, "/api/orders", token, map[string]any{
		"assetType": "fund", "assetId": fundID, "side": "sell", "orderMode": "units", "units": "200",
	}, &reqOpts{headers: map[string]string{"Idempotency-Key": "it-" + uniq()}})
	if st != http.StatusCreated {
		t.Fatalf("sell status=%d body=%s", st, raw)
	}
}

// Orders: gold quote → confirm (US-017) completes a buy.
func TestOrders_GoldQuoteConfirm(t *testing.T) {
	h := setup(t)
	token, _ := h.tradeableUser(t, "100000")
	goldID := h.seedGold(t, "41000.0000", "41200.0000")

	var q struct {
		QuoteID string `json:"quoteId"`
		Total   string `json:"total"`
	}
	st, raw := h.do(t, http.MethodPost, "/api/orders/quote", token, map[string]any{
		"assetType": "gold", "assetId": goldID, "side": "buy", "amount": "10000",
	}, nil)
	if st != http.StatusOK {
		t.Fatalf("quote status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &q)
	if q.QuoteID == "" {
		t.Fatal("empty quoteId")
	}

	st, raw = h.do(t, http.MethodPost, "/api/orders/"+q.QuoteID+"/confirm", token, nil,
		&reqOpts{headers: map[string]string{"Idempotency-Key": "it-" + uniq()}})
	if st != http.StatusCreated {
		t.Fatalf("confirm status=%d body=%s", st, raw)
	}
	var order struct {
		Status string `json:"status"`
	}
	decodeData(t, raw, &order)
	if order.Status != "completed" {
		t.Fatalf("expected completed gold order, got %q", order.Status)
	}

	// A stale/unknown quote id → 409 QUOTE_EXPIRED.
	st, raw = h.do(t, http.MethodPost, "/api/orders/q_nonexistent/confirm", token, nil,
		&reqOpts{headers: map[string]string{"Idempotency-Key": "it-" + uniq()}})
	if st != http.StatusConflict || errorCode(t, raw) != "QUOTE_EXPIRED" {
		t.Fatalf("expected 409 QUOTE_EXPIRED, got %d %s", st, raw)
	}
}

// Wallet: deposit/withdraw with conservation; over-withdraw → INSUFFICIENT_BALANCE.
func TestWallet_DepositWithdrawConservation(t *testing.T) {
	h := setup(t)
	token, _ := h.activeUser(t)

	type bal struct {
		Balance, Reserved, Available string
	}
	var b bal
	if st, raw := h.do(t, http.MethodPost, "/api/wallet/deposit", token, map[string]any{"amount": "10000"}, nil); st == http.StatusOK {
		decodeData(t, raw, &b)
	} else {
		t.Fatalf("deposit status=%d body=%s", st, raw)
	}
	if b.Balance != "10000.00" || b.Available != "10000.00" {
		t.Fatalf("after deposit: balance=%s available=%s", b.Balance, b.Available)
	}

	if st, raw := h.do(t, http.MethodPost, "/api/wallet/withdraw", token, map[string]any{"amount": "4000"}, nil); st == http.StatusOK {
		decodeData(t, raw, &b)
	} else {
		t.Fatalf("withdraw status=%d body=%s", st, raw)
	}
	if b.Balance != "6000.00" {
		t.Fatalf("after withdraw: balance=%s (want 6000.00)", b.Balance)
	}

	// Over-withdraw → 422 INSUFFICIENT_BALANCE, balance unchanged.
	st, raw := h.do(t, http.MethodPost, "/api/wallet/withdraw", token, map[string]any{"amount": "999999"}, nil)
	if st != http.StatusUnprocessableEntity || errorCode(t, raw) != "INSUFFICIENT_BALANCE" {
		t.Fatalf("expected 422 INSUFFICIENT_BALANCE, got %d %s", st, raw)
	}

	// A buy reserves/settles cash: available must drop by the buy amount.
	fundID := h.seedFund(t, 1, "10.0000")
	if st, _ := h.do(t, http.MethodPost, "/api/risk-assessment", token, map[string]any{"riskLevel": 5}, nil); st != http.StatusOK {
		t.Fatalf("risk status=%d", st)
	}
	if st, raw := h.do(t, http.MethodPost, "/api/orders", token, map[string]any{
		"assetType": "fund", "assetId": fundID, "side": "buy", "orderMode": "amount", "amount": "1000",
	}, &reqOpts{headers: map[string]string{"Idempotency-Key": "it-" + uniq()}}); st != http.StatusCreated {
		t.Fatalf("buy status=%d body=%s", st, raw)
	}
	_, raw = h.do(t, http.MethodGet, "/api/wallet", token, nil, nil)
	decodeData(t, raw, &b)
	if b.Balance != "5000.00" {
		t.Fatalf("after buy: balance=%s (want 5000.00 = 6000 - 1000)", b.Balance)
	}
}

// DCA: create a gold plan due today → manual run creates a round + transaction;
// a second run is idempotent (no duplicate round for the same date, FR-15).
func TestDCA_CreateRunIdempotent(t *testing.T) {
	h := setup(t)
	token, _ := h.tradeableUser(t, "100000")
	adminTok := h.adminToken(t)
	goldID := h.seedGold(t, "41000.0000", "41200.0000")

	// run_day = today's weekday (1=Mon..7=Sun) and start_date today so the plan
	// is due immediately.
	wd := int(time.Now().In(service.Bangkok).Weekday())
	if wd == 0 {
		wd = 7
	}
	today := time.Now().In(service.Bangkok).Format("2006-01-02")
	var plan struct {
		ID uint64 `json:"id"`
	}
	st, raw := h.do(t, http.MethodPost, "/api/dca-plans", token, map[string]any{
		"assetType": "gold", "assetId": goldID, "amountPerRound": "2000",
		"frequency": "weekly", "runDay": wd, "startDate": today,
	}, nil)
	if st != http.StatusCreated {
		t.Fatalf("create dca status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &plan)

	// Manual run (admin).
	if st, raw := h.do(t, http.MethodPost, "/api/admin/run/dca", adminTok, nil, nil); st != http.StatusOK {
		t.Fatalf("run dca status=%d body=%s", st, raw)
	}
	// A second run must not create a duplicate round for the same date.
	if st, _ := h.do(t, http.MethodPost, "/api/admin/run/dca", adminTok, nil, nil); st != http.StatusOK {
		t.Fatalf("second run dca status=%d", st)
	}

	var rounds []map[string]any
	_, raw = h.do(t, http.MethodGet, fmt.Sprintf("/api/dca-plans/%d/rounds", plan.ID), token, nil, nil)
	decodeData(t, raw, &rounds)
	runsToday := 0
	for _, r := range rounds {
		if r["runDate"] == today {
			runsToday++
		}
	}
	if runsToday != 1 {
		t.Fatalf("expected exactly 1 DCA round for today (idempotent), got %d (all=%v)", runsToday, rounds)
	}
}

// History: list/filter + cancel a pending order + xlsx export stream (US-024/27/28).
func TestHistory_ListCancelExport(t *testing.T) {
	h := setup(t)
	token, _ := h.tradeableUser(t, "100000")
	fundID := h.seedFund(t, 1, "10.0000")

	// Create a completed buy so history has at least one row.
	if st, raw := h.do(t, http.MethodPost, "/api/orders", token, map[string]any{
		"assetType": "fund", "assetId": fundID, "side": "buy", "orderMode": "amount", "amount": "1000",
	}, &reqOpts{headers: map[string]string{"Idempotency-Key": "it-" + uniq()}}); st != http.StatusCreated {
		t.Fatalf("buy status=%d body=%s", st, raw)
	}

	// List transactions (newest first).
	var txns []map[string]any
	st, raw := h.do(t, http.MethodGet, "/api/transactions?tab=all&page=1&pageSize=20", token, nil, nil)
	if st != http.StatusOK {
		t.Fatalf("list txns status=%d body=%s", st, raw)
	}
	decodeData(t, raw, &txns)
	if len(txns) == 0 {
		t.Fatal("expected at least one transaction in history")
	}

	// Filter by type=buy returns only buys.
	st, raw = h.do(t, http.MethodGet, "/api/transactions?type=buy", token, nil, nil)
	if st != http.StatusOK {
		t.Fatalf("filter status=%d", st)
	}
	var buys []map[string]any
	decodeData(t, raw, &buys)
	for _, tx := range buys {
		if tx["type"] != "buy" {
			t.Fatalf("type filter leaked a %v row", tx["type"])
		}
	}

	// Insert a cancellable pending MF buy (order + linked txn + reserved cash),
	// then cancel it (US-027, BR-13). Cancel releases the reserved cash.
	ownerID := h.userIDForToken(t, token)
	ref := h.insertPendingBuy(t, ownerID, fundID, "5000.0000")
	st, raw = h.do(t, http.MethodPost, "/api/transactions/"+ref+"/cancel", token, nil, nil)
	if st != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s (code=%s)", st, raw, errorCode(t, raw))
	}
	// Re-cancelling the now-cancelled txn → 409 CANCEL_NOT_ALLOWED.
	if st, raw = h.do(t, http.MethodPost, "/api/transactions/"+ref+"/cancel", token, nil, nil); st != http.StatusConflict || errorCode(t, raw) != "CANCEL_NOT_ALLOWED" {
		t.Fatalf("expected 409 CANCEL_NOT_ALLOWED on re-cancel, got %d %s", st, raw)
	}

	// Export streams an .xlsx (binary) with a 200.
	code, body := h.do(t, http.MethodGet, "/api/transactions/export?tab=all", token, nil, nil)
	if code != http.StatusOK {
		t.Fatalf("export status=%d body=%s", code, body)
	}
	// xlsx files are zip archives → start with "PK".
	if len(body) < 2 || body[0] != 'P' || body[1] != 'K' {
		prefixLen := len(body)
		if prefixLen > 8 {
			prefixLen = 8
		}
		t.Fatalf("export did not return an xlsx (zip) stream; got %d bytes prefix=%q", len(body), string(body[:prefixLen]))
	}
}

// userIDForToken reads the authenticated user's id by calling /api/profile.
func (h *harness) userIDForToken(t *testing.T, token string) uint64 {
	t.Helper()
	var p struct {
		ID uint64 `json:"id"`
	}
	_, raw := h.do(t, http.MethodGet, "/api/profile", token, nil, nil)
	decodeData(t, raw, &p)
	if p.ID == 0 {
		t.Fatal("could not resolve user id from profile")
	}
	return p.ID
}

// insertPendingBuy creates a genuinely cancellable pending MF buy: it reserves
// the cash on the wallet (so the cancel's release honours the reserved ≥ amount
// invariant), writes a pending order, and a submitted transaction linked to it.
// The phase-1 buy path auto-completes, so a pending order is constructed here to
// exercise the cancel flow (BR-13). Returns the transaction txn_ref.
func (h *harness) insertPendingBuy(t *testing.T, userID, fundID uint64, amount string) string {
	t.Helper()
	ref := "ITC" + uniq()
	if len(ref) > 20 {
		ref = ref[:20]
	}

	// Reserve the cash: move `amount` from available into reserved.
	if _, err := h.pool.Exec(
		"UPDATE wallets SET reserved = reserved + ? WHERE user_id = ?", amount, userID); err != nil {
		t.Fatalf("reserve cash: %v", err)
	}

	// Pending order (side buy, amount set) linked by txn_ref.
	res, err := h.pool.Exec(`INSERT INTO orders
		(txn_ref, user_id, asset_type, asset_id, side, order_mode, amount, net_amount, status, trade_date, settlement_date)
		VALUES (?, ?, 'fund', ?, 'buy', 'amount', ?, ?, 'pending', CURDATE(), CURDATE() + INTERVAL 3 DAY)`,
		ref, userID, fundID, amount, amount)
	if err != nil {
		t.Fatalf("insert pending order: %v", err)
	}
	orderID, _ := res.LastInsertId()

	if _, err := h.pool.Exec(`INSERT INTO transactions
		(txn_ref, user_id, type, asset_type, asset_id, amount, net_amount, status, order_id, occurred_at, trade_date, settlement_date)
		VALUES (?, ?, 'buy', 'fund', ?, ?, ?, 'submitted', ?, NOW(), CURDATE(), CURDATE() + INTERVAL 3 DAY)`,
		ref, userID, fundID, "-"+amount, "-"+amount, orderID); err != nil {
		t.Fatalf("insert pending txn: %v", err)
	}
	return ref
}
