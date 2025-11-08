package sqlc

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines
type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(Querier) error) error
}

type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

func (store *SQLStore) ExecTx(ctx context.Context, fn func(Querier) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	q := New(tx)
	err = fn(q) // Pass Queries Interface
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
