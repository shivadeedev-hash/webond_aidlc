package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS allows the SPA origin to call the API with credentials (refresh cookie).
// allowedOrigins is an explicit allowlist (secure-by-default: no wildcard when
// credentials are allowed). The Idempotency-Key and Authorization headers are
// permitted (api-spec.md).
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				// Let the SPA read the download file name (cross-origin responses
				// hide every header that is not explicitly exposed).
				c.Header("Access-Control-Expose-Headers", "Content-Disposition, "+RequestIDHeader)
				c.Header("Vary", "Origin")
			}
		}
		c.Header("Access-Control-Allow-Methods", strings.Join([]string{
			http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodDelete, http.MethodOptions,
		}, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join([]string{
			"Authorization", "Content-Type", "Idempotency-Key", RequestIDHeader,
		}, ", "))

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
