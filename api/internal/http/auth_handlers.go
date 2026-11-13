package http

import (
	"github.com/gin-gonic/gin"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/service"
)

// registerReq mirrors POST /api/auth/register (US-001).
type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"fullName"`
	Mobile      string `json:"mobile"`
	DateOfBirth string `json:"dateOfBirth"`
	Consent     bool   `json:"consent"`
}

func (a *API) handleRegister(c *gin.Context) {
	var req registerReq
	if !bindJSON(c, &req) {
		return
	}
	res, err := a.auth.Register(c.Request.Context(), service.RegisterInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName,
		Mobile: req.Mobile, DateOfBirth: req.DateOfBirth, Consent: req.Consent,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	// The token proves ownership of the e-mail address, so it is delivered only
	// by e-mail. Outside production it is also echoed for local testing.
	body := gin.H{"userId": res.UserID}
	if a.cfg.ExposeDevTokens() {
		body["verifyCode"] = res.VerifyCode
	}
	RespondStatus(c, 201, body)
}

// otpReq is POST /api/auth/verify: the e-mail plus the 6-digit OTP (US-002).
type otpReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (a *API) handleVerify(c *gin.Context) {
	var req otpReq
	if !bindJSON(c, &req) {
		return
	}
	if err := a.auth.Verify(c.Request.Context(), req.Email, req.Code); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"status": "active"})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *API) handleLogin(c *gin.Context) {
	var req loginReq
	if !bindJSON(c, &req) {
		return
	}
	res, err := a.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		RespondError(c, err)
		return
	}
	a.setRefreshCookie(c, res.RefreshToken)
	RespondData(c, gin.H{
		"accessToken": res.AccessToken,
		"user":        userView(res.User),
	})
}

// userView is the public user shape shared by login and refresh responses.
func userView(u db.User) gin.H {
	return gin.H{
		"id": u.ID, "email": u.Email, "fullName": u.FullName,
		"role": u.Role, "status": u.Status,
	}
}

func (a *API) handleRefresh(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie == "" {
		RespondErrorCode(c, domain.CodeUnauthorized, "Missing refresh token")
		return
	}
	res, err := a.auth.Refresh(c.Request.Context(), cookie)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{
		"accessToken": res.AccessToken,
		"user":        userView(res.User),
	})
}

func (a *API) handleLogout(c *gin.Context) {
	a.clearRefreshCookie(c)
	c.Status(204)
}

type emailReq struct {
	Email string `json:"email"`
}

func (a *API) handleForgotPassword(c *gin.Context) {
	var req emailReq
	if !bindJSON(c, &req) {
		return
	}
	// Always 202 regardless of existence (no enumeration, US-004).
	code, _ := a.auth.ForgotPassword(c.Request.Context(), req.Email)
	// The reset code is a password-equivalent secret: it is delivered only by
	// e-mail. Echoing it (local development only) would also reveal whether the
	// account exists, which is acceptable only outside production.
	body := gin.H{"accepted": true}
	if a.cfg.ExposeDevTokens() {
		body["resetCode"] = code
	}
	RespondStatus(c, 202, body)
}

// handleResendVerification e-mails a new verification OTP (US-002). Always 202.
func (a *API) handleResendVerification(c *gin.Context) {
	var req emailReq
	if !bindJSON(c, &req) {
		return
	}
	code, _ := a.auth.ResendVerification(c.Request.Context(), req.Email)
	body := gin.H{"accepted": true}
	if a.cfg.ExposeDevTokens() {
		body["verifyCode"] = code
	}
	RespondStatus(c, 202, body)
}

type resetReq struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

func (a *API) handleResetPassword(c *gin.Context) {
	var req resetReq
	if !bindJSON(c, &req) {
		return
	}
	if err := a.auth.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		RespondError(c, err)
		return
	}
	RespondData(c, gin.H{"status": "ok"})
}
