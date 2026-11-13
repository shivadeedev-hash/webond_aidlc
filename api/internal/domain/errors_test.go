package domain

import (
	"errors"
	"net/http"
	"testing"
)

// TestErrorCodeHTTPStatus verifies every api-spec.md error code maps to the
// documented HTTP status.
func TestErrorCodeHTTPStatus(t *testing.T) {
	want := map[ErrorCode]int{
		CodeInsufficientBalance: 422,
		CodeInsufficientHolding: 422,
		CodeBondOfferInvalid:    422,
		CodeQuoteExpired:        409,
		CodeCancelNotAllowed:    409,
		CodeRangeTooLarge:       400,
		CodeExportTooLarge:      400,
		CodeTxnNotFound:         404,
		CodeDuplicateEmail:      409,
		CodeRiskAckRequired:     422,
		CodeSuitabilityRequired: 403,
		CodeUnauthorized:        401,
		CodeForbidden:           403,
		CodeValidationError:     400,
		CodePriceUnavailable:    409,
	}
	for code, status := range want {
		if got := code.HTTPStatus(); got != status {
			t.Errorf("%s.HTTPStatus() = %d, want %d", code, got, status)
		}
	}
}

func TestUnknownCodeDefaultsTo500(t *testing.T) {
	if got := ErrorCode("NOPE").HTTPStatus(); got != http.StatusInternalServerError {
		t.Errorf("unknown code status = %d, want 500", got)
	}
}

func TestAppError_WrapAndUnwrap(t *testing.T) {
	cause := errors.New("db down")
	err := NewError(CodeInternal, "boom").WithCause(cause)
	if !errors.Is(err, cause) {
		t.Error("errors.Is did not find wrapped cause")
	}
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("errors.As failed to extract *AppError")
	}
	if appErr.Code != CodeInternal {
		t.Errorf("code = %s, want INTERNAL_ERROR", appErr.Code)
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := NewError(CodeValidationError, "bad").WithDetails(map[string]any{"field": "email"})
	if err.Details["field"] != "email" {
		t.Error("WithDetails did not attach details")
	}
}
