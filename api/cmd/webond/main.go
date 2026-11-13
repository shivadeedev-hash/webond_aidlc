// Command webond starts the WeBond REST API server.
//
// Boot sequence (operations.md § Startup validation + Graceful Shutdown):
//  1. load + validate config from the environment (12-factor)
//  2. build the structured logger (slog)
//  3. open + ping the MySQL pool (fatal on failure)
//  4. build the router and start serving
//  5. on SIGINT/SIGTERM: mark draining, drain in-flight requests within
//     SHUTDOWN_TIMEOUT_MS, close the DB pool, exit
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embed zoneinfo so Asia/Bangkok resolves even in minimal images

	"github.com/webond/api/db/seed"
	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/config"
	"github.com/webond/api/internal/database"
	apihttp "github.com/webond/api/internal/http"
	"github.com/webond/api/internal/logging"
	"github.com/webond/api/internal/mail"
	"github.com/webond/api/internal/migrate"
	"github.com/webond/api/internal/repo"
	"github.com/webond/api/internal/service"
)

// migrateTimeout bounds a whole `webond migrate` run (DB wait + migrations).
const migrateTimeout = 10 * time.Minute

func main() {
	// Subcommands: `webond migrate` applies pending DB migrations and exits;
	// `webond seed` loads the demo dataset and exits. They are meant for one-off
	// deploy tasks (12-factor admin processes) and need only the database
	// settings, not the JWT secrets.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			os.Exit(runMigrate())
		case "seed":
			os.Exit(runSeed())
		default:
			slog.Error("unknown command", "command", os.Args[1], "usage", "webond [migrate|seed]")
			os.Exit(2)
		}
	}

	// 1. Configuration — a validation error is fatal before any logger exists,
	// so report to stderr and exit 1 (operations.md § Fatal).
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// 2. Logger.
	logger := logging.New(cfg.LogLevel, cfg.IsProduction())
	slog.SetDefault(logger)

	// 2b. Schema migrations (default on outside production). Runs before the
	// pool is opened so a database that is still starting is waited for, and a
	// failed migration stops the boot instead of serving against a bad schema.
	ctx := context.Background()
	if cfg.MigrateOnBoot {
		migCtx, cancel := context.WithTimeout(ctx, migrateTimeout)
		merr := migrate.Up(migCtx, cfg.DatabaseURL, logger)
		cancel()
		if merr != nil {
			logger.Error("database migration failed", "error", merr)
			os.Exit(1)
		}
	}

	// 3. Database pool + ping (fatal if unreachable).
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("cannot connect to database", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// 3b. Demo seed (Task 6.1 / US-032). Idempotent and skipped in production;
	// best-effort so a seed failure never blocks the API from serving.
	if cfg.SeedOnBoot {
		if serr := seed.Apply(ctx, cfg.DatabaseURL, logger); serr != nil {
			logger.Warn("demo seed skipped", "error", serr)
		}
	}

	// 4. Service + HTTP wiring.
	tokens := auth.NewManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	store := repo.NewStore(db)
	clock := service.NewSystemClock()

	notify := service.NewNotificationService(store, clock)
	walletSvc := service.NewWalletService(store, clock)
	tradingSvc := service.NewTradingService(store, walletSvc, notify, clock)
	portfolioSvc := service.NewPortfolioService(store, clock)
	priceGen := service.NewPriceGeneratorService(store, logger, clock)
	api := apihttp.NewAPI(cfg, apihttp.APIServices{
		Auth:      service.NewAuthService(store, tokens, notify, clock),
		Profile:   service.NewProfileService(store, clock),
		Wallet:    walletSvc,
		Catalog:   service.NewCatalogService(store, clock),
		Trading:   tradingSvc,
		DCA:       service.NewDCAService(store, tradingSvc, notify, clock),
		Portfolio: portfolioSvc,
		Dashboard: service.NewDashboardService(store, portfolioSvc, walletSvc, clock),
		History:   service.NewHistoryService(store, walletSvc, clock),
		Notify:    notify,
		Bond:      service.NewBondLifecycleService(store, walletSvc, notify, clock),
		Admin:     service.NewAdminService(store, clock),
		PriceGen:  priceGen,
	})

	// Resume the price generator if its persisted state is enabled (D3-7).
	priceGen.StartIfEnabled(ctx)
	defer priceGen.Stop()

	// E-mail delivery from the notification outbox (Task 10.8).
	var sender mail.Sender = mail.LogSender{Logger: logger}
	if cfg.EmailEnabled() {
		smtpSender, serr := mail.NewSMTPSender(cfg.SMTP)
		if serr != nil {
			logger.Error("e-mail configuration error", "error", serr)
			os.Exit(1)
		}
		sender = smtpSender
		logger.Info("e-mail delivery enabled", "smtpHost", cfg.SMTP.Host, "smtpPort", cfg.SMTP.Port, "tls", cfg.SMTP.TLSMode)
	} else if cfg.IsProduction() {
		logger.Warn("SMTP_HOST is not set: e-mails are NOT delivered, so new users cannot verify their account")
	}
	mailer := service.NewEmailDispatcher(store, sender, clock, logger, cfg.AppPublicURL, cfg.EmailPollInterval)
	mailer.Start(ctx)
	defer mailer.Stop()

	health := apihttp.NewHealthHandler(db)
	router := apihttp.NewRouter(apihttp.RouterDeps{
		Config:         cfg,
		Logger:         logger,
		Tokens:         tokens,
		Health:         health,
		API:            api,
		AllowedOrigins: cfg.CORSAllowedOrigins,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", srv.Addr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received, draining")

	health.StartDraining() // readiness → 503 so load balancers stop routing

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

// runMigrate implements `webond migrate`: apply pending migrations, then exit
// 0 on success (including "already up to date") or 1 on any failure so a
// deploy pipeline can gate the rollout on it.
func runMigrate() int {
	cfg, err := config.LoadDatabase()
	if err != nil {
		slog.Error("configuration error", "error", err)
		return 1
	}
	logger := logging.New(cfg.LogLevel, cfg.IsProduction())
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()
	if err := migrate.Up(ctx, cfg.DatabaseURL, logger); err != nil {
		logger.Error("database migration failed", "error", err)
		return 1
	}
	logger.Info("migrate command finished")
	return 0
}

// allowProdSeedEnv must be set to "1" to run `webond seed` with APP_ENV=production.
// The demo dataset has publicly documented passwords, so it is never loaded into a
// production database by accident.
const allowProdSeedEnv = "WEBOND_ALLOW_PROD_SEED"

// runSeed implements `webond seed`: load the idempotent demo dataset, then exit.
func runSeed() int {
	cfg, err := config.LoadDatabase()
	if err != nil {
		slog.Error("configuration error", "error", err)
		return 1
	}
	logger := logging.New(cfg.LogLevel, cfg.IsProduction())
	slog.SetDefault(logger)

	if cfg.IsProduction() && os.Getenv(allowProdSeedEnv) != "1" {
		logger.Error("refusing to load demo data into a production database",
			"hint", "set "+allowProdSeedEnv+"=1 only for a demo environment")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()
	if err := seed.Apply(ctx, cfg.DatabaseURL, logger); err != nil {
		logger.Error("seed failed", "error", err)
		return 1
	}
	logger.Info("seed command finished")
	return 0
}
