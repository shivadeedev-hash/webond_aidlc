// Package config loads and validates the API configuration from the
// environment (12-factor). Secrets are never hardcoded; required values that
// are missing or malformed cause a clear startup error so the process can exit
// with a non-zero status (operations.md § Startup validation).
package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/webond/api/internal/mail"
)

// Default values for optional settings (operations.md § Configuration Management).
const (
	defaultPort            = "8080"
	defaultAppEnv          = "development"
	defaultLogLevel        = "info"
	defaultShutdownTimeout = 30000 // milliseconds
	defaultDBPort          = "3306"
	defaultDevCORSOrigin   = "http://localhost:5173"
	defaultSMTPPort        = 587
	defaultSMTPFrom        = "WeBond <no-reply@webond.local>"
	defaultEmailPoll       = 2 * time.Second

	// minProdSecretLen is the minimum JWT secret length accepted in production.
	minProdSecretLen = 32
	// weakSecretMarker is the placeholder shipped in .env.example; it must never
	// be used in production.
	weakSecretMarker = "change-me"

	// tlsConfigName is the name under which the optional DB CA bundle is
	// registered with the MySQL driver (referenced as tls=<name> in the DSN).
	tlsConfigName = "webond-db"
)

// Config holds all runtime configuration for the API.
type Config struct {
	AppEnv             string        // runtime env: development | production
	Port               string        // listen port
	LogLevel           slog.Level    // parsed log level
	DatabaseURL        string        // MySQL DSN (required, sensitive)
	JWTAccessSecret    string        // access-token signing secret (required, sensitive)
	JWTRefreshSecret   string        // refresh-token signing secret (required, sensitive)
	ShutdownTimeout    time.Duration // graceful shutdown window
	SeedOnBoot         bool          // apply the demo seed on boot (dev convenience, Task 6.1)
	MigrateOnBoot      bool          // apply pending DB migrations on boot (Task 10.1)
	CORSAllowedOrigins []string      // browser origins allowed to call the API

	// E-mail delivery (Task 10.8). SMTP.Host empty → messages are only logged.
	SMTP              mail.SMTPConfig // SMTP_* settings (password is sensitive)
	AppPublicURL      string          // web origin used in e-mail links
	EmailPollInterval time.Duration   // how often the outbox is checked
}

// EmailEnabled reports whether an SMTP server is configured.
func (c *Config) EmailEnabled() bool {
	return c.SMTP.Host != ""
}

// ExposeDevTokens reports whether register/forgot-password responses may echo
// the one-time token (local development only, never production).
func (c *Config) ExposeDevTokens() bool {
	return !c.IsProduction()
}

