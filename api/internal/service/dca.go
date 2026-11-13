package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/money"
	"github.com/webond/api/internal/repo"
)

// DCA minimums (BR-07) and the consecutive-failure pause threshold (EX-09).
var (
	minDcaFund = decimal.RequireFromString("500")
	minDcaGold = decimal.RequireFromString("100")
)

const maxConsecutiveFailures = 3 // pause plan after 3 fails (EX-09)

// DCAService manages recurring investment plans for MF & Gold only (bond
// excluded, BR-06), computes next-run dates, and executes due plans idempotently
// (one round per plan+run_date, FR-15). Execution delegates to TradingService
// (components.md § DCAService + Scheduler; US-019..021).
type DCAService struct {
	store   *repo.Store
	trading *TradingService
	notify  *NotificationService
	clock   Clock
}

// NewDCAService builds a DCAService.
func NewDCAService(store *repo.Store, trading *TradingService, notify *NotificationService, clock Clock) *DCAService {
	return &DCAService{store: store, trading: trading, notify: notify, clock: clock}
}

// CreatePlanInput is the create-plan request (US-019).
type CreatePlanInput struct {
	UserID         uint64
	AssetType      string // fund|gold (bond rejected, BR-06)
	AssetID        uint64
	AmountPerRound decimal.Decimal
	Frequency      string // weekly|monthly
	RunDay         int    // 1–7 weekly / 1–28 monthly
	StartDate      string // yyyy-MM-dd
	EndDate        string // optional
	MaxRounds      int    // optional
}

// CreatePlan validates and creates a DCA plan (US-019, BR-06/07).
func (s *DCAService) CreatePlan(ctx context.Context, in CreatePlanInput) (db.DcaPlan, error) {
	if in.AssetType == "bond" {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError, "DCA is available for Mutual Fund and Gold only")
	}
	if in.AssetType != "fund" && in.AssetType != "gold" {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError, "Unknown asset type for DCA")
	}
	min := minDcaFund
	if in.AssetType == "gold" {
		min = minDcaGold
	}
	amount := money.Quantize(in.AmountPerRound)
	if amount.LessThan(min) {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError,
			fmt.Sprintf("Minimum per round is %s THB", money.FormatTHB(min)))
	}
	freq := strings.ToLower(in.Frequency)
	if freq != "weekly" && freq != "monthly" {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError, "Frequency must be weekly or monthly")
	}
	if err := validateRunDay(freq, in.RunDay); err != nil {
		return db.DcaPlan{}, err
	}
	start, err := parseDate(in.StartDate, s.clock.Now())
	if err != nil {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError, "Invalid start date")
	}
	var end sql.NullTime
	if in.EndDate != "" {
		if e, perr := parseDate(in.EndDate, s.clock.Now()); perr == nil {
			end = nullTime(e)
		}
	}
	var maxRounds sql.NullInt32
	if in.MaxRounds > 0 {
		maxRounds = sql.NullInt32{Int32: int32(in.MaxRounds), Valid: true}
	}

	next := nextRunDate(freq, in.RunDay, start, start)

	res, err := s.store.Dca.CreatePlan(ctx, db.CreateDcaPlanParams{
		UserID:         in.UserID,
		AssetType:      db.DcaPlansAssetType(in.AssetType),
		AssetID:        in.AssetID,
		AmountPerRound: amount,
		Frequency:      db.DcaPlansFrequency(freq),
		RunDay:         int8(in.RunDay),
		StartDate:      start,
		EndDate:        end,
		MaxRounds:      maxRounds,
		NextRunDate:    nullTime(next),
	})
	id, err := lastInsertID(res, err)
	if err != nil {
		return db.DcaPlan{}, fmt.Errorf("create plan: %w", err)
	}
	return s.store.Dca.GetPlanByID(ctx, id)
}

// ListPlans returns the user's plans (US-020).
func (s *DCAService) ListPlans(ctx context.Context, userID uint64) ([]db.DcaPlan, error) {
	return s.store.Dca.ListPlansByUser(ctx, userID)
}

