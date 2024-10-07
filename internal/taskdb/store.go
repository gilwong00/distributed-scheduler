package taskdb

import (
	"context"
	"database/sql"
	"fmt"

	taskpostgres "github.com/gilwong00/task-runner/internal/taskdb/gen"
)

type Store struct {
	querier *taskpostgres.Queries
	db      *sql.DB
}

func newStore(db *sql.DB) *Store {
	return &Store{
		querier: taskpostgres.New(db),
		db:      db,
	}
}

func (s *Store) ReadTx(
	ctx context.Context,
	fn func(*taskpostgres.Queries) error,
) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}
	query := taskpostgres.New(tx)
	if err = fn(query); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func (s *Store) ReadWriteTx(ctx context.Context, fn func(*taskpostgres.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: false,
	})
	if err != nil {
		return err
	}
	if err := fn(taskpostgres.New(tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}
