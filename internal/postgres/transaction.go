package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const (
	txKey = contextKey("tx")
)

type TransactionManager struct {
	db DBTX
}

func NewTransactionManager(db DBTX) TransactionManager {
	return TransactionManager{
		db: db,
	}
}

func (tm TransactionManager) WithinTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	db := tm.db.(*pgxpool.Pool)

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute the function
	err = fn(tx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
