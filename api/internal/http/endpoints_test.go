package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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
	"github.com/webond/api/internal/logging"
	"github.com/webond/api/internal/repo"
	"github.com/webond/api/internal/service"
)

// buildTestServer wires the full router against the live DB (skips if absent).
func buildTestServer(t *testing.T) (*httptest.Server, *repo.Store, *sql.DB) {
	t.Helper()
	return buildTestServerWithConfig(t, &config.Config{AppEnv: "development", Port: "0"})
}

// buildTestServerWithConfig is buildTestServer with an explicit config (e.g.
// production behaviour of the auth endpoints).
func buildTestServerWithConfig(t *testing.T, cfg *config.Config) (*httptest.Server, *repo.Store, *sql.DB) {
	t.Helper()
	dsn := testdb.DSN(t)
	pool, err := database.Open(context.Background(), dsn)
	if err != nil {
		t.Skipf("cannot reach test DB (%v); skipping", err)
	}
	logger := logging.New(0, false)
	tokens := auth.NewManager("test-access", "test-refresh")
	store := repo.NewStore(pool)
	clock := service.NewSystemClock()

	notify := service.NewNotificationService(store, clock)
	ws := service.NewWalletService(store, clock)
	ts := service.NewTradingService(store, ws, notify, clock)
	pf := service.NewPortfolioService(store, clock)
	pg := service.NewPriceGeneratorService(store, logger, clock)
	api := NewAPI(cfg, APIServices{
		Auth:    service.NewAuthService(store, tokens, notify, clock),
		Profile: service.NewProfileService(store, clock), Wallet: ws,
		Catalog: service.NewCatalogService(store, clock), Trading: ts,
		DCA:       service.NewDCAService(store, ts, notify, clock),
		Portfolio: pf, Dashboard: service.NewDashboardService(store, pf, ws, clock),
		History: service.NewHistoryService(store, ws, clock), Notify: notify,
		Bond:  service.NewBondLifecycleService(store, ws, notify, clock),
		Admin: service.NewAdminService(store, clock), PriceGen: pg,
	})
	router := NewRouter(RouterDeps{
		Config: cfg, Logger: logger, Tokens: tokens,
		Health: NewHealthHandler(pool), API: api, AllowedOrigins: []string{"*"},
	})
	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close(); _ = pool.Close() })
	return srv, store, pool
}

// TestEndpoints_FullFlow exercises the wired router end-to-end: register →
// verify → login → fund wallet → assess risk → buy fund → read portfolio.
func TestEndpoints_FullFlow(t *testing.T) {
	srv, store, pool := buildTestServer(t)

	email := "ep-" + decimal.NewFromInt(time.Now().UnixNano()).String() + "@webond.test"

	// Register.
	var reg struct {
		Data struct {
			UserID     uint64 `json:"userId"`
			VerifyCode string `json:"verifyCode"`
		} `json:"data"`
	}
	doJSON(t, srv.URL+"/api/auth/register", "", map[string]any{
		"email": email, "password": "secret123", "fullName": "EP", "consent": true,
	}, &reg, http.StatusCreated)
	userID := reg.Data.UserID
	t.Cleanup(func() { cleanupUser(pool, userID) })

	// Verify.
	doJSON(t, srv.URL+"/api/auth/verify", "", map[string]any{"email": email, "code": reg.Data.VerifyCode}, nil, http.StatusOK)

	// Login.
	var login struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	doJSON(t, srv.URL+"/api/auth/login", "", map[string]any{"email": email, "password": "secret123"}, &login, http.StatusOK)
	token := login.Data.AccessToken
	if token == "" {
		t.Fatal("no access token")
	}

	// Deposit + risk assessment.
	doJSON(t, srv.URL+"/api/wallet/deposit", token, map[string]any{"amount": "50000"}, nil, http.StatusOK)
	doJSON(t, srv.URL+"/api/risk-assessment", token, map[string]any{"riskLevel": 5}, nil, http.StatusOK)

	// Seed a fund directly (admin master-data flow is covered elsewhere).
	fundID := seedFundDirect(t, store, "EP"+shortCode())
	t.Cleanup(func() { cleanupFund(pool, fundID) })

	// Place a buy order with an idempotency key.
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/orders", mustJSON(map[string]any{
		"assetType": "fund", "assetId": fundID, "side": "buy", "orderMode": "amount", "amount": "5000",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "ep-"+shortCode())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("order: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("order status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Portfolio reflects the holding.
	var pf struct {
		Data struct {
			Holdings []map[string]any `json:"holdings"`
		} `json:"data"`
	}
	doGet(t, srv.URL+"/api/portfolio", token, &pf, http.StatusOK)
	if len(pf.Data.Holdings) == 0 {
		t.Fatal("expected at least one holding")
	}

	// Public assets list works without auth (guest).
	doGet(t, srv.URL+"/api/assets?type=fund", "", nil, http.StatusOK)
}

func doJSON(t *testing.T, url, token string, body any, out any, want int) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, mustJSON(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("%s status=%d want %d", url, resp.StatusCode, want)
	}
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
}

func doGet(t *testing.T, url, token string, out any, want int) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("%s status=%d want %d", url, resp.StatusCode, want)
	}
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
}

func mustJSON(v any) *bytes.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func shortCode() string { return decimal.NewFromInt(time.Now().UnixNano() % 1_000_000).String() }

func seedFundDirect(t *testing.T, store *repo.Store, code string) uint64 {
	t.Helper()
	ctx := context.Background()
	res, err := store.Funds.Create(ctx, db.CreateFundParams{
		Code: code, Name: code, MinPurchase: decimal.RequireFromString("500"),
		SettlementDays: 3, CutoffTime: "15:30:00", IsActive: true,
		RiskLevel: sql.NullInt16{Int16: 3, Valid: true},
	})
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("seed fund: %v", err)
	}
	fundID := uint64(id)
	if err := store.Funds.UpsertNav(ctx, db.UpsertFundNavParams{
		FundID: fundID, NavDate: time.Now(), Nav: decimal.RequireFromString("10"),
	}); err != nil {
		t.Fatalf("seed nav: %v", err)
	}
	return fundID
}

func cleanupUser(pool *sql.DB, userID uint64) {
	ctx := context.Background()
	for _, q := range []string{
		"DELETE FROM transaction_status_history WHERE transaction_id IN (SELECT id FROM transactions WHERE user_id = ?)",
		"DELETE FROM transactions WHERE user_id = ?",
		"DELETE FROM orders WHERE user_id = ?",
		"DELETE FROM holdings WHERE user_id = ?",
		"DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id = ?)",
		"DELETE FROM wallets WHERE user_id = ?",
		"DELETE FROM risk_assessments WHERE user_id = ?",
		"DELETE FROM notification_outbox WHERE user_id = ?",
		"DELETE FROM audit_logs WHERE user_id = ?",
		"DELETE FROM users WHERE id = ?",
	} {
		_, _ = pool.ExecContext(ctx, q, userID)
	}
}

func cleanupFund(pool *sql.DB, fundID uint64) {
	ctx := context.Background()
	_, _ = pool.ExecContext(ctx, "DELETE FROM fund_navs WHERE fund_id = ?", fundID)
	_, _ = pool.ExecContext(ctx, "DELETE FROM funds WHERE id = ?", fundID)
}
