package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/config"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/http/middleware"
	"github.com/webond/api/internal/service"
)

// API bundles all services and shared config for the Gin handlers. It is built
// once at startup and its RegisterRoutes method wires every endpoint onto the
// router (api-spec.md). Handlers reuse the envelope (response.go) and the
// phase-4 middleware (auth/RBAC/idempotency/CORS).
type API struct {
	cfg       *config.Config
	auth      *service.AuthService
	profile   *service.ProfileService
	wallet    *service.WalletService
	catalog   *service.CatalogService
	trading   *service.TradingService
	dca       *service.DCAService
	portfolio *service.PortfolioService
	dashboard *service.DashboardService
	history   *service.HistoryService
	notify    *service.NotificationService
	bond      *service.BondLifecycleService
	admin     *service.AdminService
	pricegen  *service.PriceGeneratorService
}

// APIServices groups the constructed services passed to NewAPI.
type APIServices struct {
	Auth      *service.AuthService
	Profile   *service.ProfileService
	Wallet    *service.WalletService
	Catalog   *service.CatalogService
	Trading   *service.TradingService
	DCA       *service.DCAService
	Portfolio *service.PortfolioService
	Dashboard *service.DashboardService
	History   *service.HistoryService
	Notify    *service.NotificationService
	Bond      *service.BondLifecycleService
	Admin     *service.AdminService
	PriceGen  *service.PriceGeneratorService
}

// NewAPI builds the API handler bundle.
func NewAPI(cfg *config.Config, svc APIServices) *API {
	return &API{
		cfg: cfg, auth: svc.Auth, profile: svc.Profile, wallet: svc.Wallet,
		catalog: svc.Catalog, trading: svc.Trading, dca: svc.DCA,
		portfolio: svc.Portfolio, dashboard: svc.Dashboard, history: svc.History,
		notify: svc.Notify, bond: svc.Bond, admin: svc.Admin, pricegen: svc.PriceGen,
	}
}

// --- shared request helpers ---

// userID returns the authenticated user id set by the Auth middleware.
func userID(c *gin.Context) uint64 {
	v, ok := c.Get(middleware.ContextUserID)
	if !ok {
		return 0
	}
	id, _ := v.(int64)
	return uint64(id)
}

// bindJSON binds the request body, responding with VALIDATION_ERROR on failure.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid request body")
		return false
	}
	return true
}

// parseDecimal parses a string (or empty) into a decimal; empty → zero.
func parseDecimal(s string) (decimal.Decimal, bool) {
	if s == "" {
		return decimal.Zero, true
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, false
	}
	return d, true
}

// queryInt reads an int query param with a default.
func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// pathUint reads a uint64 path param, responding 404 on parse failure.
func pathUint(c *gin.Context, key string) (uint64, bool) {
	v := c.Param(key)
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid "+key)
		return 0, false
	}
	return n, true
}

// setRefreshCookie writes the refresh token as an httpOnly secure cookie
// (api-spec.md § Auth). Secure is enabled only in production (dev is plain HTTP).
func (a *API) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", token, int(service.RefreshCookieMaxAge.Seconds()), "/api/auth", "", a.cfg.IsProduction(), true)
}

// clearRefreshCookie removes the refresh cookie on logout.
func (a *API) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", "", -1, "/api/auth", "", a.cfg.IsProduction(), true)
}
