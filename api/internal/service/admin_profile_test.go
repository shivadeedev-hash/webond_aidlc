package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/domain"
)

// TestAdmin_CreateAssetsAndAudit covers CreateFund/CreateBond/CreateGold plus
// the audit-log write + ListAuditLogs (admin.go; US-031, NFR-02).
func TestAdmin_CreateAssetsAndAudit(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	as := NewAdminService(store, clock)
	ctx := context.Background()

	adminID := seedActiveUser(t, store, clock, decimal.Zero, 0)
	sfx := shortID()
	t.Cleanup(func() { cleanupUser(t, pool, adminID) })

	fundID, err := as.CreateFund(ctx, adminID, CreateFundInput{
		Code: "AF" + sfx, Name: "Admin Fund", Category: "Equity", RiskLevel: 4,
		MinPurchase: decimal.RequireFromString("1000"),
	})
	if err != nil {
		t.Fatalf("create fund: %v", err)
	}
	bondID, err := as.CreateBond(ctx, adminID, CreateBondInput{
		Code: "AB" + sfx, Name: "Admin Bond", Issuer: "Gov",
		CouponRate: decimal.RequireFromString("0.03"), CouponFrequency: 2,
		FaceValue: decimal.RequireFromString("1000"), RiskLevel: 3,
	})
	if err != nil {
		t.Fatalf("create bond: %v", err)
	}
	goldID, err := as.CreateGold(ctx, adminID, CreateGoldInput{Code: "AG" + sfx, Name: "Admin Gold", Purity: "96.5%"})
	if err != nil {
		t.Fatalf("create gold: %v", err)
	}
	t.Cleanup(func() {
		cleanupAsset(t, pool, "fund", fundID)
		cleanupAsset(t, pool, "bond", bondID)
		cleanupAsset(t, pool, "gold", goldID)
	})

	// Validation: missing code/name rejected.
	if _, err := as.CreateFund(ctx, adminID, CreateFundInput{Name: "No Code"}); err == nil {
		t.Fatal("expected validation error for missing fund code")
	}
	if _, err := as.CreateBond(ctx, adminID, CreateBondInput{Name: "No Code"}); err == nil {
		t.Fatal("expected validation error for missing bond code")
	}
	if _, err := as.CreateGold(ctx, adminID, CreateGoldInput{Name: "No Code"}); err == nil {
		t.Fatal("expected validation error for missing gold code")
	}

	// Audit log has the three create actions for this admin.
	logs, err := as.ListAuditLogs(ctx, adminID, 1, 50)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	actions := map[string]bool{}
	for _, l := range logs {
		actions[l.Action] = true
	}
	for _, want := range []string{"create_fund", "create_bond", "create_gold"} {
		if !actions[want] {
			t.Fatalf("audit log missing %q (got %v)", want, actions)
		}
	}
}

// TestProfile_UpdateMasksBankAccount updates a profile and confirms the bank
// account is stored masked (profile.go Update/maskBankAccount; US-006, NFR-02).
func TestProfile_UpdateMasksBankAccount(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ps := NewProfileService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	u, err := ps.Update(ctx, userID, UpdateProfileInput{
		FullName: "Jane Investor", Mobile: "0812345678", BankAccount: "1234567890",
	})
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if u.FullName != "Jane Investor" {
		t.Fatalf("full name not updated: %s", u.FullName)
	}
	if !u.BankAccountMasked.Valid || u.BankAccountMasked.String == "1234567890" {
		t.Fatalf("bank account should be masked, got %q", u.BankAccountMasked.String)
	}
	// Only the last 4 digits remain visible.
	if got := u.BankAccountMasked.String; got[len(got)-4:] != "7890" {
		t.Fatalf("masked account should end with 7890, got %q", got)
	}

	// Empty full name rejected.
	if _, err := ps.Update(ctx, userID, UpdateProfileInput{FullName: "  "}); err == nil {
		t.Fatal("expected validation error for empty full name")
	}
}

