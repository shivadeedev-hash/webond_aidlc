package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/webond/api/internal/auth"
	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

// Auth lockout policy (EX-01, US-003): 5 consecutive failed logins lock the
// account for 15 minutes.
const (
	maxFailedLogins = 5
	lockoutDuration = 15 * time.Minute
	// One-time codes e-mailed for verification and password reset (US-002/004,
	// Add-on 2026-11-12b): valid 15 minutes, 5 wrong attempts, resend every 60 s.
	otpTTL            = 15 * time.Minute
	otpMaxAttempts    = 5
	otpResendCooldown = 60 * time.Second
	riskExpiryTTL     = 2 * 365 * 24 * time.Hour // risk assessment valid 2 years (US-007)
)

// AuthService implements registration, email verification, login/logout,
// refresh and password reset (components.md § AuthService; api-spec.md § Auth).
// Passwords are bcrypt-hashed (cost 12); sessions are JWT access/refresh.
type AuthService struct {
	store  *repo.Store
	tokens *auth.Manager
	notify *NotificationService
	clock  Clock
}

// RefreshCookieMaxAge is the refresh-token cookie lifetime (matches the token
// TTL, 7 days) used by the HTTP layer when setting the httpOnly cookie.
const RefreshCookieMaxAge = auth.RefreshTokenTTL

// NewAuthService builds an AuthService.
func NewAuthService(store *repo.Store, tokens *auth.Manager, notify *NotificationService, clock Clock) *AuthService {
	return &AuthService{store: store, tokens: tokens, notify: notify, clock: clock}
}

// RegisterInput is the register request (US-001).
type RegisterInput struct {
	Email       string
	Password    string
	FullName    string
	Mobile      string
	DateOfBirth string // ISO date yyyy-MM-dd (optional)
	Consent     bool
}

// RegisterResult is returned on a successful registration.
type RegisterResult struct {
	UserID uint64
	// VerifyCode is the e-mailed OTP, returned so tests and local development
	// can verify without reading the mail. The HTTP layer never echoes it in
	// production.
	VerifyCode string
}

// Register creates a pending_verification account (US-001): validates fields,
// enforces email uniqueness (EX-02 → DUPLICATE_EMAIL), hashes the password with
// bcrypt, creates the user + wallet in one transaction, and writes a
// verification token to the outbox.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (RegisterResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || !strings.Contains(email, "@") {
		return RegisterResult{}, domain.NewError(domain.CodeValidationError, "A valid email is required")
	}
	if len(in.Password) < 8 {
		return RegisterResult{}, domain.NewError(domain.CodeValidationError, "Password must be at least 8 characters")
	}
	if strings.TrimSpace(in.FullName) == "" {
		return RegisterResult{}, domain.NewError(domain.CodeValidationError, "Full name is required")
	}
	if !in.Consent {
		return RegisterResult{}, domain.NewError(domain.CodeValidationError, "You must accept the terms and PDPA consent")
	}

	// Uniqueness check (EX-02).
	if _, err := s.store.Users.GetByEmail(ctx, email); err == nil {
		return RegisterResult{}, domain.NewError(domain.CodeDuplicateEmail, "This email is already registered")
	} else if !isNoRows(err) {
		return RegisterResult{}, fmt.Errorf("lookup email: %w", err)
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hash password: %w", err)
	}

	var dob sql.NullTime
	if in.DateOfBirth != "" {
		if t, perr := time.ParseInLocation("2006-01-02", in.DateOfBirth, Bangkok); perr == nil {
			dob = nullTime(t)
		}
	}

	var userID uint64
	var code string
	// READ COMMITTED: the one-time-code bookkeeping must not take gap locks
	// that deadlock with other people registering at the same moment.
	err = s.store.WithinTxReadCommitted(ctx, func(tx *repo.Store) error {
		res, cerr := tx.Users.Create(ctx, db.CreateUserParams{
			Email:         email,
			PasswordHash:  hash,
			FullName:      strings.TrimSpace(in.FullName),
			Mobile:        nullString(strings.TrimSpace(in.Mobile)),
			DateOfBirth:   dob,
			Role:          db.UsersRoleInvestor,
			Status:        db.UsersStatusPendingVerification,
			PdpaConsentAt: nullTime(s.clock.Now()),
		})
		id, cerr := lastInsertID(res, cerr)
		if cerr != nil {
			return fmt.Errorf("create user: %w", cerr)
		}
		userID = id
		if _, cerr := tx.Wallets.Create(ctx, userID); cerr != nil {
			return fmt.Errorf("create wallet: %w", cerr)
		}
		// Verification OTP e-mail (US-001/002, FR-01).
		var oerr error
		code, oerr = s._issueOTP(ctx, tx, userID, db.NotificationOutboxCategoryVerify)
		return oerr
	})
	if err != nil {
		return RegisterResult{}, err
	}

	s.audit(ctx, userID, "register", "user", int64(userID), nil)
	return RegisterResult{UserID: userID, VerifyCode: code}, nil
}

