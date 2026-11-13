// Package migrate applies the embedded SQL migrations (api/db/migrations) to
// MySQL using golang-migrate. It writes to the same schema_migrations table as
// the golang-migrate CLI, so a database created with the CLI is recognised as
// already up to date.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	gm "github.com/golang-migrate/migrate/v4"
	gmmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/go-sql-driver/mysql" // database/sql driver registration

	"github.com/webond/api/db/migrations"
)

// connectRetries / connectDelay bound how long Up waits for the database to
// accept connections (e.g. RDS or MySQL still starting when the task boots).
const (
	connectRetries = 20
	connectDelay   = 3 * time.Second
)

// Up applies every pending migration embedded in the binary. It is idempotent:
// when the database is already current it logs and returns nil. A dirty
// database (a previous migration failed half-way) is NOT repaired
// automatically — it returns an error so an operator can inspect it.
func Up(ctx context.Context, dsn string, logger *slog.Logger) error {
	return up(ctx, dsn, migrations.FS, logger, connectRetries, connectDelay)
}

func up(ctx context.Context, dsn string, src fs.FS, logger *slog.Logger, retries int, delay time.Duration) error {
	// The migration SQL files contain several statements each, which the MySQL
	// driver only accepts when multiStatements is on. This dedicated
	// connection is separate from the application pool.
	db, err := sql.Open("mysql", withMultiStatements(dsn))
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := waitForDB(ctx, db, retries, delay, logger); err != nil {
		return err
	}

	driver, err := gmmysql.WithInstance(db, &gmmysql.Config{})
	if err != nil {
		return fmt.Errorf("init migrate mysql driver: %w", err)
	}
	source, err := iofs.New(src, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}
	m, err := gm.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	before, _, _ := m.Version() // ErrNilVersion on an empty DB is fine

	switch err := m.Up(); {
	case err == nil:
		after, _, _ := m.Version()
		if logger != nil {
			logger.Info("database migrated", "from", before, "to", after)
		}
		return nil
	case errors.Is(err, gm.ErrNoChange):
		if logger != nil {
			logger.Info("database schema is up to date", "version", before)
		}
		return nil
	default:
		var dirty gm.ErrDirty
		if errors.As(err, &dirty) {
			return fmt.Errorf("database is in a dirty state at version %d (a previous migration failed); fix it manually before retrying: %w",
				dirty.Version, err)
		}
		return fmt.Errorf("apply migrations: %w", err)
	}
}

// waitForDB pings the database until it answers or the retries run out.
func waitForDB(ctx context.Context, db *sql.DB, retries int, delay time.Duration, logger *slog.Logger) error {
	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		lastErr = db.PingContext(pingCtx)
		cancel()
		if lastErr == nil {
			return nil
		}
		if logger != nil {
			logger.Warn("waiting for database", "attempt", attempt, "of", retries, "error", lastErr)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for database: %w", ctx.Err())
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("database not reachable after %d attempts: %w", retries, lastErr)
}

// withMultiStatements appends multiStatements=true to the DSN query string.
func withMultiStatements(dsn string) string {
	if strings.Contains(dsn, "multiStatements=") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&multiStatements=true"
	}
	return dsn + "?multiStatements=true"
}
