package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakePinger lets us control readiness outcomes.
type fakePinger struct{ err error }

func (f fakePinger) PingContext(ctx context.Context) error { return f.err }

func newTestEngine(h *HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", h.Live)
	r.GET("/health/ready", h.Ready)
	return r
}

func TestHealth_Live(t *testing.T) {
	h := NewHealthHandler(fakePinger{})
	r := newTestEngine(h)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("live status = %d, want 200", w.Code)
	}
}

func TestHealth_Ready_OK(t *testing.T) {
	h := NewHealthHandler(fakePinger{err: nil})
	r := newTestEngine(h)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", w.Code)
	}
}

func TestHealth_Ready_DBDown(t *testing.T) {
	h := NewHealthHandler(fakePinger{err: errors.New("down")})
	r := newTestEngine(h)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready status = %d, want 503", w.Code)
	}
}

func TestHealth_Ready_Draining(t *testing.T) {
	h := NewHealthHandler(fakePinger{err: nil})
	h.StartDraining()
	r := newTestEngine(h)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("draining ready status = %d, want 503", w.Code)
	}
}