// Verify activates a pending account with the e-mailed OTP (US-002). Unknown
// e-mails get the same generic error as a wrong code.
func (s *AuthService) Verify(ctx context.Context, email, code string) error {
	user, err := s._userForOTP(ctx, email, code)
	if err != nil {
		return err
	}
	if user.Status == db.UsersStatusActive {
		return nil // already verified: idempotent
	}
	if user.Status != db.UsersStatusPendingVerification {
		return errInvalidOTP()
	}
	if err := s._checkOTP(ctx, user.ID, db.NotificationOutboxCategoryVerify, code, func(tx *repo.Store) error {
		return tx.Users.SetStatus(ctx, db.SetUserStatusParams{Status: db.UsersStatusActive, ID: user.ID})
	}); err != nil {
		return err
	}
	s.audit(ctx, user.ID, "verify_email", "user", int64(user.ID), nil)
	return nil
}

// ResendVerification e-mails a new verification OTP to a pending account
// (US-002). It never reveals whether the e-mail exists and sends at most one
// code per otpResendCooldown. The code is returned for local development only.
func (s *AuthService) ResendVerification(ctx context.Context, email string) (string, error) {
	user, err := s.store.Users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if isNoRows(err) {
			return "", nil
		}
		return "", fmt.Errorf("lookup user: %w", err)
	}
	if user.Status != db.UsersStatusPendingVerification {
		return "", nil
	}
	return s._issueOTPWithCooldown(ctx, user.ID, db.NotificationOutboxCategoryVerify)
}

// LoginResult carries the issued tokens and the authenticated user (US-003).
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         db.User
}

// Login authenticates email+password (US-003). On success it resets the failed
// counter and issues JWT access (15m) + refresh (7d). On failure it increments
// the failed counter and locks the account after 5 fails (EX-01). All failure
// paths return the same generic UNAUTHORIZED error (no enumeration).
func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	generic := domain.NewError(domain.CodeUnauthorized, "Invalid email or password")
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.store.Users.GetByEmail(ctx, email)
	if err != nil {
		if isNoRows(err) {
			return LoginResult{}, generic
		}
		return LoginResult{}, fmt.Errorf("lookup user: %w", err)
	}

	now := s.clock.Now()
	// Locked account: reject until lockout expires (EX-01).
	if user.Status == db.UsersStatusLocked {
		if user.LockedUntil.Valid && now.Before(user.LockedUntil.Time) {
			return LoginResult{}, domain.NewError(domain.CodeUnauthorized,
				"Account locked due to repeated failed logins. Try again later.")
		}
		// Lockout window elapsed: clear it and let the password check proceed.
		_ = s.store.Users.ResetFailedLogin(ctx, user.ID)
		_ = s.store.Users.SetStatus(ctx, db.SetUserStatusParams{Status: db.UsersStatusActive, ID: user.ID})
		user.Status = db.UsersStatusActive
		user.FailedLoginCount = 0
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		s.registerFailedLogin(ctx, user)
		return LoginResult{}, generic
	}

	// Must be verified to log in fully (BR-01). Pending accounts are rejected
	// with a clear (non-enumerating for password) message.
	if user.Status == db.UsersStatusPendingVerification {
		return LoginResult{}, domain.NewError(domain.CodeForbidden, "Please verify your email before signing in")
	}

	// Success: clear failure state.
	if user.FailedLoginCount > 0 {
		_ = s.store.Users.ResetFailedLogin(ctx, user.ID)
	}

	access, err := s.tokens.IssueAccessToken(int64(user.ID), auth.Role(user.Role))
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.tokens.IssueRefreshToken(int64(user.ID), auth.Role(user.Role))
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue refresh token: %w", err)
	}
	s.audit(ctx, user.ID, "login", "user", int64(user.ID), nil)
	return LoginResult{AccessToken: access, RefreshToken: refresh, User: user}, nil
}

