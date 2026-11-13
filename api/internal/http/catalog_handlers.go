package http

import (
	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/service"
)

// --- Market / Assets (US-011/012). Public (guest-accessible). ---

func (a *API) handleListAssets(c *gin.Context) {
	items, err := a.catalog.List(c.Request.Context(), service.ListParams{
		Type:      c.Query("type"),
		Query:     c.Query("q"),
		RiskLevel: queryInt(c, "riskLevel", 0),
		Category:  c.Query("category"),
		Period:    c.Query("period"),
		Sort:      c.Query("sort"),
		Order:     c.Query("order"),
		Page:      queryInt(c, "page", 1),
		PageSize:  queryInt(c, "pageSize", 50),
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, items)
}

func (a *API) handleAssetDetail(c *gin.Context) {
	assetType := c.Param("type")
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	d, err := a.catalog.Detail(c.Request.Context(), assetType, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, d)
}

func (a *API) handleAssetPrices(c *gin.Context) {
	assetType := c.Param("type")
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	points, err := a.catalog.PriceHistory(c.Request.Context(), assetType, id, c.Query("period"))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, points)
}
