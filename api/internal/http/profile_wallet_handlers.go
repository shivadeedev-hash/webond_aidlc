package http

import (
	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/service"
)

// --- Profile & Risk (US-006/007) ---

func (a *API) handleGetProfile(c *gin.Context) {
	u, err := a.profile.Get(c.Request.Context(), userID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{
		"id": u.ID, "email": u.Email, "fullName": u.FullName, "mobile": u.Mobile.String,
		"bankAccount": u.BankAccountMasked.String, "status": u.Status,
		"emailNotificationsEnabled": u.EmailNotificationsEnabled,
	})
}

type updateProfileReq struct {
	FullName    string `json:"fullName"`
	Mobile      string `json:"mobile"`
	BankAccount string `json:"bankAccount"`
}

func (a *API) handleUpdateProfile(c *gin.Context) {
	var req updateProfileReq
	if !bindJSON(c, &req) {
		return
	}
	u, err := a.profile.Update(c.Request.Context(), userID(c), service.UpdateProfileInput{
		FullName: req.FullName, Mobile: req.Mobile, BankAccount: req.BankAccount,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{
		"id": u.ID, "fullName": u.FullName, "mobile": u.Mobile.String, "bankAccount": u.BankAccountMasked.String,
	})
}

func (a *API) handleGetRisk(c *gin.Context) {
	r, ok, err := a.profile.GetRisk(c.Request.Context(), userID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	if !ok {
		RespondData(c, nil)
		return
	}
	RespondData(c, gin.H{"riskLevel": r.RiskLevel, "assessedAt": r.AssessedAt, "expiresAt": r.ExpiresAt, "valid": r.Valid})
}

type riskReq struct {
	Answers   map[string]int `json:"answers"`
	RiskLevel int8           `json:"riskLevel"`
}

func (a *API) handleSubmitRisk(c *gin.Context) {
	var req riskReq
	if !bindJSON(c, &req) {
		return
	}
	r, err := a.profile.SubmitRisk(c.Request.Context(), userID(c), service.SubmitRiskInput{
		Answers: req.Answers, RiskLevel: req.RiskLevel,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"riskLevel": r.RiskLevel, "assessedAt": r.AssessedAt, "expiresAt": r.ExpiresAt})
}

type emailSettingsReq struct {
	EmailEnabled bool `json:"emailEnabled"`
}

func (a *API) handleNotificationSettings(c *gin.Context) {
	var req emailSettingsReq
	if !bindJSON(c, &req) {
		return
	}
	if err := a.profile.SetEmailNotifications(c.Request.Context(), userID(c), req.EmailEnabled); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"emailEnabled": req.EmailEnabled})
}

// --- Wallet (US-008/009/010) ---

func (a *API) handleGetWallet(c *gin.Context) {
	b, err := a.wallet.Get(c.Request.Context(), userID(c))
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, balancePayload(b))
}

type amountReq struct {
	Amount string `json:"amount"`
}

func (a *API) handleDeposit(c *gin.Context) {
	var req amountReq
	if !bindJSON(c, &req) {
		return
	}
	amt, ok := parseDecimal(req.Amount)
	if !ok {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid amount")
		return
	}
	b, err := a.wallet.Deposit(c.Request.Context(), userID(c), amt)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, balancePayload(b))
}

func (a *API) handleWithdraw(c *gin.Context) {
	var req amountReq
	if !bindJSON(c, &req) {
		return
	}
	amt, ok := parseDecimal(req.Amount)
	if !ok {
		RespondErrorCode(c, domain.CodeValidationError, "Invalid amount")
		return
	}
	b, err := a.wallet.Withdraw(c.Request.Context(), userID(c), amt)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, balancePayload(b))
}

func (a *API) handleWalletTransactions(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	rows, total, err := a.wallet.ListTransactions(c.Request.Context(), userID(c), page, pageSize)
	if err != nil {
		RespondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id": r.ID, "type": r.Type, "amount": money.FormatTHB(r.Amount),
			"balanceAfter": money.FormatTHB(r.BalanceAfter), "status": r.Status,
			"createdAt": r.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}
	RespondList(c, out, page, pageSize, total)
}

func balancePayload(b service.Balances) gin.H {
	return gin.H{
		"balance": money.FormatTHB(b.Balance), "reserved": money.FormatTHB(b.Reserved),
		"available": money.FormatTHB(b.Available),
	}
}