// UpdatePlan edits amount/run day and recomputes next-run (US-020).
func (s *DCAService) UpdatePlan(ctx context.Context, userID, planID uint64, amount decimal.Decimal, runDay int) (db.DcaPlan, error) {
	plan, err := s.ownedPlan(ctx, userID, planID)
	if err != nil {
		return db.DcaPlan{}, err
	}
	min := minDcaFund
	if plan.AssetType == db.DcaPlansAssetTypeGold {
		min = minDcaGold
	}
	amount = money.Quantize(amount)
	if amount.LessThan(min) {
		return db.DcaPlan{}, domain.NewError(domain.CodeValidationError,
			fmt.Sprintf("Minimum per round is %s THB", money.FormatTHB(min)))
	}
	if err := validateRunDay(string(plan.Frequency), runDay); err != nil {
		return db.DcaPlan{}, err
	}
	next := nextRunDate(string(plan.Frequency), runDay, s.clock.Now(), plan.StartDate)
	if err := s.store.Dca.UpdatePlan(ctx, db.UpdateDcaPlanParams{
		AmountPerRound: amount, RunDay: int8(runDay), NextRunDate: nullTime(next), ID: planID,
	}); err != nil {
		return db.DcaPlan{}, fmt.Errorf("update plan: %w", err)
	}
	return s.store.Dca.GetPlanByID(ctx, planID)
}

// SetStatus pauses/resumes/cancels a plan (US-020).
func (s *DCAService) SetStatus(ctx context.Context, userID, planID uint64, status db.DcaPlansStatus) error {
	if _, err := s.ownedPlan(ctx, userID, planID); err != nil {
		return err
	}
	if err := s.store.Dca.SetPlanStatus(ctx, db.SetDcaPlanStatusParams{Status: status, ID: planID}); err != nil {
		return fmt.Errorf("set plan status: %w", err)
	}
	return nil
}

// ListRounds returns a plan's round history (US-020).
func (s *DCAService) ListRounds(ctx context.Context, userID, planID uint64) ([]db.DcaRound, error) {
	if _, err := s.ownedPlan(ctx, userID, planID); err != nil {
		return nil, err
	}
	return s.store.Dca.ListRounds(ctx, planID)
}

// RunDue executes all active plans whose next_run_date is on/before "now"
// (US-021; manual trigger endpoint). It is idempotent per plan+run_date via the
// UNIQUE(dca_plan_id, run_date) constraint (FR-15). Returns the number of plans
// processed.
func (s *DCAService) RunDue(ctx context.Context) (int, error) {
	today := s.today()
	plans, err := s.store.Dca.ListDuePlans(ctx, nullTime(today))
	if err != nil {
		return 0, fmt.Errorf("list due plans: %w", err)
	}
	for _, p := range plans {
		s.executeRound(ctx, p, today)
	}
	return len(plans), nil
}

// executeRound runs a single DCA round for a plan on runDate (US-021). Idempotent
// per plan+date; on failure it records a failed round, notifies, and pauses the
// plan after 3 consecutive failures (EX-09).
func (s *DCAService) executeRound(ctx context.Context, plan db.DcaPlan, runDate time.Time) {
	// Idempotency guard (FR-15): if a round already exists for this date, skip.
	rounds, _ := s.store.Dca.ListRounds(ctx, plan.ID)
	for _, r := range rounds {
		if sameDate(r.RunDate, runDate) {
			return
		}
	}

	orderErr := s.placeDcaBuy(ctx, plan)
	nextRun := nextRunDate(string(plan.Frequency), int(plan.RunDay), runDate.AddDate(0, 0, 1), plan.StartDate)

	if orderErr != nil {
		failures := plan.ConsecutiveFailures + 1
		status := db.DcaPlansStatusActive
		if int(failures) >= maxConsecutiveFailures {
			status = db.DcaPlansStatusPaused
		}
		_, _ = s.store.Dca.CreateRound(ctx, db.CreateDcaRoundParams{
			DcaPlanID: plan.ID, RunDate: runDate, Status: db.DcaRoundsStatusFailed,
			FailureReason: nullString(orderErr.Error()),
		})
		_ = s.store.Dca.RecordRun(ctx, db.RecordDcaRunParams{
			RoundsExecuted: plan.RoundsExecuted, TotalInvested: plan.TotalInvested,
			ConsecutiveFailures: failures, NextRunDate: nullTime(nextRun), Status: status, ID: plan.ID,
		})
		_ = s.notify.Emit(ctx, Notice{
			UserID: plan.UserID, Category: db.NotificationOutboxCategoryDca,
			Subject: "DCA round failed", Body: "Your DCA round could not be executed: " + orderErr.Error(),
			EmailEnabled: s.emailEnabled(ctx, plan.UserID),
		})
		return
	}

	// Success: record the round, advance counters, reset failures.
	roundNum := plan.RoundsExecuted + 1
	completed := db.DcaPlansStatusActive
	if plan.MaxRounds.Valid && roundNum >= plan.MaxRounds.Int32 {
		completed = db.DcaPlansStatusCompleted
	}
	_, _ = s.store.Dca.CreateRound(ctx, db.CreateDcaRoundParams{
		DcaPlanID: plan.ID, RunDate: runDate, Status: db.DcaRoundsStatusSuccess,
	})
	_ = s.store.Dca.RecordRun(ctx, db.RecordDcaRunParams{
		RoundsExecuted: roundNum, TotalInvested: money.Quantize(plan.TotalInvested.Add(plan.AmountPerRound)),
		ConsecutiveFailures: 0, NextRunDate: nullTime(nextRun), Status: completed, ID: plan.ID,
	})
	_ = s.notify.Emit(ctx, Notice{
		UserID: plan.UserID, Category: db.NotificationOutboxCategoryDca,
		Subject: "DCA round executed", Body: "Your DCA round was invested successfully.",
		EmailEnabled: s.emailEnabled(ctx, plan.UserID),
	})
}

