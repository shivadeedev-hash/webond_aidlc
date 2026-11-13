package http

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger is the minimal readiness dependency: anything that can ping the DB.
// *sql.DB satisfies this via PingContext.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// readinessTimeout bounds the DB ping (operations.md § Readiness checks: 2s).
const readinessTimeout = 2 * time.Second

// HealthHandler serves liveness and readiness. A "draining" flag flips
// readiness to 503 during graceful shutdown (operations.md § Graceful Shutdown).
type HealthHandler struct {
	db        Pinger
	startedAt time.Time
	draining  atomic.Bool
}

// NewHealthHandler builds a health handler bound to the DB pinger. db may be
// nil (e.g. in unit tests); readiness then reports the DB as unconfigured.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db, startedAt: time.Now()}
}

// StartDraining marks the service as draining so readiness returns 503.
func (h *HealthHandler) StartDraining() { h.draining.Store(true) }

// Live is the liveness probe: 200 as long as the process is up.
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"uptime": int64(time.Since(h.startedAt).Seconds()),
	})
}

// Ready is the readiness probe: 200 only when not draining and the DB ping
// succeeds; otherwise 503 (operations.md § Health & Readiness).
func (h *HealthHandler) Ready(c *gin.Context) {
	if h.draining.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "draining",
			"checks": gin.H{"database": gin.H{"status": "skipped"}},
		})
		return
	}

	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"database": gin.H{"status": "unconfigured"}},
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
	defer cancel()

	start := time.Now()
	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"database": gin.H{"status": "error"}},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": gin.H{
				"status":     "ok",
				"latency_ms": time.Since(start).Milliseconds(),
			},
		},
	})
}
