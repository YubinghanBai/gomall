package user

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gomall/db/sqlc"
)

// Repository defines the data access interface for user domain
// This interface only includes user-related methods to maintain clear domain boundaries
type Repository interface {
	// CreateUser User operations
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByID(ctx context.Context, id int64) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	GetUserByUsername(ctx context.Context, username string) (sqlc.User, error)
	UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) error
	UpdateUserPassword(ctx context.Context, arg sqlc.UpdateUserPasswordParams) error
	UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error
	VerifyUserEmail(ctx context.Context, id int64) error

	// CreateSession Session operations (user authentication related)
	CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) (sqlc.Session, error)
	GetSession(ctx context.Context, id uuid.UUID) (sqlc.Session, error)
	GetUserSessions(ctx context.Context, userID int64) ([]sqlc.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID int64) error

	// CreateVerificationCode Verification code operations (user verification related)
	CreateVerificationCode(ctx context.Context, arg sqlc.CreateVerificationCodeParams) (sqlc.VerificationCode, error)
	GetVerificationCode(ctx context.Context, arg sqlc.GetVerificationCodeParams) (sqlc.VerificationCode, error)
	MarkCodeAsUsed(ctx context.Context, id int64) error

	// ExecTx Transaction support
	ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error
}

// repository implements Repository interface
type repository struct {
	store sqlc.Store
}

// NewRepository creates a new Repository instance
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		store: sqlc.NewStore(pool),
	}
}

// User operations

func (r *repository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.store.CreateUser(ctx, arg)
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (sqlc.User, error) {
	return r.store.GetUserByID(ctx, id)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.store.GetUserByEmail(ctx, email)
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (sqlc.User, error) {
	return r.store.GetUserByUsername(ctx, username)
}

func (r *repository) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) error {
	return r.store.UpdateUser(ctx, arg)
}

func (r *repository) UpdateUserPassword(ctx context.Context, arg sqlc.UpdateUserPasswordParams) error {
	return r.store.UpdateUserPassword(ctx, arg)
}

func (r *repository) UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error {
	return r.store.UpdateUserLastLogin(ctx, arg)
}

func (r *repository) VerifyUserEmail(ctx context.Context, id int64) error {
	return r.store.VerifyUserEmail(ctx, id)
}

// Session operations

func (r *repository) CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) (sqlc.Session, error) {
	return r.store.CreateSession(ctx, arg)
}

func (r *repository) GetSession(ctx context.Context, id uuid.UUID) (sqlc.Session, error) {
	return r.store.GetSession(ctx, id)
}

func (r *repository) GetUserSessions(ctx context.Context, userID int64) ([]sqlc.Session, error) {
	return r.store.GetUserSessions(ctx, userID)
}

func (r *repository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.store.DeleteSession(ctx, id)
}

func (r *repository) DeleteUserSessions(ctx context.Context, userID int64) error {
	return r.store.DeleteUserSessions(ctx, userID)
}

// Verification code operations

func (r *repository) CreateVerificationCode(ctx context.Context, arg sqlc.CreateVerificationCodeParams) (sqlc.VerificationCode, error) {
	return r.store.CreateVerificationCode(ctx, arg)
}

func (r *repository) GetVerificationCode(ctx context.Context, arg sqlc.GetVerificationCodeParams) (sqlc.VerificationCode, error) {
	return r.store.GetVerificationCode(ctx, arg)
}

func (r *repository) MarkCodeAsUsed(ctx context.Context, id int64) error {
	return r.store.MarkCodeAsUsed(ctx, id)
}

// Transaction support

func (r *repository) ExecTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	return r.store.ExecTx(ctx, fn)
}
