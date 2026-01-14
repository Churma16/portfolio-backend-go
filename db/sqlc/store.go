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
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx menjalankan fungsi di dalam database transaction
func (store *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