// IsProduction reports whether the API runs in a production environment.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// Load reads configuration from the environment and validates it. It returns an
// error (rather than exiting) so the caller controls process termination and
// tests can exercise validation without killing the test binary.
func Load() (*Config, error) {
	cfg, err := loadBase()
	if err != nil {
		return nil, err
	}

	// Required, sensitive values — fail fast with a clear message.
	var missing []string

	cfg.JWTAccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	if strings.TrimSpace(cfg.JWTAccessSecret) == "" {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}

	cfg.JWTRefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	if strings.TrimSpace(cfg.JWTRefreshSecret) == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// Production must never run with the placeholder / short JWT secrets.
	if cfg.IsProduction() {
		for name, secret := range map[string]string{
			"JWT_ACCESS_SECRET":  cfg.JWTAccessSecret,
			"JWT_REFRESH_SECRET": cfg.JWTRefreshSecret,
		} {
			if isWeakSecret(secret) {
				return nil, fmt.Errorf("%s is too weak for production (need >= %d characters and not the %q placeholder)",
					name, minProdSecretLen, weakSecretMarker)
			}
		}
	}

	// Parse the shutdown timeout; fall back to the default on bad input.
	ms := defaultShutdownTimeout
	if raw := os.Getenv("SHUTDOWN_TIMEOUT_MS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("SHUTDOWN_TIMEOUT_MS must be a positive integer, got %q", raw)
		}
		ms = parsed
	}
	cfg.ShutdownTimeout = time.Duration(ms) * time.Millisecond

	// Seed-on-boot: default true in non-production so a fresh `docker compose up`
	// has demo data; always off in production. SEED_ON_BOOT overrides explicitly.
	cfg.SeedOnBoot, err = boolEnv("SEED_ON_BOOT", !cfg.IsProduction())
	if err != nil {
		return nil, err
	}

	// Migrate-on-boot: default true in non-production for a zero-step local
	// start; off in production, where migrations run as a one-off task
	// (`webond migrate`) before the new version is rolled out.
	cfg.MigrateOnBoot, err = boolEnv("MIGRATE_ON_BOOT", !cfg.IsProduction())
	if err != nil {
		return nil, err
	}

	// CORS: explicit list wins; otherwise the local dev origin in non-production
	// and no cross-origin access in production (same-origin behind CloudFront).
	cfg.CORSAllowedOrigins = splitList(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(cfg.CORSAllowedOrigins) == 0 && !cfg.IsProduction() {
		cfg.CORSAllowedOrigins = []string{defaultDevCORSOrigin}
	}

	if err := loadEmail(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadEmail reads the SMTP settings (Task 10.8). Without SMTP_HOST the API
// still runs and only logs that e-mails were not delivered.
func loadEmail(cfg *Config) error {
	cfg.AppPublicURL = strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/")
	if cfg.AppPublicURL == "" {
		if cfg.IsProduction() {
			return fmt.Errorf("APP_PUBLIC_URL is required in production (used for links in e-mails)")
		}
		cfg.AppPublicURL = defaultDevCORSOrigin
	}
	if !strings.HasPrefix(cfg.AppPublicURL, "http://") && !strings.HasPrefix(cfg.AppPublicURL, "https://") {
		return fmt.Errorf("APP_PUBLIC_URL must start with http:// or https://, got %q", cfg.AppPublicURL)
	}

	cfg.EmailPollInterval = defaultEmailPoll
	if raw := os.Getenv("EMAIL_POLL_INTERVAL_MS"); raw != "" {
		ms, err := strconv.Atoi(raw)
		if err != nil || ms < 200 {
			return fmt.Errorf("EMAIL_POLL_INTERVAL_MS must be an integer >= 200, got %q", raw)
		}
		cfg.EmailPollInterval = time.Duration(ms) * time.Millisecond
	}

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		return nil
	}
	port := defaultSMTPPort
	if raw := os.Getenv("SMTP_PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("SMTP_PORT must be a number, got %q", raw)
		}
		port = p
	}
	cfg.SMTP = mail.SMTPConfig{
		Host:     host,
		Port:     port,
		Username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		TLSMode:  strings.ToLower(getEnv("SMTP_TLS", defaultSMTPTLS(port))),
	}
	// Gmail and most providers require From to be the authenticated account.
	if cfg.SMTP.From == "" && strings.Contains(cfg.SMTP.Username, "@") {
		cfg.SMTP.From = "WeBond <" + cfg.SMTP.Username + ">"
	}
	if cfg.SMTP.From == "" {
		cfg.SMTP.From = defaultSMTPFrom
	}
	if err := cfg.SMTP.Validate(); err != nil {
		return fmt.Errorf("e-mail configuration: %w", err)
	}
	return nil
}

// defaultSMTPTLS picks the usual TLS mode for a port: 465 implicit TLS,
// 587 STARTTLS, anything else (e.g. Mailpit 1025) none.
func defaultSMTPTLS(port int) string {
	switch port {
	case 465:
		return mail.TLSImplicit
	case 587:
		return mail.TLSStart
	default:
		return mail.TLSNone
	}
}

// LoadDatabase loads only what the one-off `webond migrate` command needs: the
// environment, log level and database connection. JWT secrets are not required
// so a migration task never needs access to signing material.
func LoadDatabase() (*Config, error) {
	return loadBase()
}

// loadBase reads the settings shared by the server and the migrate command.
func loadBase() (*Config, error) {
	cfg := &Config{
		AppEnv:   getEnv("APP_ENV", defaultAppEnv),
		Port:     getEnv("PORT", defaultPort),
		LogLevel: parseLogLevel(getEnv("LOG_LEVEL", defaultLogLevel)),
	}

	dsn, err := databaseDSN()
	if err != nil {
		return nil, err
	}
	cfg.DatabaseURL = dsn
	return cfg, nil
}