// registerFailedLogin increments the failed counter and locks the account when
// the threshold is reached (EX-01).
func (s *AuthService) registerFailedLogin(ctx context.Context, user db.User) {
	_ = s.store.Users.IncrementFailedLogin(ctx, user.ID)
	if int(user.FailedLoginCount)+1 >= maxFailedLogins {
		_ = s.store.Users.Lock(ctx, db.LockUserParams{
			LockedUntil: nullTime(s.clock.Now().Add(lockoutDuration)),
			ID:          user.ID,
		})
		s.audit(ctx, user.ID, "account_locked", "user", int64(user.ID), nil)
	}
}

// RefreshResult is a rotated access token plus the user it belongs to, so the
// SPA can restore its session after a page reload (D3-4 silent refresh).
type RefreshResult struct {
	AccessToken string
	User        db.User
}

// Refresh rotates an access token from a valid refresh token (EX-03).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	claims, err := s.tokens.VerifyRefreshToken(refreshToken)
	if err != nil {
		return RefreshResult{}, domain.NewError(domain.CodeUnauthorized, "Invalid or expired session")
	}
	uid, err := claims.UserID()
	if err != nil {
		return RefreshResult{}, domain.NewError(domain.CodeUnauthorized, "Invalid session")
	}
	user, err := s.store.Users.GetByID(ctx, uint64(uid))
	if err != nil {
		return RefreshResult{}, domain.NewError(domain.CodeUnauthorized, "Invalid session")
	}
	if user.Status != db.UsersStatusActive {
		return RefreshResult{}, domain.NewError(domain.CodeUnauthorized, "Account is not active")
	}
	access, err := s.tokens.IssueAccessToken(int64(user.ID), auth.Role(user.Role))
	if err != nil {
		return RefreshResult{}, fmt.Errorf("issue access token: %w", err)
	}
	return RefreshResult{AccessToken: access, User: user}, nil
}

// ForgotPassword always succeeds without revealing whether the email exists
// (US-004 — no enumeration). When it exists a 6-digit reset OTP (15 minutes,
// one use) is e-mailed, at most once per otpResendCooldown. The code is
// returned for local development only.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.store.Users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if isNoRows(err) {
			return "", nil // silently succeed (no enumeration)
		}
		return "", fmt.Errorf("lookup user: %w", err)
	}
	return s._issueOTPWithCooldown(ctx, user.ID, db.NotificationOutboxCategoryReset)
}

// ResetPassword sets a new password with the e-mailed reset OTP (US-004).
// Refresh tokens already issued stay valid until expiry (phase-1 limitation).
func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if len(newPassword) < 8 {
		return domain.NewError(domain.CodeValidationError, "Password must be at least 8 characters")
	}
	user, err := s._userForOTP(ctx, email, code)
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s._checkOTP(ctx, user.ID, db.NotificationOutboxCategoryReset, code, func(tx *repo.Store) error {
		return tx.Users.UpdatePassword(ctx, db.UpdateUserPasswordParams{PasswordHash: hash, ID: user.ID})
	}); err != nil {
		return err
	}
	s.audit(ctx, user.ID, "reset_password", "user", int64(user.ID), nil)
	return nil
}

