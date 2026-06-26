package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store menyediakan semua fungsi query + fungsi transaction
type Store struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore membuat instance Store baru
func NewStore(database *pgxpool.Pool) *Store {
	return &Store{
		db:      database,
		Queries: New(database),
	}
}

// ExecTx menjalankan fungsi di dalam database transaction
func (store *Store) ExecTx(ctx context.Context, transactionFunc func(*Queries) error) error {
	// Begin a new transaction
	transaction, transactionError := store.db.Begin(ctx)
	if transactionError != nil {
		return transactionError
	}

	queries := New(transaction)
	executionError := transactionFunc(queries)
	if executionError != nil {
		// Rollback the transaction in case of an error
		rollbackError := transaction.Rollback(ctx)
		if rollbackError != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", executionError, rollbackError)
		}
		return executionError
	}

	// Commit the transaction
	return transaction.Commit(ctx)
}
