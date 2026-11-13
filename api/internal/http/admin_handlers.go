package http

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/service"
)

// decimalHalf is the default volatility used by manual publish/tick endpoints.
func decimalHalf() decimal.Decimal { return decimal.RequireFromString("0.5") }

// --- Admin run triggers (US-021/023, D1-3) ---

func (a *API) handleRunDca(c *gin.Context) {
	n, err := a.dca.RunDue(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"plansProcessed": n})
}

func (a *API) handleRunCoupons(c *gin.Context) {
	n, err := a.bond.RunCoupons(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"couponsProcessed": n})
}

func (a *API) handleRunMaturities(c *gin.Context) {
	n, err := a.bond.RunMaturities(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"maturitiesProcessed": n})
}

// --- Price generator (US-030 Add-on) ---

func (a *API) handleGetPriceGenerator(c *gin.Context) {
	st, err := a.pricegen.GetState(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, st)
}

type priceGenReq struct {
	Enabled         bool   `json:"enabled"`
	IntervalSeconds int    `json:"intervalSeconds"`
	VolatilityPct   string `json:"volatilityPct"`
}

func (a *API) handleSetPriceGenerator(c *gin.Context) {
	var req priceGenReq
	if !bindJSON(c, &req) {
		return
	}
	vol, ok := parseDecimal(req.VolatilityPct)
	if !ok {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid volatilityPct")
		return
	}
	st, err := a.pricegen.SetState(c.Request.Context(), req.Enabled, req.IntervalSeconds, vol)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, st)
}

func (a *API) handlePublishNav(c *gin.Context) {
	if err := a.pricegen.PublishNav(c.Request.Context(), decimalHalf()); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"status": "ok"})
}

func (a *API) handleTickGold(c *gin.Context) {
	if err := a.pricegen.TickGold(c.Request.Context(), decimalHalf()); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"status": "ok"})
}

// --- Admin master data (US-031) ---

type createFundReq struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	RiskLevel   int16  `json:"riskLevel"`
	MinPurchase string `json:"minPurchase"`
}

func (a *API) handleCreateFund(c *gin.Context) {
	var req createFundReq
	if !bindJSON(c, &req) {
		return
	}
	min, _ := parseDecimal(req.MinPurchase)
	id, err := a.admin.CreateFund(c.Request.Context(), userID(c), service.CreateFundInput{
		Code: req.Code, Name: req.Name, Category: req.Category, RiskLevel: req.RiskLevel, MinPurchase: min,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, gin.H{"id": id})
}

type createBondReq struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Issuer          string `json:"issuer"`
	CouponRate      string `json:"couponRate"`
	CouponFrequency int8   `json:"couponFrequency"`
	FaceValue       string `json:"faceValue"`
	RiskLevel       int16  `json:"riskLevel"`
}

func (a *API) handleCreateBond(c *gin.Context) {
	var req createBondReq
	if !bindJSON(c, &req) {
		return
	}
	rate, _ := parseDecimal(req.CouponRate)
	face, _ := parseDecimal(req.FaceValue)
	id, err := a.admin.CreateBond(c.Request.Context(), userID(c), service.CreateBondInput{
		Code: req.Code, Name: req.Name, Issuer: req.Issuer, CouponRate: rate,
		CouponFrequency: req.CouponFrequency, FaceValue: face, RiskLevel: req.RiskLevel,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, gin.H{"id": id})
}

type createGoldReq struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Purity string `json:"purity"`
}

func (a *API) handleCreateGold(c *gin.Context) {
	var req createGoldReq
	if !bindJSON(c, &req) {
		return
	}
	id, err := a.admin.CreateGold(c.Request.Context(), userID(c), service.CreateGoldInput{
		Code: req.Code, Name: req.Name, Purity: req.Purity,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, gin.H{"id": id})
}

func (a *API) handleAuditLogs(c *gin.Context) {
	page := queryInt(c, "page", 1)
	rows, err := a.admin.ListAuditLogs(c.Request.Context(), userID(c), page, 50)
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id": r.ID, "action": r.Action, "entity": r.Entity.String, "entityId": r.EntityID.Int64,
			"createdAt": r.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}
	RespondData(c, out)
}
