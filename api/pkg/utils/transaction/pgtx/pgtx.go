// SPDX-License-Identifier: AGPL-3.0-or-later

package pgtx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"gorm.io/gorm"
)

var _ transaction.Transactor = (*PGTransactor)(nil)

const MaxAttempts = 3

const serializationFailure = "40001"

type txKey struct{}

func From(ctx context.Context, root *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}

	return root
}

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGTransactor struct {
	deps Deps
}

func New(deps Deps) (*PGTransactor, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &PGTransactor{deps: deps}, nil
}

func (t *PGTransactor) WithinTx(ctx context.Context, opts transaction.TxOpts, fn func(context.Context) error) error {
	var err error

	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		err = t.runOnce(ctx, opts, fn)
		if !isSerializationFailure(err) {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}

	return nil
}

func (t *PGTransactor) runOnce(ctx context.Context, opts transaction.TxOpts, fn func(context.Context) error) error {
	err := t.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	}, txOptions(opts))
	if err != nil {
		return fmt.Errorf("db.Transaction: %w", err)
	}

	return nil
}

func txOptions(opts transaction.TxOpts) *sql.TxOptions {
	isolation := sql.LevelDefault
	if opts.Isolation == transaction.IsolationSerializable {
		isolation = sql.LevelSerializable
	}

	return &sql.TxOptions{Isolation: isolation}
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == serializationFailure
}
