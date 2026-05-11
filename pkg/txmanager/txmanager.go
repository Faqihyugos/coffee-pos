package txmanager

import (
	"context"
	"database/sql"
)

type TxManager struct {
	db *sql.DB
}

func New(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
