// Package testdb picks the database that DB-backed tests may use. It is imported
// only by _test.go files.
//
// Tests create users, orders and DCA rounds, and some run system-wide jobs (for
// example "run all due DCA plans"). Pointed at the application database they
// change the demo data, so:
//   - only WEBOND_TEST_DSN is read (never DATABASE_URL, which points at the app);
//   - a DSN whose database is the app database ("webond") is refused unless
//     WEBOND_TEST_ALLOW_APP_DB=1 is set on purpose.
//
// Create the separate test database with ./scripts/test-db.sh.
package testdb

import (
	"testing"

	"os"

	"github.com/go-sql-driver/mysql"
)

// AppDatabase is the database name the application itself uses.
const AppDatabase = "webond"

// DSN returns the test database DSN, or skips the test when none is configured.
// It fails the test when the DSN points at the application database.
func DSN(t testing.TB) string {
	t.Helper()
	dsn, reason := Resolve(os.Getenv("WEBOND_TEST_DSN"), os.Getenv("WEBOND_TEST_ALLOW_APP_DB") == "1")
	switch {
	case dsn != "":
		return dsn
	case reason == "":
		t.Skip("WEBOND_TEST_DSN not set; skipping DB-backed test (see ./scripts/test-db.sh)")
	default:
		t.Fatal(reason)
	}
	return ""
}

// Resolve applies the rules without a testing.T so they can be unit tested.
// It returns the DSN to use, or "" with a reason ("" reason = not configured).
func Resolve(dsn string, allowAppDB bool) (string, string) {
	if dsn == "" {
		return "", ""
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", "WEBOND_TEST_DSN is not a valid MySQL DSN: " + err.Error()
	}
	if cfg.DBName == AppDatabase && !allowAppDB {
		return "", "WEBOND_TEST_DSN points at the application database \"" + AppDatabase +
			"\"; tests would modify the demo data. Use a separate test database (./scripts/test-db.sh) " +
			"or set WEBOND_TEST_ALLOW_APP_DB=1 if you really mean it."
	}
	return dsn, ""
}
