// Package logging builds the application's structured logger (operations.md
// § Logging). Production emits JSON; development emits human-readable text.
package logging

import (
	"log/slog"
	"os"
)

// New returns a slog.Logger. When production is true it uses a JSON handler
// (machine-parseable event stream, 12-factor logs); otherwise a text handler
// for readable local development.
func New(level slog.Level, production bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if production {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// MaskEmail reduces an email to its first 3 characters plus the domain so the
// full address is never written to logs (operations.md § Sensitive Data Rules).
func MaskEmail(email string) string {
	at := -1
	for i, r := range email {
		if r == '@' {
			at = i
			break
		}
	}
	if at <= 0 {
		return "***"
	}
	local := email[:at]
	domain := email[at:]
	if len(local) <= 3 {
		return local[:1] + "***" + domain
	}
	return local[:3] + "***" + domain
}
