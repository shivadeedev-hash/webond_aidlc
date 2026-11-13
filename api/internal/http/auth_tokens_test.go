package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/config"
)

// TestAuth_ProductionNeverEchoesTokens: in production the verify and reset
// tokens travel only by e-mail; the API responses must not contain them
// (Task 10.8 — echoing them would let anyone activate or take over accounts).
func TestAuth_ProductionNeverEchoesTokens(t *testing.T) {
	srv, _, pool := buildTestServerWithConfig(t, &config.Config{AppEnv: "production", Port: "0"})
	email := "prod-" + decimal.NewFromInt(time.Now().UnixNano()).String() + "@webond.test"

	post := func(path string, body map[string]any, want int) string {
		t.Helper()
		b, _ := json.Marshal(body)
		resp, err := http.Post(srv.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("POST %s: %v", path, err)
		}
		defer resp.Body.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		if resp.StatusCode != want {
			t.Fatalf("POST %s: status %d, want %d (%s)", path, resp.StatusCode, want, buf.String())
		}
		return buf.String()
	}

	reg := post("/api/auth/register", map[string]any{
		"email": email, "password": "secret123", "fullName": "Prod", "consent": true,
	}, http.StatusCreated)
	var parsed struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal([]byte(reg), &parsed)
	if id, ok := parsed.Data["userId"].(float64); ok {
		t.Cleanup(func() { cleanupUser(pool, uint64(id)) })
	}
	if strings.Contains(reg, "verifyCode") {
		t.Fatalf("production register response leaks the verify token: %s", reg)
	}

	// Same response shape whether or not the account exists (no enumeration).
	known := post("/api/auth/forgot-password", map[string]any{"email": email}, http.StatusAccepted)
	unknown := post("/api/auth/forgot-password", map[string]any{"email": "nobody-" + email}, http.StatusAccepted)
	if strings.Contains(known, "resetCode") || strings.Contains(unknown, "resetCode") {
		t.Fatalf("production forgot-password leaks the reset token: %s / %s", known, unknown)
	}
	strip := func(s string) string { // requestId differs per call
		var v map[string]any
		_ = json.Unmarshal([]byte(s), &v)
		delete(v, "meta")
		out, _ := json.Marshal(v)
		return string(out)
	}
	if strip(known) != strip(unknown) {
		t.Fatalf("forgot-password reveals account existence: %s vs %s", known, unknown)
	}
}
