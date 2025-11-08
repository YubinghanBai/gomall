package user

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "gomall/db/sqlc"
)

// For Mock
type Repository interface {
	// User
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByID(ctx context.Context, id int64) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	GetUserByUsername(ctx context.Context, username string) (sqlc.User, error)
	UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) error
	UpdateUserPassword(ctx context.Context, arg sqlc.UpdateUserPasswordParams) error
	VerifyUserEmail(ctx context.Context, id int64) error
	UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error

	// Session
	CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) (sqlc.Session, error)
	GetSession(ctx context.Context, id uuid.UUID) (sqlc.Session, error)
	GetUserSessions(ctx context.Context, userID int64) ([]sqlc.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID int64) error
	BlockSession(ctx context.Context, id uuid.UUID) error

	// Verification Code
	CreateVerificationCode(ctx context.Context, arg sqlc.CreateVerificationCodeParams) (sqlc.VerificationCode, error)
	GetVerificationCode(ctx context.Context, arg sqlc.GetVerificationCodeParams) (sqlc.VerificationCode, error)
	MarkCodeAsUsed(ctx context.Context, id int64) error
}

type repository struct {
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		//Inject database connection pool to sqlc.Queries
		queries: sqlc.New(pool),
	}
}

// User Methods
func (r *repository) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (sqlc.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (sqlc.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *repository) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) error {
	return r.queries.UpdateUser(ctx, arg)
}

func (r *repository) UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error {
	return r.queries.UpdateUserLastLogin(ctx, arg)
}

func (r *repository) UpdateUserPassword(ctx context.Context, arg sqlc.UpdateUserPasswordParams) error {
	return r.queries.UpdateUserPassword(ctx, arg)
}

func (r *repository) VerifyUserEmail(ctx context.Context, id int64) error {
	return r.queries.VerifyUserEmail(ctx, id)
}

// Session Methods
func (r *repository) CreateSession(ctx context.Context, arg sqlc.CreateSessionParams) (sqlc.Session, error) {
	return r.queries.CreateSession(ctx, arg)
}

func (r *repository) GetSession(ctx context.Context, id uuid.UUID) (sqlc.Session, error) {
	return r.queries.GetSession(ctx, id)
}

func (r *repository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteSession(ctx, id)
}

func (r *repository) BlockSession(ctx context.Context, id uuid.UUID) error {
	return r.queries.BlockSession(ctx, id)
}

func (r *repository) GetUserSessions(ctx context.Context, userID int64) ([]sqlc.Session, error) {
	return r.queries.GetUserSessions(ctx, userID)
}

func (r *repository) DeleteUserSessions(ctx context.Context, userID int64) error {
	return r.queries.DeleteUserSessions(ctx, userID)
}

// VerificationCode
func (r *repository) CreateVerificationCode(ctx context.Context, arg sqlc.CreateVerificationCodeParams) (sqlc.VerificationCode, error) {
	return r.queries.CreateVerificationCode(ctx, arg)
}

func (r *repository) GetVerificationCode(ctx context.Context, arg sqlc.GetVerificationCodeParams) (sqlc.VerificationCode, error) {
	return r.queries.GetVerificationCode(ctx, arg)
}

func (r *repository) MarkCodeAsUsed(ctx context.Context, id int64) error {
	return r.queries.MarkCodeAsUsed(ctx, id)
}
