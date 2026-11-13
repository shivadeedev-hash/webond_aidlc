package http

import (
	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/http/middleware"
	"github.com/webond/api/internal/service"
)

// orderReq mirrors POST /api/orders (US-013..018).
type orderReq struct {
	AssetType   string `json:"assetType"`
	AssetID     uint64 `json:"assetId"`
	Side        string `json:"side"`
	OrderMode   string `json:"orderMode"`
	Amount      string `json:"amount"`
	Units       string `json:"units"`
	WeightGrams string `json:"weightGrams"`
	RiskAck     bool   `json:"riskAck"`
}

func (a *API) handlePlaceOrder(c *gin.Context) {
	var req orderReq
	if !bindJSON(c, &req) {
		return
	}
	amount, ok1 := parseDecimal(req.Amount)
	units, ok2 := parseDecimal(req.Units)
	weight, ok3 := parseDecimal(req.WeightGrams)
	if !ok1 || !ok2 || !ok3 {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid numeric field")
		return
	}
	key, _ := c.Get(middleware.ContextIdempotencyKey)
	idemKey, _ := key.(string)
	res, err := a.trading.PlaceOrder(c.Request.Context(), service.OrderInput{
		UserID: userID(c), AssetType: req.AssetType, AssetID: req.AssetID, Side: req.Side,
		OrderMode: req.OrderMode, Amount: amount, Units: units, WeightGrams: weight,
		RiskAck: req.RiskAck, IdempotencyKey: idemKey,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, res)
}

// quoteReq mirrors POST /api/orders/quote (gold, US-017).
type quoteReq struct {
	AssetType   string `json:"assetType"`
	AssetID     uint64 `json:"assetId"`
	Side        string `json:"side"`
	Amount      string `json:"amount"`
	WeightGrams string `json:"weightGrams"`
}

func (a *API) handleQuote(c *gin.Context) {
	var req quoteReq
	if !bindJSON(c, &req) {
		return
	}
	amount, ok1 := parseDecimal(req.Amount)
	weight, ok2 := parseDecimal(req.WeightGrams)
	if !ok1 || !ok2 {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid numeric field")
		return
	}
	q, err := a.trading.Quote(c.Request.Context(), userID(c), req.AssetID, req.Side, amount, weight)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, q)
}

func (a *API) handleConfirmQuote(c *gin.Context) {
	quoteID := c.Param("quoteId")
	key, _ := c.Get(middleware.ContextIdempotencyKey)
	idemKey, _ := key.(string)
	res, err := a.trading.ConfirmQuote(c.Request.Context(), userID(c), quoteID, idemKey)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, res)
}

func (a *API) handleGetOrder(c *gin.Context) {
	// The route is /orders/:quoteId (gin needs one name per path segment, shared
	// with /orders/:quoteId/confirm); here the segment holds a txn reference.
	res, err := a.trading.GetOrder(c.Request.Context(), userID(c), c.Param("quoteId"))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, res)
}