// placeDcaBuy places the buy order for a DCA round via TradingService, tagging
// it as a DCA buy linked to the plan.
func (s *DCAService) placeDcaBuy(ctx context.Context, plan db.DcaPlan) error {
	in := OrderInput{
		UserID: plan.UserID, AssetType: string(plan.AssetType), AssetID: plan.AssetID,
		Side: "buy", OrderMode: "amount", Amount: plan.AmountPerRound,
		RiskAck: true, // DCA plans are pre-acknowledged at creation
		dcaType: true, dcaPlanID: int64(plan.ID),
	}
	_, err := s.trading.PlaceOrder(ctx, in)
	return err
}

// ownedPlan loads a plan enforcing ownership.
func (s *DCAService) ownedPlan(ctx context.Context, userID, planID uint64) (db.DcaPlan, error) {
	p, err := s.store.Dca.GetPlanByID(ctx, planID)
	if err != nil || p.UserID != userID {
		return db.DcaPlan{}, domain.NewError(domain.CodeTxnNotFound, "DCA plan not found")
	}
	return p, nil
}

func (s *DCAService) emailEnabled(ctx context.Context, userID uint64) bool {
	u, err := s.store.Users.GetByID(ctx, userID)
	if err != nil {
		return false
	}
	return u.EmailNotificationsEnabled
}

func (s *DCAService) today() time.Time {
	now := s.clock.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Bangkok)
}

// --- date helpers ---

func parseDate(s string, def time.Time) (time.Time, error) {
	if s == "" {
		return time.Date(def.Year(), def.Month(), def.Day(), 0, 0, 0, 0, Bangkok), nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, Bangkok)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func validateRunDay(freq string, day int) error {
	if freq == "weekly" {
		if day < 1 || day > 7 {
			return domain.NewError(domain.CodeValidationError, "Weekly run day must be 1–7")
		}
	} else {
		if day < 1 || day > 28 {
			return domain.NewError(domain.CodeValidationError, "Monthly run day must be 1–28")
		}
	}
	return nil
}

// nextRunDate computes the next execution date on/after `from`, not before the
// plan's start date, for the given frequency and run day.
func nextRunDate(freq string, runDay int, from, start time.Time) time.Time {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, Bangkok)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, Bangkok)
	if from.Before(start) {
		from = start
	}
	if freq == "weekly" {
		// runDay 1=Mon..7=Sun; Go Weekday Sunday=0..Saturday=6.
		target := time.Weekday(runDay % 7)
		for int(from.Weekday()) != int(target) {
			from = from.AddDate(0, 0, 1)
		}
		return from
	}
	// monthly: next occurrence of day-of-month runDay.
	y, m := from.Year(), from.Month()
	candidate := time.Date(y, m, runDay, 0, 0, 0, 0, Bangkok)
	if candidate.Before(from) {
		candidate = time.Date(y, m, runDay, 0, 0, 0, 0, Bangkok).AddDate(0, 1, 0)
	}
	return candidate
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
