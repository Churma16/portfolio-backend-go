package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store menyediakan semua fungsi query + fungsi transaction
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore membuat instance Store baru
func NewStore(database *sql.DB) *Store {
	return &Store{
		db:      database,
		Queries: New(database),
	}
}

// execTx menjalankan fungsi di dalam database transaction
func (store *Store) ExecTx(context context.Context, transactionFunc func(*Queries) error) error {
	// Begin a new transaction
	transaction, transactionError := store.db.BeginTx(context, nil)
	if transactionError != nil {
		return transactionError
	}

	queries := New(transaction)
	executionError := transactionFunc(queries)
	if executionError != nil {
		// Rollback the transaction in case of an error
		rollbackError := transaction.Rollback()
		if rollbackError != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", executionError, rollbackError)
		}
		return executionError
	}

	// Commit the transaction
	return transaction.Commit()
}
