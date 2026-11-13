// Package http contains the Gin router, middleware, and the API response
// envelope. The envelope shape mirrors api-spec.md § Conventions:
//
//	success: { "data": <payload>, "meta": {...}? }
//	error:   { "error": { "code", "message", "details"? }, "meta": { "requestId" } }
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/http/middleware"
)

// RequestIDFromContext returns the correlation id set by the RequestID
// middleware, or "" when absent.
func RequestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(middleware.RequestIDKey()); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// Meta is the optional envelope metadata. RequestID is always set on error
// responses; pagination fields are set on list responses.
type Meta struct {
	RequestID string `json:"requestId,omitempty"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	Total     int64  `json:"total,omitempty"`
	HasMore   *bool  `json:"hasMore,omitempty"`
}

// successEnvelope wraps a successful payload.
type successEnvelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// errorBody is the error detail block.
type errorBody struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Details map[string]any   `json:"details,omitempty"`
}

// errorEnvelope wraps an error, always carrying the request id in meta.
type errorEnvelope struct {
	Error errorBody `json:"error"`
	Meta  Meta      `json:"meta"`
}

// RespondData writes a 200 success envelope with the given payload.
func RespondData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successEnvelope{Data: data})
}

// RespondStatus writes a success envelope with a specific status (e.g. 201).
func RespondStatus(c *gin.Context, status int, data any) {
	c.JSON(status, successEnvelope{Data: data})
}

// RespondList writes a 200 success envelope with pagination meta.
func RespondList(c *gin.Context, data any, page, pageSize int, total int64) {
	hasMore := int64(page*pageSize) < total
	c.JSON(http.StatusOK, successEnvelope{
		Data: data,
		Meta: &Meta{Page: page, PageSize: pageSize, Total: total, HasMore: &hasMore},
	})
}

// RespondError renders any error into the error envelope with the mapped HTTP
// status. Non-AppErrors become a generic 500 so internal details never leak.
func RespondError(c *gin.Context, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		appErr = domain.NewError(domain.CodeInternal, "An unexpected error occurred")
	}

	c.JSON(appErr.Code.HTTPStatus(), errorEnvelope{
		Error: errorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Meta: Meta{RequestID: RequestIDFromContext(c)},
	})
}

// RespondErrorCode is a convenience that builds and renders an AppError.
func RespondErrorCode(c *gin.Context, code domain.ErrorCode, message string) {
	RespondError(c, domain.NewError(code, message))
}