// TestProfile_SubmitRiskAndGet stores a risk assessment and reads it back,
// covering derivation and the 1–5 bounds (profile.go SubmitRisk/GetRisk/
// deriveRiskLevel; US-007, BR-01).
func TestProfile_SubmitRiskAndGet(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ps := NewProfileService(store, clock)
	ctx := context.Background()

	userID := seedActiveUser(t, store, clock, decimal.Zero, 0)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	// No assessment yet.
	if _, ok, err := ps.GetRisk(ctx, userID); err != nil || ok {
		t.Fatalf("expected no risk assessment, ok=%v err=%v", ok, err)
	}

	// Explicit level.
	r, err := ps.SubmitRisk(ctx, userID, SubmitRiskInput{RiskLevel: 4, Answers: map[string]int{"q1": 3}})
	if err != nil {
		t.Fatalf("submit risk: %v", err)
	}
	if r.RiskLevel != 4 || !r.Valid {
		t.Fatalf("unexpected risk result: %+v", r)
	}

	got, ok, err := ps.GetRisk(ctx, userID)
	if err != nil || !ok {
		t.Fatalf("get risk: ok=%v err=%v", ok, err)
	}
	if got.RiskLevel != 4 || !got.Valid {
		t.Fatalf("read-back risk mismatch: %+v", got)
	}

	// Derivation from answers (avg == 3 → level 3).
	der, err := ps.SubmitRisk(ctx, userID, SubmitRiskInput{Answers: map[string]int{"a": 3, "b": 3}})
	if err != nil {
		t.Fatalf("submit derived risk: %v", err)
	}
	if der.RiskLevel != 3 {
		t.Fatalf("derived risk level=%d want 3", der.RiskLevel)
	}

	// Out-of-range explicit level rejected.
	_, err = ps.SubmitRisk(ctx, userID, SubmitRiskInput{RiskLevel: 9})
	var appErr *domain.AppError
	if err == nil || !asApp(err, &appErr) || appErr.Code != domain.CodeValidationError {
		t.Fatalf("expected VALIDATION_ERROR for level 9, got %v", err)
	}

	// Toggle email notifications (profile.go SetEmailNotifications; US-029).
	if err := ps.SetEmailNotifications(ctx, userID, true); err != nil {
		t.Fatalf("set email notifications: %v", err)
	}
	u, _ := ps.Get(ctx, userID)
	if !u.EmailNotificationsEnabled {
		t.Fatal("email notifications flag not set")
	}
}

// The profile form shows only the masked account number, so saving after
// editing just the name/mobile sends no bank account. That must keep the account
// on file (it used to be overwritten with NULL).
func TestProfile_UpdateWithoutBankAccountKeepsExisting(t *testing.T) {
	store := openTestStore(t)
	pool := rawPool(t)
	clock := testClock()
	ps := NewProfileService(store, clock)
	ctx := context.Background()
	userID := seedActiveUser(t, store, clock, decimal.Zero, 3)
	t.Cleanup(func() { cleanupUser(t, pool, userID) })

	first, err := ps.Update(ctx, userID, UpdateProfileInput{
		FullName: "Jane Investor", Mobile: "0812345678", BankAccount: "1234567890",
	})
	if err != nil {
		t.Fatalf("first update: %v", err)
	}
	want := first.BankAccountMasked.String
	if want == "" {
		t.Fatal("expected a masked account after the first update")
	}

	second, err := ps.Update(ctx, userID, UpdateProfileInput{FullName: "Jane Renamed", Mobile: "0899999999"})
	if err != nil {
		t.Fatalf("second update: %v", err)
	}
	if second.FullName != "Jane Renamed" || second.Mobile.String != "0899999999" {
		t.Errorf("name/mobile were not updated: %+v", second)
	}
	if second.BankAccountMasked.String != want {
		t.Errorf("bank account changed from %q to %q when none was supplied", want, second.BankAccountMasked.String)
	}

	// Supplying a new account replaces it.
	third, err := ps.Update(ctx, userID, UpdateProfileInput{FullName: "Jane Renamed", BankAccount: "9999900001"})
	if err != nil {
		t.Fatalf("third update: %v", err)
	}
	if got := third.BankAccountMasked.String; got == want || got[len(got)-4:] != "0001" {
		t.Errorf("new account should replace the old one and end in 0001, got %q", got)
	}
}
