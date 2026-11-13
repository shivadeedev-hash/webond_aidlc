// Package middleware holds the HTTP cross-cutting concerns: request-id,
// structured logging, recovery, auth/RBAC, idempotency, and CORS
// (components.md § HTTP Layer).
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the canonical header used to propagate a correlation id.
const RequestIDHeader = "X-Request-Id"

// requestIDKey is the gin context key under which the request id is stored.
const requestIDKey = "requestId"

// RequestID ensures every request has a correlation id. It reuses an inbound
// X-Request-Id when present, otherwise generates one, stores it in the context
// and echoes it on the response (operations.md § Logging — Correlation).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// RequestIDKey exposes the context key for other packages (e.g. response).
func RequestIDKey() string { return requestIDKey }
