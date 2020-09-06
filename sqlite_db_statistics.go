package buna

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type dbEntity int

const (
	brewings dbEntity = iota
	brewingMethods
	coffees
	coffeePurchases
	cuppings
	grinders
)

var (
	dbEntityToStringMap = map[dbEntity]string{
		brewings:        "brewings",
		brewingMethods:  "brewing_methods",
		coffees:         "coffees",
		coffeePurchases: "purchases",
		cuppings:        "cuppings",
		grinders:        "grinders",
	}

	dbEntityToName = map[dbEntity]string{
		brewings:        "brewings",
		brewingMethods:  "brewing methods",
		coffees:         "coffees",
		coffeePurchases: "coffee purchases",
		cuppings:        "cuppings",
		grinders:        "grinders",
	}
)

// The following fields are used from the brewingFilter argument:
// brewingMethodName, v60FilterType, coffeeName, coffeeRoaster, grinderName
func (s *SQLiteDB) getAverageBrewingRating(ctx context.Context, brewingFilter brewing) (float64, error) {
	var averageBrewingRatingFloat float64
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var averageBrewingRating interface{}
		if err := tx.QueryRowContext(ctx, `
			SELECT avg(b.rating)
			FROM brewings AS b
			INNER JOIN coffees AS c
				ON c.id = b.coffee_id
			INNER JOIN brewing_methods AS m
				ON m.id = b.method_id
			INNER JOIN grinders AS g
				ON g.id = b.grinder_id
			WHERE (m.name = :brewingMethodName OR "" = :brewingMethodName)
			AND (b.v60_filter_type = :v60FilterType OR "" = :v60FilterType)
			AND (c.name = :coffeeName OR "" = :coffeeName)
			AND (c.roaster = :coffeeRoaster OR "" = :coffeeRoaster)
			AND (g.name = :grinderName OR "" = :grinderName)
		`,
			sql.Named("brewingMethodName", brewingFilter.brewingMethodName),
			sql.Named("v60FilterType", brewingFilter.v60FilterType),
			sql.Named("coffeeName", brewingFilter.coffeeName),
			sql.Named("coffeeRoaster", brewingFilter.coffeeRoaster),
			sql.Named("grinderName", brewingFilter.grinderName),
		).Scan(&averageBrewingRating); err != nil {
			return fmt.Errorf("buna: sqlite_db_statistics: failed to retrieve average brewing rating from db: %w", err)
		}

		if v := reflect.ValueOf(averageBrewingRating); v.Kind() == reflect.Float64 {
			averageBrewingRatingFloat = averageBrewingRating.(float64)
		} else {
			// no brewings exist
			return nil
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_statistics: getAverageBrewingRating transaction failed: %w", err)
	}

	return averageBrewingRatingFloat, nil
}

func (s *SQLiteDB) getTotalCount(ctx context.Context, entity dbEntity) (int, error) {
	dbEntityString, ok := dbEntityToStringMap[entity]
	if !ok {
		return 0, fmt.Errorf("buna: sqlite_db_statistics: unable to map dbEntity to string")
	}

	var count int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		query := `
			SELECT count(*)
			FROM ?
		`
		query = strings.Replace(query, "?", dbEntityString, 1)
		if err := tx.QueryRowContext(ctx, query, sql.Named("entity", dbEntityString)).Scan(&count); err != nil {
			return fmt.Errorf("buna: sqlite_db_statistics: failed to retrieve total count from db: %w", err)
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("buna: sqlite_db_statistics: getTotalCount transaction failed: %w", err)
	}

	return count, nil
}
