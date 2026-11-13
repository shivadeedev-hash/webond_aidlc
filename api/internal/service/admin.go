package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/webond/api/internal/db"
	"github.com/webond/api/internal/domain"
	"github.com/webond/api/internal/repo"
)

// AdminService provides minimal master-data CRUD with audit logging
// (components.md § AdminService; FR-25, US-031, RBAC admin). Scope-limited per
// D1-1: create funds/bonds/gold and list audit logs.
type AdminService struct {
	store *repo.Store
	clock Clock
}

// NewAdminService builds an AdminService.
func NewAdminService(store *repo.Store, clock Clock) *AdminService {
	return &AdminService{store: store, clock: clock}
}

// CreateFundInput is the admin create-fund payload (US-031).
type CreateFundInput struct {
	Code        string
	Name        string
	Category    string
	RiskLevel   int16
	MinPurchase decimal.Decimal
}

// CreateFund creates a fund and writes an audit log entry (US-031, NFR-02).
func (s *AdminService) CreateFund(ctx context.Context, adminID uint64, in CreateFundInput) (uint64, error) {
	if in.Code == "" || in.Name == "" {
		return 0, domain.NewError(domain.CodeValidationError, "Code and name are required")
	}
	min := in.MinPurchase
	if min.LessThanOrEqual(decimal.Zero) {
		min = decimal.RequireFromString("500")
	}
	id, err := lastInsertID(s.store.Funds.Create(ctx, db.CreateFundParams{
		Code: in.Code, Name: in.Name,
		Category:    nullString(in.Category),
		RiskLevel:   nullInt16(in.RiskLevel),
		MinPurchase: min, SettlementDays: 3, CutoffTime: "15:30:00", IsActive: true,
	}))
	if err != nil {
		return 0, fmt.Errorf("create fund: %w", err)
	}
	s.audit(ctx, adminID, "create_fund", "fund", int64(id), in)
	return id, nil
}

// CreateBondInput is the admin create-bond payload (US-031).
type CreateBondInput struct {
	Code            string
	Name            string
	Issuer          string
	CouponRate      decimal.Decimal
	CouponFrequency int8
	FaceValue       decimal.Decimal
	RiskLevel       int16
}

// CreateBond creates a bond and writes an audit log entry (US-031).
func (s *AdminService) CreateBond(ctx context.Context, adminID uint64, in CreateBondInput) (uint64, error) {
	if in.Code == "" || in.Name == "" {
		return 0, domain.NewError(domain.CodeValidationError, "Code and name are required")
	}
	face := in.FaceValue
	if face.LessThanOrEqual(decimal.Zero) {
		face = decimal.RequireFromString("1000")
	}
	id, err := lastInsertID(s.store.Bonds.Create(ctx, db.CreateBondParams{
		Code: in.Code, Name: in.Name, Issuer: nullString(in.Issuer),
		CouponRate: in.CouponRate, CouponFrequency: in.CouponFrequency, FaceValue: face,
		MinLot: decimal.RequireFromString("100000"), RiskLevel: nullInt16(in.RiskLevel), IsActive: true,
	}))
	if err != nil {
		return 0, fmt.Errorf("create bond: %w", err)
	}
	s.audit(ctx, adminID, "create_bond", "bond", int64(id), in)
	return id, nil
}

// CreateGoldInput is the admin create-gold payload (US-031).
type CreateGoldInput struct {
	Code   string
	Name   string
	Purity string
}

// CreateGold creates a gold product and writes an audit log entry (US-031).
func (s *AdminService) CreateGold(ctx context.Context, adminID uint64, in CreateGoldInput) (uint64, error) {
	if in.Code == "" || in.Name == "" {
		return 0, domain.NewError(domain.CodeValidationError, "Code and name are required")
	}
	id, err := lastInsertID(s.store.Gold.Create(ctx, db.CreateGoldProductParams{
		Code: in.Code, Name: in.Name, Purity: nullString(in.Purity), IsActive: true,
	}))
	if err != nil {
		return 0, fmt.Errorf("create gold: %w", err)
	}
	s.audit(ctx, adminID, "create_gold", "gold", int64(id), in)
	return id, nil
}

// ListAuditLogs returns a page of audit logs for a user (US-031; admin view).
func (s *AdminService) ListAuditLogs(ctx context.Context, userID uint64, page, pageSize int) ([]db.AuditLog, error) {
	limit, offset := pageBounds(page, pageSize)
	return s.store.Audit.ListByUser(ctx, db.ListAuditLogsByUserParams{
		UserID: nullInt64(int64(userID)), Limit: int32(limit), Offset: int32(offset),
	})
}

func (s *AdminService) audit(ctx context.Context, adminID uint64, action, entity string, entityID int64, detail any) {
	raw, _ := json.Marshal(detail)
	_, _ = s.store.Audit.Create(ctx, db.CreateAuditLogParams{
		UserID: nullInt64(int64(adminID)), Action: action, Entity: nullString(entity),
		EntityID: nullInt64(entityID), DetailJson: raw,
	})
}
