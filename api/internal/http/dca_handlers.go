package http

import (
	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/service"
)

func dcaPlanPayload(p db.DcaPlan) gin.H {
	h := gin.H{
		"id": p.ID, "assetType": p.AssetType, "assetId": p.AssetID,
		"amountPerRound": money.FormatTHB(p.AmountPerRound), "frequency": p.Frequency,
		"runDay": p.RunDay, "status": p.Status, "roundsExecuted": p.RoundsExecuted,
		"totalInvested": money.FormatTHB(p.TotalInvested),
		"startDate":     p.StartDate.Format("2006-01-02"),
	}
	if p.NextRunDate.Valid {
		h["nextRunDate"] = p.NextRunDate.Time.Format("2006-01-02")
	}
	return h
}

func (a *API) handleListDcaPlans(c *gin.Context) {
	plans, err := a.dca.ListPlans(c.Request.Context(), userID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(plans))
	for _, p := range plans {
		out = append(out, dcaPlanPayload(p))
	}
	RespondData(c, out)
}

type createDcaReq struct {
	AssetType      string `json:"assetType"`
	AssetID        uint64 `json:"assetId"`
	AmountPerRound string `json:"amountPerRound"`
	Frequency      string `json:"frequency"`
	RunDay         int    `json:"runDay"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	MaxRounds      int    `json:"maxRounds"`
}

func (a *API) handleCreateDca(c *gin.Context) {
	var req createDcaReq
	if !bindJSON(c, &req) {
		return
	}
	amount, ok := parseDecimal(req.AmountPerRound)
	if !ok {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid amount")
		return
	}
	p, err := a.dca.CreatePlan(c.Request.Context(), service.CreatePlanInput{
		UserID: userID(c), AssetType: req.AssetType, AssetID: req.AssetID,
		AmountPerRound: amount, Frequency: req.Frequency, RunDay: req.RunDay,
		StartDate: req.StartDate, EndDate: req.EndDate, MaxRounds: req.MaxRounds,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondStatus(c, 201, dcaPlanPayload(p))
}

type updateDcaReq struct {
	AmountPerRound string `json:"amountPerRound"`
	RunDay         int    `json:"runDay"`
}

func (a *API) handleUpdateDca(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req updateDcaReq
	if !bindJSON(c, &req) {
		return
	}
	amount, aok := parseDecimal(req.AmountPerRound)
	if !aok {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid amount")
		return
	}
	p, err := a.dca.UpdatePlan(c.Request.Context(), userID(c), id, amount, req.RunDay)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, dcaPlanPayload(p))
}

func (a *API) handlePauseDca(c *gin.Context)  { a.setDcaStatus(c, db.DcaPlansStatusPaused) }
func (a *API) handleResumeDca(c *gin.Context) { a.setDcaStatus(c, db.DcaPlansStatusActive) }
func (a *API) handleCancelDca(c *gin.Context) { a.setDcaStatus(c, db.DcaPlansStatusCancelled) }

func (a *API) setDcaStatus(c *gin.Context, status db.DcaPlansStatus) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := a.dca.SetStatus(c.Request.Context(), userID(c), id, status); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"id": id, "status": status})
}

func (a *API) handleDcaRounds(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	rounds, err := a.dca.ListRounds(c.Request.Context(), userID(c), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rounds))
	for _, r := range rounds {
		out = append(out, gin.H{
			"id": r.ID, "runDate": r.RunDate.Format("2006-01-02"), "status": r.Status,
			"failureReason": r.FailureReason.String,
		})
	}
	RespondData(c, out)
}
