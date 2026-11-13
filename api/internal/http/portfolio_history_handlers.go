package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/service"
)

// --- Portfolio & Dashboard (US-022/023) ---

func (a *API) handlePortfolio(c *gin.Context) {
	pf, err := a.portfolio.Get(c.Request.Context(), userID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, pf)
}

func (a *API) handleDashboard(c *gin.Context) {
	d, err := a.dashboard.Get(c.Request.Context(), userID(c), c.Query("trend"))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, d)
}

// --- History (US-024..028) ---

func (a *API) historyFilter(c *gin.Context) service.HistoryFilter {
	return service.HistoryFilter{
		Tab: c.Query("tab"), AssetType: c.Query("assetType"), Type: c.Query("type"),
		Status: c.Query("status"), From: c.Query("from"), To: c.Query("to"),
		Query: c.Query("q"), Sort: c.Query("sort"), Order: c.Query("order"),
		Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 20),
	}
}

func (a *API) handleListTransactions(c *gin.Context) {
	f := a.historyFilter(c)
	rows, total, err := a.history.List(c.Request.Context(), userID(c), f)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondList(c, rows, f.Page, f.PageSize, total)
}

func (a *API) handleTransactionDetail(c *gin.Context) {
	d, err := a.history.Detail(c.Request.Context(), userID(c), c.Param("txnRef"))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, d)
}

func (a *API) handleCancelTransaction(c *gin.Context) {
	if err := a.history.Cancel(c.Request.Context(), userID(c), c.Param("txnRef")); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"status": "cancelled"})
}

func (a *API) handleExport(c *gin.Context) {
	f := a.historyFilter(c)
	data, filename, err := a.history.Export(c.Request.Context(), userID(c), f)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (a *API) handleStatement(c *gin.Context) {
	data, filename, err := a.history.MonthlyStatement(c.Request.Context(), userID(c), c.Param("yyyyMM"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/pdf", data)
}

// --- Notifications (US-029) ---

func (a *API) handleListNotifications(c *gin.Context) {
	page := queryInt(c, "page", 1)
	rows, _, err := a.notify.ListInApp(c.Request.Context(), userID(c), page, 20)
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id": r.ID, "category": r.Category, "subject": r.Subject.String, "body": r.Body.String,
			"readAt":    nullTimeStr(r.ReadAt.Valid, r.ReadAt.Time.Format("2006-01-02T15:04:05-07:00")),
			"createdAt": r.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}
	RespondData(c, out)
}

func (a *API) handleMarkNotificationRead(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := a.notify.MarkRead(c.Request.Context(), userID(c), id); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"id": id, "read": true})
}

func (a *API) handleInbox(c *gin.Context) {
	page := queryInt(c, "page", 1)
	rows, _, err := a.notify.ListInbox(c.Request.Context(), userID(c), page, 20)
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id": r.ID, "category": r.Category, "subject": r.Subject.String, "body": r.Body.String,
			"token": r.Token.String, "createdAt": r.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}
	RespondData(c, out)
}

func nullTimeStr(valid bool, s string) any {
	if !valid {
		return nil
	}
	return s
}
