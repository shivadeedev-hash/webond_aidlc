package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/domain"
)

// Recovery converts panics into a sanitized 500 error envelope and logs the
// stack trace at error level. The stack is logged server-side only and never
// returned to clients (operations.md § Error Handling).
func Recovery(log *slog.Logger, production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Get(requestIDKey)
				attrs := []any{
					slog.Any("request-id", reqID),
					slog.String("path", c.Request.URL.Path),
					slog.Any("panic", r),
				}
				if !production {
					attrs = append(attrs, slog.String("stack", string(debug.Stack())))
				}
				log.Error("panic recovered", attrs...)

				requestID, _ := reqID.(string)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    domain.CodeInternal,
						"message": "An unexpected error occurred",
					},
					"meta": gin.H{"requestId": requestID},
				})
			}
		}()
		c.Next()
	}
}
