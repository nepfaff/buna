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
	insertCupping(ctx context.Context, cupping cupping) error
	insertGrinder(ctx context.Context, grinder grinder) error

	// retrieve
	getCoffeeNameSuggestions(ctx context.Context, limit int) ([]string, error)
	getBrewingsByLastAdded(ctx context.Context, limit int) ([]brewing, error)
	getBrewingSuggestions(ctx context.Context, limit int, brewingFilter brewing) ([]brewing, error)
	getCoffeeIDByNameRoaster(ctx context.Context, name string, roaster string) (int, error)
	getCoffeesByLastAdded(ctx context.Context, limit int) ([]coffee, error)
	getCuppingsByLastAdded(ctx context.Context, limit int) ([]cupping, error)
	getGrinderIDByName(ctx context.Context, name string) (int, error)
	getMethodIDByName(ctx context.Context, name string) (int, error)
	getLastCoffeeRoastDate(ctx context.Context, coffeeName string) (date, error)
	getMostRecentlyUsedBrewingMethodNames(ctx context.Context, limit int) ([]string, error)
	getMostRecentlyUsedCoffeeGrinderNames(ctx context.Context, limit int) ([]string, error)
	getMostRecentlyUsedCoffeeWeights(ctx context.Context, brewingMethodName string, coffeeGrinderName string, limit int) ([]int, error)
	getMostRecentlyUsedWaterWeights(ctx context.Context, brewingMethodName string, coffeeGrinderName string, limit int) ([]int, error)
	getRoastersByCoffeeName(ctx context.Context, name string, limit int) ([]string, error)

	// statistics
	getAverageBrewingRating(ctx context.Context, brewingFilter brewing) (float64, error)
	getTotalCount(ctx context.Context, entity dbEntity) (int, error)

	// general
	TransactContext(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
	Close() error
}
