package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/domain"
)

// Context keys for values the auth middleware stores for downstream handlers.
const (
	ContextUserID = "userID"
	ContextRole   = "role"
)

// errorEnvelope mirrors the response package's shape so middleware can abort
// with a correct error body without importing the http package (avoids an
// import cycle: http imports middleware).
func abortError(c *gin.Context, code domain.ErrorCode, message string) {
	reqID, _ := c.Get(requestIDKey)
	requestID, _ := reqID.(string)
	c.AbortWithStatusJSON(code.HTTPStatus(), gin.H{
		"error": gin.H{"code": code, "message": message},
		"meta":  gin.H{"requestId": requestID},
	})
}

// Auth validates the bearer access token and stores the user id and role in the
// context. Missing/invalid tokens abort with 401 UNAUTHORIZED (api-spec.md).
func Auth(tokens *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortError(c, domain.CodeUnauthorized, "Missing or malformed authorization header")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))

		claims, err := tokens.VerifyAccessToken(raw)
		if err != nil {
			abortError(c, domain.CodeUnauthorized, "Invalid or expired token")
			return
		}
		userID, err := claims.UserID()
		if err != nil {
			abortError(c, domain.CodeUnauthorized, "Invalid token subject")
			return
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// RequireRole enforces RBAC: the authenticated user must hold the given role,
// else 403 FORBIDDEN. Use after Auth (api-spec.md § RBAC — admin routes).
func RequireRole(role auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		got, ok := c.Get(ContextRole)
		if !ok {
			abortError(c, domain.CodeUnauthorized, "Authentication required")
			return
		}
		if got != role {
			abortError(c, domain.CodeForbidden, "Insufficient permissions")
			return
		}
		c.Next()
	}
}
