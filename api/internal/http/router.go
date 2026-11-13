package http

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/config"
	"github.com/webond/api/internal/http/middleware"
)

// RouterDeps are the dependencies needed to build the HTTP router.
type RouterDeps struct {
	Config *config.Config
	Logger *slog.Logger
	Tokens *auth.Manager
	Health *HealthHandler
	// API carries the wired services whose handlers are registered under /api.
	API *API
	// AllowedOrigins is the CORS allowlist (SPA dev + served origins).
	AllowedOrigins []string
}

// NewRouter builds the Gin engine with the base middleware stack, the health
// endpoints, and all domain routes under /api (api-spec.md; components.md §
// HTTP Layer).
func NewRouter(deps RouterDeps) *gin.Engine {
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Base middleware order: correlation id first, then recovery (so panics are
	// logged with the id), structured request logging, and CORS.
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(deps.Logger, deps.Config.IsProduction()))
	r.Use(middleware.Logger(deps.Logger))
	r.Use(middleware.CORS(deps.AllowedOrigins))

	// Health endpoints (unauthenticated — operations.md § Health & Readiness).
	r.GET("/health", deps.Health.Live)
	r.GET("/health/ready", deps.Health.Ready)

	if deps.API != nil {
		deps.API.RegisterRoutes(r, deps.Tokens)
	}

	return r
}

// RegisterRoutes wires every REST endpoint under /api using the envelope and
// the phase-4 middleware (auth/RBAC/idempotency) — paths/shapes/codes match
// api-spec.md exactly.
func (a *API) RegisterRoutes(r *gin.Engine, tokens *auth.Manager) {
	idem := middleware.Idempotency(middleware.NewMemoryIdempotencyStore(24 * time.Hour))
	authMW := middleware.Auth(tokens)
	adminMW := middleware.RequireRole(auth.RoleAdmin)

	api := r.Group("/api")

	// --- Auth (public) ---
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", a.handleRegister)
		authGroup.POST("/verify", a.handleVerify)
		authGroup.POST("/login", a.handleLogin)
		authGroup.POST("/refresh", a.handleRefresh)
		authGroup.POST("/logout", a.handleLogout)
		authGroup.POST("/resend-verification", a.handleResendVerification)
		authGroup.POST("/forgot-password", a.handleForgotPassword)
		authGroup.POST("/reset-password", a.handleResetPassword)
	}

	// --- Market / Assets (public, guest-accessible — US-005/011/012) ---
	api.GET("/assets", a.handleListAssets)
	api.GET("/assets/:type/:id", a.handleAssetDetail)
	api.GET("/assets/:type/:id/prices", a.handleAssetPrices)

	// --- Authenticated investor routes ---
	authed := api.Group("")
	authed.Use(authMW)
	{
		// Profile & risk
		authed.GET("/profile", a.handleGetProfile)
		authed.PUT("/profile", a.handleUpdateProfile)
		authed.PUT("/profile/notification-settings", a.handleNotificationSettings)
		authed.GET("/risk-assessment", a.handleGetRisk)
		authed.POST("/risk-assessment", a.handleSubmitRisk)

		// Wallet
		authed.GET("/wallet", a.handleGetWallet)
		authed.GET("/wallet/transactions", a.handleWalletTransactions)
		authed.POST("/wallet/deposit", a.handleDeposit)
		authed.POST("/wallet/withdraw", a.handleWithdraw)

		// Orders / trading (order POSTs are idempotent — EX-11)
		authed.POST("/orders", idem, a.handlePlaceOrder)
		authed.POST("/orders/quote", a.handleQuote)
		authed.POST("/orders/:quoteId/confirm", idem, a.handleConfirmQuote)
		authed.GET("/orders/:quoteId", a.handleGetOrder) // :quoteId reused as txnRef

		// DCA
		authed.GET("/dca-plans", a.handleListDcaPlans)
		authed.POST("/dca-plans", a.handleCreateDca)
		authed.PUT("/dca-plans/:id", a.handleUpdateDca)
		authed.POST("/dca-plans/:id/pause", a.handlePauseDca)
		authed.POST("/dca-plans/:id/resume", a.handleResumeDca)
		authed.DELETE("/dca-plans/:id", a.handleCancelDca)
		authed.GET("/dca-plans/:id/rounds", a.handleDcaRounds)

		// Portfolio & dashboard
		authed.GET("/portfolio", a.handlePortfolio)
		authed.GET("/dashboard", a.handleDashboard)

		// History
		authed.GET("/transactions", a.handleListTransactions)
		authed.GET("/transactions/export", a.handleExport)
		authed.GET("/transactions/:txnRef", a.handleTransactionDetail)
		authed.POST("/transactions/:txnRef/cancel", a.handleCancelTransaction)
		authed.GET("/statements/:yyyyMM", a.handleStatement)

		// Notifications & inbox
		authed.GET("/notifications", a.handleListNotifications)
		authed.POST("/notifications/:id/read", a.handleMarkNotificationRead)
		authed.GET("/inbox", a.handleInbox)
	}

	// --- Admin (RBAC admin) ---
	admin := api.Group("/admin")
	admin.Use(authMW, adminMW)
	{
		admin.POST("/run/dca", a.handleRunDca)
		admin.POST("/run/coupons", a.handleRunCoupons)
		admin.POST("/run/maturities", a.handleRunMaturities)

		admin.GET("/price-generator", a.handleGetPriceGenerator)
		admin.POST("/price-generator", a.handleSetPriceGenerator)
		admin.POST("/publish-nav", a.handlePublishNav)
		admin.POST("/tick-gold", a.handleTickGold)

		admin.POST("/funds", a.handleCreateFund)
		admin.POST("/bonds", a.handleCreateBond)
		admin.POST("/gold", a.handleCreateGold)
		admin.GET("/audit-logs", a.handleAuditLogs)
	}
}
