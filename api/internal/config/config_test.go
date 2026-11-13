package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// setRequired sets the mandatory env vars for a valid config load.
func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "webond:webond@tcp(127.0.0.1:3306)/webond?parseTime=true")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
}

func TestLoad_Defaults(t *testing.T) {
	setRequired(t)
	// Clear optional vars so defaults are exercised regardless of the ambient
	// environment (t.Setenv restores prior values after the test).
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 30s", cfg.ShutdownTimeout)
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true, want false in development")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for missing required env vars, got nil")
	}
}

func TestLoad_BadDSN(t *testing.T) {
	setRequired(t)
	t.Setenv("DATABASE_URL", "not-a-dsn")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for malformed DSN, got nil")
	}
}

func TestLoad_BadShutdownTimeout(t *testing.T) {
	setRequired(t)
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "-5")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive SHUTDOWN_TIMEOUT_MS, got nil")
	}
}

const strongSecretA = "0123456789abcdef0123456789abcdef-A"
const strongSecretB = "0123456789abcdef0123456789abcdef-B"

func TestLoad_ProductionEnv(t *testing.T) {
	setRequired(t)
	t.Setenv("JWT_ACCESS_SECRET", strongSecretA)
	t.Setenv("JWT_REFRESH_SECRET", strongSecretB)
	t.Setenv("APP_ENV", "production")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("APP_PUBLIC_URL", "https://app.example.com/")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
	if cfg.LogLevel != slog.LevelWarn {
		t.Errorf("LogLevel = %v, want warn", cfg.LogLevel)
	}
	// Production defaults: no demo seed, no auto-migrate, no implicit CORS.
	if cfg.SeedOnBoot || cfg.MigrateOnBoot {
		t.Errorf("production must default SeedOnBoot/MigrateOnBoot to false, got %v/%v", cfg.SeedOnBoot, cfg.MigrateOnBoot)
	}
	if len(cfg.CORSAllowedOrigins) != 0 {
		t.Errorf("production CORS default should be empty, got %v", cfg.CORSAllowedOrigins)
	}
}

func TestLoad_ProductionRejectsWeakSecrets(t *testing.T) {
	cases := map[string]struct{ access, refresh string }{
		"placeholder": {"change-me-dev", strongSecretB},
		"too short":   {strongSecretA, "short-secret"},
		"marker":      {strongSecretA, "0123456789abcdef-change-me-0123456789abcdef"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setRequired(t)
			t.Setenv("APP_ENV", "production")
			t.Setenv("JWT_ACCESS_SECRET", c.access)
			t.Setenv("JWT_REFRESH_SECRET", c.refresh)
			if _, err := Load(); err == nil {
				t.Fatal("expected weak-secret error in production, got nil")
			}
		})
	}
}