// databaseDSN returns the MySQL DSN. DATABASE_URL wins when present; otherwise
// it is composed from DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD (what ECS
// injects from Secrets Manager). DB_TLS_CA_FILE optionally enables verified TLS.
func databaseDSN() (string, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	if dsn == "" {
		host := strings.TrimSpace(os.Getenv("DB_HOST"))
		name := strings.TrimSpace(os.Getenv("DB_NAME"))
		user := strings.TrimSpace(os.Getenv("DB_USER"))
		pass := os.Getenv("DB_PASSWORD")

		var missing []string
		if host == "" {
			missing = append(missing, "DB_HOST")
		}
		if name == "" {
			missing = append(missing, "DB_NAME")
		}
		if user == "" {
			missing = append(missing, "DB_USER")
		}
		if pass == "" {
			missing = append(missing, "DB_PASSWORD")
		}
		if len(missing) > 0 {
			return "", fmt.Errorf("missing required environment variables: DATABASE_URL (or all of DB_HOST, DB_NAME, DB_USER, DB_PASSWORD; missing %s)",
				strings.Join(missing, ", "))
		}

		my := mysql.NewConfig()
		my.User = user
		my.Passwd = pass
		my.Net = "tcp"
		my.Addr = host + ":" + getEnv("DB_PORT", defaultDBPort)
		my.DBName = name
		my.ParseTime = true
		my.Loc = bangkok()
		dsn = my.FormatDSN() // escapes special characters in the password
	}

	// Validate DSN shape early (full connectivity is checked by a DB ping).
	if !looksLikeMySQLDSN(dsn) {
		return "", fmt.Errorf("DATABASE_URL does not look like a MySQL DSN (expected user:pass@tcp(host:port)/db)")
	}

	if caFile := strings.TrimSpace(os.Getenv("DB_TLS_CA_FILE")); caFile != "" {
		withTLS, err := enableTLS(dsn, caFile)
		if err != nil {
			return "", err
		}
		dsn = withTLS
	}
	return dsn, nil
}

// enableTLS registers the CA bundle with the MySQL driver and appends
// tls=<name> to the DSN so every connection (pool, migrate, seed) verifies the
// server certificate (encryption in transit to the database).
func enableTLS(dsn, caFile string) (string, error) {
	pem, err := os.ReadFile(caFile) //nolint:gosec // path comes from operator-controlled env
	if err != nil {
		return "", fmt.Errorf("read DB_TLS_CA_FILE: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return "", fmt.Errorf("DB_TLS_CA_FILE %q contains no valid PEM certificates", caFile)
	}
	if err := mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return "", fmt.Errorf("register DB TLS config: %w", err)
	}
	if strings.Contains(dsn, "tls=") {
		return dsn, nil
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "tls=" + tlsConfigName, nil
}

// isWeakSecret reports whether a JWT secret is unacceptable in production.
func isWeakSecret(s string) bool {
	return len(s) < minProdSecretLen || strings.Contains(strings.ToLower(s), weakSecretMarker)
}

// boolEnv parses an optional boolean env var, returning fallback when unset.
func boolEnv(key string, fallback bool) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean, got %q", key, raw)
	}
	return parsed, nil
}

// splitList splits a comma-separated list, trimming blanks and empties.
func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// bangkok returns the project timezone (Asia/Bangkok), falling back to a fixed
// +07:00 zone when tzdata is unavailable (e.g. distroless images).
func bangkok() *time.Location {
	if loc, err := time.LoadLocation("Asia/Bangkok"); err == nil {
		return loc
	}
	return time.FixedZone("Asia/Bangkok", 7*3600)
}

// getEnv returns the env var value or a fallback when unset/empty.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// parseLogLevel maps a string level to a slog.Level (defaults to info).
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// looksLikeMySQLDSN does a lightweight shape check on the go-sql-driver DSN
// format: it must contain the "@tcp(" protocol marker and a "/" database path.
func looksLikeMySQLDSN(dsn string) bool {
	return strings.Contains(dsn, "@tcp(") && strings.Contains(dsn, ")/")
}
