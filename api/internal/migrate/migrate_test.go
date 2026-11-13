package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/webond/api/db/migrations"
)

// scratchDB creates a throwaway database on the MySQL server named by
// WEBOND_TEST_ROOT_DSN (a privileged account, e.g.
// root:webond-root@tcp(127.0.0.1:3306)/) and returns a DSN for it plus a handle
// to inspect it. The test is skipped when the variable is unset so the suite
// still passes without a database.
func scratchDB(t *testing.T) (dsn string, conn *sql.DB) {
	t.Helper()
	rootDSN := os.Getenv("WEBOND_TEST_ROOT_DSN")
	if rootDSN == "" {
		t.Skip("WEBOND_TEST_ROOT_DSN not set; skipping DB-backed migration test")
	}
	cfg, err := mysql.ParseDSN(rootDSN)
	if err != nil {
		t.Fatalf("parse WEBOND_TEST_ROOT_DSN: %v", err)
	}
	admin, err := sql.Open("mysql", rootDSN)
	if err != nil {
		t.Fatalf("open admin: %v", err)
	}
	name := fmt.Sprintf("webond_migtest_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4"); err != nil {
		t.Fatalf("create scratch db: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP DATABASE IF EXISTS `" + name + "`")
		_ = admin.Close()
	})
	cfg.DBName = name
	cfg.ParseTime = true
	dsn = cfg.FormatDSN()
	conn, err = sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open scratch: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return dsn, conn
}

func countTables(t *testing.T, conn *sql.DB) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name <> 'schema_migrations'",
	).Scan(&n); err != nil {
		t.Fatalf("count tables: %v", err)
	}
	return n
}

func TestUp_FreshDatabaseThenIdempotent(t *testing.T) {
	dsn, conn := scratchDB(t)
	ctx := context.Background()

	if err := Up(ctx, dsn, nil); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	if got := countTables(t, conn); got != 20 {
		t.Fatalf("tables after Up = %d, want 20", got)
	}
	var version int
	var dirty bool
	if err := conn.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if version != 11 || dirty {
		t.Fatalf("schema_migrations = (%d, dirty=%v), want (11, false)", version, dirty)
	}

	// Second run is a no-op and must not error (idempotent boot).
	if err := Up(ctx, dsn, nil); err != nil {
		t.Fatalf("second Up should be a no-op, got: %v", err)
	}
	if got := countTables(t, conn); got != 20 {
		t.Fatalf("tables after second Up = %d, want 20", got)
	}
}

func TestUp_DirtyDatabaseIsNotAutoRepaired(t *testing.T) {
	dsn, conn := scratchDB(t)
	ctx := context.Background()
	if err := Up(ctx, dsn, nil); err != nil {
		t.Fatalf("setup Up: %v", err)
	}
	if _, err := conn.Exec("UPDATE schema_migrations SET dirty = 1"); err != nil {
		t.Fatalf("mark dirty: %v", err)
	}
	err := Up(ctx, dsn, nil)
	if err == nil {
		t.Fatal("expected an error for a dirty database")
	}
	if !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("error should mention the dirty state, got: %v", err)
	}
}

func TestUp_UnreachableDatabaseFailsClearly(t *testing.T) {
	// Port 1 is not listening; one attempt, no delay keeps the test instant.
	err := up(context.Background(), "u:p@tcp(127.0.0.1:1)/x", migrations.FS, nil, 1, 0)
	if err == nil {
		t.Fatal("expected an error for an unreachable database")
	}
	if !strings.Contains(err.Error(), "not reachable") {
		t.Fatalf("error should say the database is not reachable, got: %v", err)
	}
}

func TestWithMultiStatements(t *testing.T) {
	cases := map[string]string{
		"u:p@tcp(h:1)/d":                       "u:p@tcp(h:1)/d?multiStatements=true",
		"u:p@tcp(h:1)/d?parseTime=true":        "u:p@tcp(h:1)/d?parseTime=true&multiStatements=true",
		"u:p@tcp(h:1)/d?multiStatements=false": "u:p@tcp(h:1)/d?multiStatements=false",
	}
	for in, want := range cases {
		if got := withMultiStatements(in); got != want {
			t.Errorf("withMultiStatements(%q) = %q, want %q", in, got, want)
		}
	}
}
