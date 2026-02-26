package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Querier
	WithTx(ctx context.Context, fn func(q Querier) error) error
}
type repository struct {
	*Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		Queries: New(pool),
		pool:    pool,
	}
}

// Обертка для транзакций
func (r *repository) WithTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err = fn(New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