func TestLoad_DevDefaults_MigrateSeedCORS(t *testing.T) {
	setRequired(t)
	t.Setenv("APP_ENV", "")
	t.Setenv("MIGRATE_ON_BOOT", "")
	t.Setenv("SEED_ON_BOOT", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.MigrateOnBoot || !cfg.SeedOnBoot {
		t.Errorf("dev should default MigrateOnBoot/SeedOnBoot to true, got %v/%v", cfg.MigrateOnBoot, cfg.SeedOnBoot)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("dev CORS default = %v, want [http://localhost:5173]", cfg.CORSAllowedOrigins)
	}
}

func TestLoad_ExplicitOverrides(t *testing.T) {
	setRequired(t)
	t.Setenv("MIGRATE_ON_BOOT", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://a.example.com , https://b.example.com ,, ")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MigrateOnBoot {
		t.Error("MIGRATE_ON_BOOT=false must disable migrate-on-boot")
	}
	want := []string{"https://a.example.com", "https://b.example.com"}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != want[0] || cfg.CORSAllowedOrigins[1] != want[1] {
		t.Errorf("CORS = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
	t.Setenv("MIGRATE_ON_BOOT", "maybe")
	if _, err := Load(); err == nil {
		t.Error("expected error for non-boolean MIGRATE_ON_BOOT")
	}
}

func TestLoad_DSNFromParts(t *testing.T) {
	setRequired(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_NAME", "webond")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "p@ss/w:rd?&=")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The password's special characters must survive a round trip through the
	// driver's own DSN parser (i.e. be escaped correctly).
	parsed, err := mysql.ParseDSN(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("composed DSN does not parse: %v (%s)", err, cfg.DatabaseURL)
	}
	if parsed.User != "app" || parsed.Passwd != "p@ss/w:rd?&=" || parsed.Addr != "db.internal:3307" || parsed.DBName != "webond" || !parsed.ParseTime {
		t.Errorf("unexpected parsed DSN: %+v", parsed)
	}
}

func TestLoad_DSNPartsMissing(t *testing.T) {
	setRequired(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DB_* parts are incomplete")
	}
}

func TestLoadDatabase_DoesNotNeedJWTSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "webond:webond@tcp(127.0.0.1:3306)/webond?parseTime=true")
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")
	if _, err := LoadDatabase(); err != nil {
		t.Fatalf("LoadDatabase must not require JWT secrets: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("Load must still require JWT secrets")
	}
}

func TestLoad_TLSCAFileInvalid(t *testing.T) {
	setRequired(t)
	t.Setenv("DB_TLS_CA_FILE", "/definitely/not/here.pem")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for unreadable DB_TLS_CA_FILE")
	}
	bad := t.TempDir() + "/bad.pem"
	if err := os.WriteFile(bad, []byte("not a pem"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DB_TLS_CA_FILE", bad)
	if _, err := Load(); err == nil {
		t.Fatal("expected error for DB_TLS_CA_FILE without valid certificates")
	}
}

// writeTestCA writes a self-signed CA certificate in PEM form to a temp file.
func writeTestCA(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "webond-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/ca.pem"
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_TLSCAFileAddsTLSToDSN(t *testing.T) {
	setRequired(t)
	t.Setenv("DB_TLS_CA_FILE", writeTestCA(t))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := mysql.ParseDSN(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("DSN with tls does not parse: %v (%s)", err, cfg.DatabaseURL)
	}
	if parsed.TLSConfig != tlsConfigName {
		t.Errorf("TLSConfig = %q, want %q (dsn=%s)", parsed.TLSConfig, tlsConfigName, cfg.DatabaseURL)
	}
}

// clearEmailEnv blanks every e-mail setting so a test starts from defaults.
func clearEmailEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM", "SMTP_TLS", "APP_PUBLIC_URL", "EMAIL_POLL_INTERVAL_MS", "APP_ENV"} {
		t.Setenv(k, "")
	}
}

func TestLoad_EmailDisabledByDefault(t *testing.T) {
	setRequired(t)
	clearEmailEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.EmailEnabled() {
		t.Error("e-mail must be disabled without SMTP_HOST")
	}
	if cfg.AppPublicURL != "http://localhost:5173" || !cfg.ExposeDevTokens() {
		t.Errorf("dev defaults wrong: url=%q exposeTokens=%v", cfg.AppPublicURL, cfg.ExposeDevTokens())
	}
}

func TestLoad_EmailMailpitAndGmailStyles(t *testing.T) {
	setRequired(t)
	clearEmailEnv(t)
	t.Setenv("SMTP_HOST", "mailpit")
	t.Setenv("SMTP_PORT", "1025")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("mailpit config: %v", err)
	}
	if cfg.SMTP.TLSMode != "none" || cfg.SMTP.From == "" {
		t.Errorf("mailpit defaults wrong: %+v", cfg.SMTP.TLSMode)
	}

	t.Setenv("SMTP_HOST", "smtp.gmail.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USERNAME", "someone@gmail.com")
	t.Setenv("SMTP_PASSWORD", "app-password")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("gmail config: %v", err)
	}
	if cfg.SMTP.TLSMode != "tls" || cfg.SMTP.From != "WeBond <someone@gmail.com>" {
		t.Errorf("gmail defaults wrong: tls=%q from=%q", cfg.SMTP.TLSMode, cfg.SMTP.From)
	}
}

func TestLoad_EmailRejectsUnsafeConfig(t *testing.T) {
	setRequired(t)
	clearEmailEnv(t)
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "25")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pw")
	if _, err := Load(); err == nil {
		t.Fatal("credentials over an unencrypted connection must be rejected")
	}
}

func TestLoad_ProductionRequiresPublicURL(t *testing.T) {
	setRequired(t)
	clearEmailEnv(t)
	t.Setenv("JWT_ACCESS_SECRET", strongSecretA)
	t.Setenv("JWT_REFRESH_SECRET", strongSecretB)
	t.Setenv("APP_ENV", "production")
	if _, err := Load(); err == nil {
		t.Fatal("production without APP_PUBLIC_URL must fail")
	}
	t.Setenv("APP_PUBLIC_URL", "https://app.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ExposeDevTokens() {
		t.Error("production must never expose one-time tokens")
	}
}
