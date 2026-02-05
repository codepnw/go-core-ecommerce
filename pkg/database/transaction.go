package database

import (
	"context"
	"database/sql"
	"errors"
)

//go:generate mockgen -source=transaction.go -destination=transaction_mock.go -package=database
type TxManager interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
}

type txManager struct {
	db *sql.DB
}

func NewDBTransaction(db *sql.DB) TxManager {
	return &txManager{db: db}
}

func (t *txManager) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors.Join(err, rbErr)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
