package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/auth"
)

func init() { gin.SetMode(gin.TestMode) }

func TestRequestID_GeneratesAndEchoes(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := w.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected a generated X-Request-Id header")
	}
}

func TestRequestID_ReusesInbound(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "fixed-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get(RequestIDHeader); got != "fixed-id" {
		t.Fatalf("X-Request-Id = %q, want fixed-id", got)
	}
}

func authRouter(tokens *auth.Manager, role auth.Role, admin bool) *gin.Engine {
	r := gin.New()
	r.Use(RequestID())
	grp := r.Group("")
	grp.Use(Auth(tokens))
	if admin {
		grp.Use(RequireRole(role))
	}
	grp.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestAuth_MissingHeader(t *testing.T) {
	tokens := auth.NewManager("a", "r")
	r := authRouter(tokens, auth.RoleInvestor, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	tokens := auth.NewManager("a", "r")
	tok, _ := tokens.IssueAccessToken(1, auth.RoleInvestor)
	r := authRouter(tokens, auth.RoleInvestor, false)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	tokens := auth.NewManager("a", "r")
	tok, _ := tokens.IssueAccessToken(1, auth.RoleInvestor) // investor...
	r := authRouter(tokens, auth.RoleAdmin, true)           // ...but admin required

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	tokens := auth.NewManager("a", "r")
	tok, _ := tokens.IssueAccessToken(1, auth.RoleAdmin)
	r := authRouter(tokens, auth.RoleAdmin, true)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestIdempotency(t *testing.T) {
	store := NewMemoryIdempotencyStore(time.Minute)
	r := gin.New()
	r.Use(RequestID())
	grp := r.Group("")
	grp.Use(Idempotency(store))
	grp.POST("/orders", func(c *gin.Context) { c.Status(http.StatusCreated) })

	// Missing key → 400.
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/orders", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing key status = %d, want 400", w.Code)
	}

	// First use → 201.
	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req.Header.Set(IdempotencyHeader, "key-1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("first use status = %d, want 201", w.Code)
	}

	// Duplicate → 409.
	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyHeader, "key-1")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req2)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want 409", w.Code)
	}
}

func TestCORS_Preflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("allowed origin not echoed for preflight")
	}
	// The SPA must be able to read the download file name cross-origin.
	if got := w.Header().Get("Access-Control-Expose-Headers"); !strings.Contains(got, "Content-Disposition") {
		t.Errorf("Access-Control-Expose-Headers = %q, want it to include Content-Disposition", got)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORS([]string{"http://localhost:5173"}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin was echoed: %q", got)
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); got != "" {
		t.Errorf("headers must not be exposed to a disallowed origin: %q", got)
	}
}
