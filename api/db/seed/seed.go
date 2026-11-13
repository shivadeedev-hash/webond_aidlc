// Package seed applies the demo dataset (seed.sql, Task 6.1 / US-032) to the
// database on API boot.
//
// Why on boot (not only a mysql init-dir mount): the MySQL image only runs its
// /docker-entrypoint-initdb.d scripts on a *fresh* data volume. When the volume
// already exists (the common case in dev / CI reruns), that hook never fires.
// Running the seed from the API after the DB is reachable covers both cases.
//
// Safe to run every boot: seed.sql is idempotent (fixed PKs with
// ON DUPLICATE KEY UPDATE and guarded inserts), so re-applying converges to the
// same state instead of duplicating rows. Seeding is skipped in production and
// when SEED_ON_BOOT=false.
package seed

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // driver registration (dedicated conn)
)

// sqlScript is the demo seed, embedded at build time so the binary is
// self-contained (no file to ship alongside the distroless image).
//
//go:embed seed.sql
var sqlScript string

// applyTimeout bounds the seed run so a boot never hangs indefinitely.
const applyTimeout = 60 * time.Second

// Apply runs the embedded seed against the database identified by dsn.
//
// It opens a dedicated short-lived connection with multiStatements enabled (the
// seed relies on session state — SET @now and TEMPORARY tables — that must
// persist across statements within one connection). The main application pool
// is left untouched. Errors are returned to the caller, which logs and decides
// whether to continue (seeding is best-effort and must not block serving).
func Apply(ctx context.Context, dsn string, logger *slog.Logger) error {
	conn, err := sql.Open("mysql", withMultiStatements(dsn))
	if err != nil {
		return fmt.Errorf("open seed connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Keep the whole script on a single connection so SET @now and the
	// TEMPORARY tables are visible to every statement.
	conn.SetMaxOpenConns(1)

	runCtx, cancel := context.WithTimeout(ctx, applyTimeout)
	defer cancel()

	// Schema guard: if migrations have not been applied yet (e.g. a brand-new
	// volume), skip rather than erroring so the API still boots. The next boot
	// after `migrate up` will seed. Idempotency makes re-seeding safe.
	var hasUsers int
	if err := conn.QueryRowContext(runCtx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'",
	).Scan(&hasUsers); err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if hasUsers == 0 {
		if logger != nil {
			logger.Warn("demo seed skipped: schema not migrated yet (run `migrate up` first)")
		}
		return nil
	}

	if _, err := conn.ExecContext(runCtx, sqlScript); err != nil {
		return fmt.Errorf("apply seed: %w", err)
	}
	if logger != nil {
		logger.Info("demo seed applied", "credentials", "see README § Demo credentials")
	}
	return nil
}

// withMultiStatements appends multiStatements=true to the DSN query string so a
// single Exec can run the whole script. It is added only for the seed
// connection — never for the application pool — which keeps the app's blast
// radius for SQL injection unchanged (parameterized queries only).
func withMultiStatements(dsn string) string {
	if strings.Contains(dsn, "multiStatements=") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&multiStatements=true"
	}
	return dsn + "?multiStatements=true"
}