// audit writes a best-effort audit log entry (NFR-02). Failures are swallowed
// so auditing never blocks the primary flow, but are logged by the repo layer.
func (s *AuthService) audit(ctx context.Context, userID uint64, action, entity string, entityID int64, detail []byte) {
	_, _ = s.store.Audit.Create(ctx, db.CreateAuditLogParams{
		UserID:     nullInt64(int64(userID)),
		Action:     action,
		Entity:     nullString(entity),
		EntityID:   nullInt64(entityID),
		DetailJson: detail,
	})
}

// ---- one-time codes (Add-on 2026-11-12b) ----

func errInvalidOTP() *domain.AppError {
	return domain.NewError(domain.CodeValidationError, "Invalid or expired code. Request a new code.")
}

// _userForOTP validates the code format and loads the user; unknown e-mails
// return the same error as a wrong code.
func (s *AuthService) _userForOTP(ctx context.Context, email, code string) (db.User, error) {
	if !auth.IsOTPFormat(strings.TrimSpace(code)) {
		return db.User{}, domain.NewError(domain.CodeValidationError, "Enter the 6-digit code from the e-mail")
	}
	user, err := s.store.Users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if isNoRows(err) {
			return db.User{}, errInvalidOTP()
		}
		return db.User{}, fmt.Errorf("lookup user: %w", err)
	}
	return user, nil
}

// _issueOTP closes older codes for the purpose, creates a new one and writes
// the e-mail to the outbox inside tx. Only an HMAC of the code is stored in
// `token`; the plain code is in the e-mail body until it has been sent.
func (s *AuthService) _issueOTP(ctx context.Context, tx *repo.Store, userID uint64, category db.NotificationOutboxCategory) (string, error) {
	now := s.clock.Now()
	if err := tx.Notifications.ConsumeOTPs(ctx, db.ConsumeOTPsParams{
		ReadAt: nullTime(now), UserID: nullInt64(int64(userID)), Category: category,
	}); err != nil {
		return "", fmt.Errorf("close old codes: %w", err)
	}
	code, err := auth.NewOTP()
	if err != nil {
		return "", err
	}
	subject, body := otpEmail(category, code)
	if err := s.notify.EmitTx(ctx, tx, Notice{
		UserID:      userID,
		Category:    category,
		Subject:     subject,
		Body:        body,
		Token:       auth.HashOTP(s.tokens.OTPKey(), userID, string(category), code),
		TokenExpiry: nullTime(now.Add(otpTTL)),
		EmailOnly:   true,
	}); err != nil {
		return "", err
	}
	return code, nil
}

// _issueOTPWithCooldown issues a new code unless one was issued less than
// otpResendCooldown ago (then it silently does nothing).
func (s *AuthService) _issueOTPWithCooldown(ctx context.Context, userID uint64, category db.NotificationOutboxCategory) (string, error) {
	var code string
	err := s.store.WithinTxReadCommitted(ctx, func(tx *repo.Store) error {
		// One request per user at a time (two quick "resend" clicks).
		if err := tx.Users.LockRow(ctx, userID); err != nil {
			return fmt.Errorf("lock user: %w", err)
		}
		latest, err := tx.Notifications.GetLatestOTP(ctx, db.GetLatestOTPParams{
			UserID: nullInt64(int64(userID)), Category: category,
		})
		if err == nil && latest.TokenExpiresAt.Valid {
			issued := latest.TokenExpiresAt.Time.Add(-otpTTL)
			if s.clock.Now().Before(issued.Add(otpResendCooldown)) {
				return nil
			}
		} else if err != nil && !isNoRows(err) {
			return fmt.Errorf("lookup code: %w", err)
		}
		var ierr error
		code, ierr = s._issueOTP(ctx, tx, userID, category)
		return ierr
	})
	return code, err
}

