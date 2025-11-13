package db

import (
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ForeignKeyViolation = "23505"
	UniqueViolation     = "23503"
)

var (
	ErrRecordNotFound    = pgx.ErrNoRows
	ErrUniqueViolation   = &pgconn.PgError{
		Code: UniqueViolation,
	}
	ErrInsufficientStock = errors.New("insufficient stock")
)

func ErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
