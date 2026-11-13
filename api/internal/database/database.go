// Package database owns the MySQL connection pool and a transaction helper
// (components.md § Repositories — all order/balance mutations run inside a DB
// transaction, NFR-05).
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver registration
)

// Pool-tuning defaults. Conservative values suitable for a single API node.
const (
	maxOpenConns    = 25
	maxIdleConns    = 10
	connMaxLifetime = 5 * time.Minute
	pingTimeout     = 5 * time.Second
)

// Open creates and verifies a MySQL connection pool from the DSN. It pings once
// so startup fails fast (exit 1 upstream) when the DB is unreachable
// (operations.md § Startup validation).
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// TxFn is a unit of work executed within a transaction.
type TxFn func(tx *sql.Tx) error

// WithinTx runs fn inside a single transaction, committing on success and
// rolling back on error or panic. This is the canonical way services perform
// multi-statement mutations atomically (e.g. reserve cash + create order).
func WithinTx(ctx context.Context, db *sql.DB, fn TxFn) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-panic after cleanup so the recovery middleware sees it
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				err = fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
			}
			return
		}
		if cmErr := tx.Commit(); cmErr != nil {
			err = fmt.Errorf("commit tx: %w", cmErr)
		}
	}()

	err = fn(tx)
	return err
}