// _checkOTP verifies code against the newest open code for the purpose. A
// wrong code counts an attempt (committed even though the call fails); after
// otpMaxAttempts the code is dead. On success all open codes are consumed and
// onSuccess runs in the same transaction.
func (s *AuthService) _checkOTP(ctx context.Context, userID uint64, category db.NotificationOutboxCategory,
	code string, onSuccess func(tx *repo.Store) error) error {
	var result error
	err := s.store.WithinTxReadCommitted(ctx, func(tx *repo.Store) error {
		// Serialise attempts per user so concurrent guesses are all counted.
		if err := tx.Users.LockRow(ctx, userID); err != nil {
			return fmt.Errorf("lock user: %w", err)
		}
		row, err := tx.Notifications.GetLatestOTP(ctx, db.GetLatestOTPParams{
			UserID: nullInt64(int64(userID)), Category: category,
		})
		if err != nil {
			if isNoRows(err) {
				result = errInvalidOTP()
				return nil
			}
			return fmt.Errorf("lookup code: %w", err)
		}
		if !row.TokenExpiresAt.Valid || s.clock.Now().After(row.TokenExpiresAt.Time) {
			result = domain.NewError(domain.CodeValidationError, "This code has expired. Request a new code.")
			return nil
		}
		if row.TokenAttempts >= otpMaxAttempts {
			result = domain.NewError(domain.CodeValidationError, "Too many wrong attempts. Request a new code.")
			return nil
		}
		if !auth.OTPMatches(s.tokens.OTPKey(), userID, string(category), strings.TrimSpace(code), row.Token.String) {
			if err := tx.Notifications.IncrementOTPAttempts(ctx, row.ID); err != nil {
				return fmt.Errorf("count attempt: %w", err)
			}
			left := otpMaxAttempts - int(row.TokenAttempts) - 1
			if left <= 0 {
				result = domain.NewError(domain.CodeValidationError, "Too many wrong attempts. Request a new code.")
			} else {
				result = domain.NewError(domain.CodeValidationError,
					fmt.Sprintf("Incorrect code. %d attempt(s) left.", left))
			}
			return nil
		}
		if err := tx.Notifications.ConsumeOTPs(ctx, db.ConsumeOTPsParams{
			ReadAt: nullTime(s.clock.Now()), UserID: nullInt64(int64(userID)), Category: category,
		}); err != nil {
			return fmt.Errorf("consume code: %w", err)
		}
		return onSuccess(tx)
	})
	if err != nil {
		return err
	}
	return result
}

// otpEmail is the bilingual (Thai first, the app default) subject and body.
func otpEmail(category db.NotificationOutboxCategory, code string) (subject, body string) {
	if category == db.NotificationOutboxCategoryReset {
		return "WeBond: รหัสตั้งรหัสผ่านใหม่ / Password reset code",
			"รหัสสำหรับตั้งรหัสผ่านใหม่ของคุณคือ\n\n    " + code + "\n\n" +
				"กรอกรหัสนี้ในหน้า \"ตั้งรหัสผ่านใหม่\" ของ WeBond รหัสใช้ได้ 15 นาทีและใช้ได้ครั้งเดียว\n" +
				"หากคุณไม่ได้ขอเปลี่ยนรหัสผ่าน ไม่ต้องทำอะไร รหัสผ่านเดิมยังใช้ได้\n\n" +
				"---\n" +
				"Your WeBond password reset code is " + code + ".\n" +
				"Enter it on the WeBond reset password page. It expires in 15 minutes and works once.\n" +
				"If you did not ask for this, ignore this e-mail; your password stays the same.\n"
	}
	return "WeBond: รหัสยืนยันอีเมล / Verification code",
		"รหัสยืนยันอีเมลของคุณคือ\n\n    " + code + "\n\n" +
			"กรอกรหัสนี้ในหน้า \"ยืนยันอีเมล\" ของ WeBond รหัสใช้ได้ 15 นาทีและใช้ได้ครั้งเดียว\n" +
			"หากคุณไม่ได้สมัคร WeBond ไม่ต้องทำอะไร\n\n" +
			"---\n" +
			"Your WeBond verification code is " + code + ".\n" +
			"Enter it on the WeBond verification page. It expires in 15 minutes and works once.\n" +
			"If you did not sign up for WeBond, ignore this e-mail.\n"
}
