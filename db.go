package buna

import (
	"context"
	"database/sql"
)

type DB interface {
	// insert
	insertCoffee(ctx context.Context, coffee coffee) error
	insertCoffeePurchase(ctx context.Context, coffeePurchase coffeePurchase) error

	// retrieve

	// general
	TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
	Close() error
}
