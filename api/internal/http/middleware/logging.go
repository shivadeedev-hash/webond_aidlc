package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs the start and end of each request with the correlation id,
// method, path, status and duration (operations.md § Per-Component Logging).
// It never logs request bodies or secrets.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID, _ := c.Get(requestIDKey)

		c.Next()

		attrs := []any{
			slog.Any("request-id", reqID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		}
		// Include the authenticated user id when present (set by Auth middleware).
		if uid, ok := c.Get(ContextUserID); ok {
			attrs = append(attrs, slog.Any("user-id", uid))
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			log.Error("request completed", attrs...)
		case status >= 400:
			log.Warn("request completed", attrs...)
		default:
			log.Info("request completed", attrs...)
		}
	}
}
