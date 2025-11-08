package user

import (
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "gomall/db/sqlc"
)

// Repository Direct use Store Interface
type Repository interface {
	sqlc.Store
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return sqlc.NewStore(pool)
}
