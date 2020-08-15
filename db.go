package buna

import (
	"context"
	"database/sql"
)

type DB interface {
	insertCoffee(ctx context.Context, coffee coffee) error
	TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
	Close() error
}
