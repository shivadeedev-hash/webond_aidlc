package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

// ProfileService implements profile maintenance and the risk-suitability
// assessment (components.md § ProfileService; FR-04, BR-01, US-006/007).
type ProfileService struct {
	store *repo.Store
	clock Clock
}

// NewProfileService builds a ProfileService.
func NewProfileService(store *repo.Store, clock Clock) *ProfileService {
	return &ProfileService{store: store, clock: clock}
}

// Get returns the user's profile row (US-006).
func (s *ProfileService) Get(ctx context.Context, userID uint64) (db.User, error) {
	u, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return db.User{}, domain.NewError(domain.CodeTxnNotFound, "User not found")
		}
		return db.User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// UpdateProfileInput is the profile edit payload (US-006).
type UpdateProfileInput struct {
	FullName    string
	Mobile      string
	BankAccount string // raw; stored masked (NFR-02)
}

// Update persists profile changes, masking the bank account on storage (US-006,
// NFR-02 on-screen masking). An empty bank account leaves the field cleared.
func (s *ProfileService) Update(ctx context.Context, userID uint64, in UpdateProfileInput) (db.User, error) {
	if strings.TrimSpace(in.FullName) == "" {
		return db.User{}, domain.NewError(domain.CodeValidationError, "Full name is required")
	}
	// The form never shows the full account number (only the masked value), so an
	// update that leaves the field blank means "unchanged" — it must not wipe the
	// account already on file.
	bank := nullString(maskBankAccount(strings.TrimSpace(in.BankAccount)))
	if strings.TrimSpace(in.BankAccount) == "" {
		current, err := s.store.Users.GetByID(ctx, userID)
		if err != nil {
			return db.User{}, fmt.Errorf("load profile: %w", err)
		}
		bank = current.BankAccountMasked
	}
	if err := s.store.Users.UpdateProfile(ctx, db.UpdateUserProfileParams{
		FullName:          strings.TrimSpace(in.FullName),
		Mobile:            nullString(strings.TrimSpace(in.Mobile)),
		BankAccountMasked: bank,
		ID:                userID,
	}); err != nil {
		return db.User{}, fmt.Errorf("update profile: %w", err)
	}
	return s.Get(ctx, userID)
}

// RiskResult is the latest assessment view (US-007).
type RiskResult struct {
	RiskLevel  int8
	AssessedAt string
	ExpiresAt  string
	Valid      bool
}

// GetRisk returns the latest risk assessment (or Valid=false when absent).
func (s *ProfileService) GetRisk(ctx context.Context, userID uint64) (RiskResult, bool, error) {
	r, err := s.store.Risk.GetLatest(ctx, userID)
	if err != nil {
		if isNoRows(err) {
			return RiskResult{}, false, nil
		}
		return RiskResult{}, false, fmt.Errorf("get risk: %w", err)
	}
	return RiskResult{
		RiskLevel:  r.RiskLevel,
		AssessedAt: r.AssessedAt.Format("2006-01-02T15:04:05-07:00"),
		ExpiresAt:  r.ExpiresAt.Format("2006-01-02T15:04:05-07:00"),
		Valid:      s.clock.Now().Before(r.ExpiresAt),
	}, true, nil
}

// SubmitRiskInput carries the questionnaire answers and the computed level.
// The caller may pass an explicit RiskLevel (1–5); if 0, it is derived from the
// answers by simple summation bucketing.
type SubmitRiskInput struct {
	Answers   map[string]int
	RiskLevel int8
}

// SubmitRisk stores a new risk assessment with a 2-year expiry (US-007). The
// latest row governs trading eligibility (BR-01) and risk-mismatch (BR-08).
func (s *ProfileService) SubmitRisk(ctx context.Context, userID uint64, in SubmitRiskInput) (RiskResult, error) {
	level := in.RiskLevel
	if level == 0 {
		level = deriveRiskLevel(in.Answers)
	}
	if level < 1 || level > 5 {
		return RiskResult{}, domain.NewError(domain.CodeValidationError, "Risk level must be between 1 and 5")
	}
	answers, _ := json.Marshal(in.Answers)
	if len(answers) == 0 {
		answers = []byte(`{}`)
	}
	now := s.clock.Now()
	expires := now.Add(riskExpiryTTL)
	if _, err := s.store.Risk.Create(ctx, db.CreateRiskAssessmentParams{
		UserID:      userID,
		RiskLevel:   level,
		AnswersJson: answers,
		AssessedAt:  now,
		ExpiresAt:   expires,
	}); err != nil {
		return RiskResult{}, fmt.Errorf("create risk: %w", err)
	}
	return RiskResult{
		RiskLevel:  level,
		AssessedAt: now.Format("2006-01-02T15:04:05-07:00"),
		ExpiresAt:  expires.Format("2006-01-02T15:04:05-07:00"),
		Valid:      true,
	}, nil
}

// deriveRiskLevel buckets the summed answer scores into 1–5. With no answers it
// defaults to the most conservative level (1).
func deriveRiskLevel(answers map[string]int) int8 {
	if len(answers) == 0 {
		return 1
	}
	total := 0
	for _, v := range answers {
		total += v
	}
	avg := total / len(answers)
	switch {
	case avg <= 1:
		return 1
	case avg == 2:
		return 2
	case avg == 3:
		return 3
	case avg == 4:
		return 4
	default:
		return 5
	}
}

// SetEmailNotifications updates the user's email-notification preference (US-029).
func (s *ProfileService) SetEmailNotifications(ctx context.Context, userID uint64, enabled bool) error {
	if err := s.store.Users.SetEmailNotifications(ctx, db.SetEmailNotificationsParams{
		EmailNotificationsEnabled: enabled, ID: userID,
	}); err != nil {
		return fmt.Errorf("set email notifications: %w", err)
	}
	return nil
}
