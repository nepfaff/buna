package buna

import (
	"context"
	"database/sql"
)

type DB interface {
	// insert
	insertBrewing(ctx context.Context, brewing brewing) error
	insertBrewingMethod(ctx context.Context, brewingMethod brewingMethod) error
	insertCoffee(ctx context.Context, coffee coffee) error
	insertCoffeePurchase(ctx context.Context, coffeePurchase coffeePurchase) error
	insertGrinder(ctx context.Context, grinder grinder) error

	// retrieve
	getMostRecentBrewedCoffeeNames(ctx context.Context, limit int) ([]string, error)
	getMostRecentlyUsedBrewingMethodNames(ctx context.Context, limit int) ([]string, error)
	getRoastersByCoffeeName(ctx context.Context, name string, limit int) ([]string, error)

	// general
	TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
	Close() error
}
