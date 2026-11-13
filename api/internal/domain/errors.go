// Package domain holds core entities and the typed API error codes shared
// across services and the HTTP layer. Error codes and their HTTP status
// mappings mirror api-spec.md § Error codes exactly.
package domain

import "net/http"

// ErrorCode is a stable, machine-readable error identifier returned in the API
// envelope (api-spec.md). UI maps these to user messages.
type ErrorCode string

// Typed error codes (api-spec.md § Error codes). Keep in sync with the frontend
// error-code→message map.
const (
	CodeInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE" // 422
	CodeInsufficientHolding ErrorCode = "INSUFFICIENT_HOLDING" // 422
	CodeBondOfferInvalid    ErrorCode = "BOND_OFFER_INVALID"   // 422
	CodeQuoteExpired        ErrorCode = "QUOTE_EXPIRED"        // 409
	CodeCancelNotAllowed    ErrorCode = "CANCEL_NOT_ALLOWED"   // 409
	CodeRangeTooLarge       ErrorCode = "RANGE_TOO_LARGE"      // 400
	CodeExportTooLarge      ErrorCode = "EXPORT_TOO_LARGE"     // 400
	CodeTxnNotFound         ErrorCode = "TXN_NOT_FOUND"        // 404
	CodeDuplicateEmail      ErrorCode = "DUPLICATE_EMAIL"      // 409
	CodeRiskAckRequired     ErrorCode = "RISK_ACK_REQUIRED"    // 422
	CodeSuitabilityRequired ErrorCode = "SUITABILITY_REQUIRED" // 403
	CodeUnauthorized        ErrorCode = "UNAUTHORIZED"         // 401
	CodeForbidden           ErrorCode = "FORBIDDEN"            // 403
	CodeValidationError     ErrorCode = "VALIDATION_ERROR"     // 400
	CodePriceUnavailable    ErrorCode = "PRICE_UNAVAILABLE"    // 409
	// CodeInternal is used for unexpected 5xx failures (not in api-spec's
	// client list but required for operational errors — operations.md).
	CodeInternal ErrorCode = "INTERNAL_ERROR" // 500
)

// httpStatusByCode maps each error code to its HTTP status (api-spec.md).
var httpStatusByCode = map[ErrorCode]int{
	CodeInsufficientBalance: http.StatusUnprocessableEntity, // 422
	CodeInsufficientHolding: http.StatusUnprocessableEntity, // 422
	CodeBondOfferInvalid:    http.StatusUnprocessableEntity, // 422
	CodeQuoteExpired:        http.StatusConflict,            // 409
	CodeCancelNotAllowed:    http.StatusConflict,            // 409
	CodeRangeTooLarge:       http.StatusBadRequest,          // 400
	CodeExportTooLarge:      http.StatusBadRequest,          // 400
	CodeTxnNotFound:         http.StatusNotFound,            // 404
	CodeDuplicateEmail:      http.StatusConflict,            // 409
	CodeRiskAckRequired:     http.StatusUnprocessableEntity, // 422
	CodeSuitabilityRequired: http.StatusForbidden,           // 403
	CodeUnauthorized:        http.StatusUnauthorized,        // 401
	CodeForbidden:           http.StatusForbidden,           // 403
	CodeValidationError:     http.StatusBadRequest,          // 400
	CodePriceUnavailable:    http.StatusConflict,            // 409
	CodeInternal:            http.StatusInternalServerError, // 500
}

// HTTPStatus returns the HTTP status for a code, defaulting to 500 for unknown
// codes so unmapped failures never leak as success.
func (c ErrorCode) HTTPStatus() int {
	if status, ok := httpStatusByCode[c]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// AppError is a domain error carrying a typed code, a safe message, and
// optional structured details. Services return these; the HTTP layer renders
// them into the error envelope with the correct status.
type AppError struct {
	Code    ErrorCode
	Message string
	Details map[string]any
	// Err is an optional wrapped cause for logging; never serialized to clients.
	Err error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return string(e.Code) + ": " + e.Message + ": " + e.Err.Error()
	}
	return string(e.Code) + ": " + e.Message
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As.
func (e *AppError) Unwrap() error { return e.Err }

// NewError builds an AppError with a code and message.
func NewError(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// WithDetails attaches structured details and returns the error for chaining.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// WithCause wraps an underlying error (kept out of the client response).
func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}
