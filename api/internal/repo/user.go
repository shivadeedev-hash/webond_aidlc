package repo

import (
	"context"
	"database/sql"

	"github.com/webond/api/internal/db"
)

// UserRepo is account & auth persistence (components.md § AuthService).
type UserRepo interface {
	// LockRow takes a primary-key row lock on the user (inside a transaction).
	LockRow(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	Create(ctx context.Context, arg db.CreateUserParams) (sql.Result, error)
	SetStatus(ctx context.Context, arg db.SetUserStatusParams) error
	UpdateProfile(ctx context.Context, arg db.UpdateUserProfileParams) error
	UpdatePassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
	SetEmailNotifications(ctx context.Context, arg db.SetEmailNotificationsParams) error
	IncrementFailedLogin(ctx context.Context, id uint64) error
	ResetFailedLogin(ctx context.Context, id uint64) error
	Lock(ctx context.Context, arg db.LockUserParams) error
}

type userRepo struct{ q *db.Queries }

func (r *userRepo) LockRow(ctx context.Context, id uint64) error {
	_, err := r.q.LockUserRow(ctx, id)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uint64) (db.User, error) {
	return r.q.GetUserByID(ctx, id)
}
func (r *userRepo) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}
func (r *userRepo) Create(ctx context.Context, arg db.CreateUserParams) (sql.Result, error) {
	return r.q.CreateUser(ctx, arg)
}
func (r *userRepo) SetStatus(ctx context.Context, arg db.SetUserStatusParams) error {
	return r.q.SetUserStatus(ctx, arg)
}
func (r *userRepo) UpdateProfile(ctx context.Context, arg db.UpdateUserProfileParams) error {
	return r.q.UpdateUserProfile(ctx, arg)
}
func (r *userRepo) UpdatePassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return r.q.UpdateUserPassword(ctx, arg)
}
func (r *userRepo) SetEmailNotifications(ctx context.Context, arg db.SetEmailNotificationsParams) error {
	return r.q.SetEmailNotifications(ctx, arg)
}
func (r *userRepo) IncrementFailedLogin(ctx context.Context, id uint64) error {
	return r.q.IncrementFailedLogin(ctx, id)
}
func (r *userRepo) ResetFailedLogin(ctx context.Context, id uint64) error {
	return r.q.ResetFailedLogin(ctx, id)
}
func (r *userRepo) Lock(ctx context.Context, arg db.LockUserParams) error {
	return r.q.LockUser(ctx, arg)
}

// RiskRepo is risk-assessment persistence (FR-04, BR-01).
type RiskRepo interface {
	Create(ctx context.Context, arg db.CreateRiskAssessmentParams) (sql.Result, error)
	GetLatest(ctx context.Context, userID uint64) (db.RiskAssessment, error)
}

type riskRepo struct{ q *db.Queries }

func (r *riskRepo) Create(ctx context.Context, arg db.CreateRiskAssessmentParams) (sql.Result, error) {
	return r.q.CreateRiskAssessment(ctx, arg)
}
func (r *riskRepo) GetLatest(ctx context.Context, userID uint64) (db.RiskAssessment, error) {
	return r.q.GetLatestRiskAssessment(ctx, userID)
}
