// Package service holds the WeBond business-logic layer (components.md §
// Backend Components). Services orchestrate repositories (sqlc over MySQL),
// the money package (shopspring/decimal math), the auth package (JWT/bcrypt)
// and the domain error codes, and expose methods the Gin handlers call. All
// order/balance mutations run inside Store.WithinTx (NFR-05).
package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
)

// Clock is an injectable time source so time-dependent logic (login lockout,
// cut-off, DCA next-run, quote expiry, token expiry) is deterministic in tests
// (integration.md § Integration Testing — injectable clock).
type Clock interface {
	Now() time.Time
}

// systemClock is the production clock backed by time.Now in Asia/Bangkok.
type systemClock struct{ loc *time.Location }

// Bangkok is the project timezone (product.md — timezone Asia/Bangkok).
var Bangkok = mustLoadBangkok()

func mustLoadBangkok() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		// Fall back to a fixed +07:00 zone if the tzdata is unavailable.
		return time.FixedZone("Asia/Bangkok", 7*3600)
	}
	return loc
}

// NewSystemClock returns a Clock that reports the current time in Asia/Bangkok.
func NewSystemClock() Clock { return &systemClock{loc: Bangkok} }

func (c *systemClock) Now() time.Time { return time.Now().In(c.loc) }

// FixedClock is a test clock that always returns a set time; advance with Set.
type FixedClock struct {
	mu sync.Mutex
	t  time.Time
}

// NewFixedClock builds a FixedClock pinned to t.
func NewFixedClock(t time.Time) *FixedClock { return &FixedClock{t: t} }

// Now returns the pinned time.
func (c *FixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

// Set repins the clock (used by tests to advance time).
func (c *FixedClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = t
}

// Advance moves the pinned time forward by d.
func (c *FixedClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// txnRefCounter is a process-local monotonic counter backing the 7-digit
// running segment of a txn_ref. The DB UNIQUE(txn_ref) constraint is the
// durable guarantee; this counter plus a random fallback keeps collisions
// vanishingly unlikely within a single day (BR-16).
var (
	txnRefMu      sync.Mutex
	txnRefSeq     uint64
	txnRefSeqDate string
)

// NewTxnRef builds a transaction reference of the form WB + yyMMdd + 7-digit
// number (BR-16), e.g. WB2610280000045. The 7-digit segment combines a per-day
// running counter with a random offset so references stay unique across
// processes and API instances on the same day (the previous pure per-process
// counter collided when two processes — e.g. the API container and a test
// process — both started from 0 on the same date). The DB UNIQUE(txn_ref)
// constraint remains the durable guarantee; callers retry on the rare collision.
func NewTxnRef(now time.Time) string {
	txnRefMu.Lock()
	datePart := now.Format("060102")
	if datePart != txnRefSeqDate {
		txnRefSeqDate = datePart
		// Seed the day's counter at a random offset so independent processes do
		// not all begin at 0 and generate the same early references.
		txnRefSeq = randomSeqSeed()
	}
	txnRefSeq = (txnRefSeq + 1) % 10_000_000
	seq := txnRefSeq
	txnRefMu.Unlock()
	return fmt.Sprintf("WB%s%07d", datePart, seq)
}

// randomSeqSeed returns a crypto-random starting offset in [0, 10_000_000) for
// the per-day running counter. Falls back to 0 if the RNG is unavailable (the
// DB UNIQUE constraint still protects correctness).
func randomSeqSeed() uint64 {
	n, err := rand.Int(rand.Reader, big.NewInt(10_000_000))
	if err != nil {
		return 0
	}
	return n.Uint64()
}

// randomToken returns a URL-safe random hex token of n bytes (OTP/verify/reset
// tokens — US-002/US-004). Uses crypto/rand so tokens are unguessable.
func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// nullString wraps a possibly-empty string as sql.NullString (empty → NULL).
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullStringPtr wraps *string as sql.NullString (nil → unchanged marker not
// supported here; callers pass concrete values).
func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

// nullTime wraps a time as sql.NullTime.
func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

// nullInt16 wraps an int16 as sql.NullInt16 (0 → NULL).
func nullInt16(v int16) sql.NullInt16 {
	if v == 0 {
		return sql.NullInt16{}
	}
	return sql.NullInt16{Int16: v, Valid: true}
}

// nullDecimal wraps a decimal as decimal.NullDecimal (valid).
func nullDecimal(d decimal.Decimal) decimal.NullDecimal {
	return decimal.NullDecimal{Decimal: d, Valid: true}
}

// lastInsertID extracts the auto-increment id from an exec result.
func lastInsertID(res sql.Result, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return uint64(id), nil
}

// isNoRows reports whether err is sql.ErrNoRows (used to turn "not found" into
// a typed domain error at the service boundary).
func isNoRows(err error) bool {
	return err == sql.ErrNoRows
}

// ctxTODO is a small helper to keep call sites terse where a background context
// is acceptable (internal jobs). Request handlers pass the request context.
func ctxTODO() context.Context { return context.Background() }

// maskBankAccount masks all but the last 4 characters of a bank account number
// for display/storage (US-006, NFR-02 on-screen masking). Non-digits are kept
// positionally; short values are fully masked.
func maskBankAccount(acct string) string {
	if acct == "" {
		return ""
	}
	runes := []rune(acct)
	if len(runes) <= 4 {
		masked := make([]rune, len(runes))
		for i := range masked {
			masked[i] = 'x'
		}
		return string(masked)
	}
	keep := 4
	for i := 0; i < len(runes)-keep; i++ {
		if runes[i] != ' ' && runes[i] != '-' {
			runes[i] = 'x'
		}
	}
	return string(runes)
}

// decimalToDB ensures a decimal is quantized to the 4-dp storage scale before
// it is written (defensive; money helpers already quantize).
func decimalToDB(d decimal.Decimal) decimal.Decimal { return d.Round(4) }

// assetTypeValid reports whether s is one of the supported asset types.
func assetTypeValid(s string) bool {
	switch s {
	case string(db.OrdersAssetTypeFund), string(db.OrdersAssetTypeBond), string(db.OrdersAssetTypeGold):
		return true
	default:
		return false
	}
}
