package buna

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

// limit determines the number of strings in the returned slice.
// The first suggestions are the most recently brewed coffees.
// The last suggestion is the most recently purchased coffee (The last two if limit > 5).
func (s *SQLiteDB) getCoffeeNameSuggestions(ctx context.Context, limit int) ([]string, error) {
	brewedLimit := limit - 1
	purchasedLimit := 1
	if limit > 5 {
		brewedLimit--
		purchasedLimit++
	}

	var names []string
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		bRows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT c.name 
			FROM brewings as b
			INNER JOIN coffees as c
				ON b.coffee_id = c.id
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", brewedLimit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve recently brewed coffee names: %w", err)
		}
		defer bRows.Close()

		for bRows.Next() {
			var name string
			if err := bRows.Scan(&name); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan bRow: %w", err)
			}

			names = append(names, name)
		}

		if err := bRows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last bRow: %w", err)
		}

		pRows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT c.name 
			FROM purchases as p
			INNER JOIN coffees as c
				ON p.coffee_id = c.id
			ORDER BY p.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", purchasedLimit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve recently purchased coffee names: %w", err)
		}
		defer pRows.Close()

		for pRows.Next() {
			var name string
			if err := pRows.Scan(&name); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan pRow: %w", err)
			}

			names = append(names, name)
		}

		if err := pRows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last pRow: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getCoffeeNameSuggestions transaction failed: %w", err)
	}

	names = removeStrDuplicates(names)

	return names, nil
}

func (s *SQLiteDB) getLastCoffeeRoastDate(ctx context.Context, coffeeName string) (date, error) {
	var (
		dateStr string
		isEmpty bool
	)
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT p.roast_date
			FROM purchases as p
			INNER JOIN coffees as c
				ON p.coffee_id = c.id
			WHERE c.name = :coffeeName
			ORDER BY p.id DESC
			LIMIT 1
		`,
			sql.Named("coffeeName", coffeeName),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve roast date: %w", err)
		}
		defer rows.Close()

		// Check if no rows available
		if ok := rows.Next(); !ok {
			isEmpty = true
			return nil
		}

		var val interface{}
		if err := rows.Scan(&val); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		// Deal with possible NULL values
		if v := reflect.ValueOf(val); v.Kind() == reflect.String {
			dateStr = val.(string)
		} else {
			isEmpty = true
			return nil
		}

		return nil
	}); err != nil {
		return date{}, fmt.Errorf("buna: input_suggestions: getLastCoffeeRoastDate transaction failed: %w", err)
	}

	if isEmpty {
		return date{}, nil
	}

	roastDate, err := createDateFromDateString(dateStr)
	if err != nil {
		return date{}, fmt.Errorf("buna: input_suggestions: failed to convert dateStr into date: %w", err)
	}

	return roastDate, nil
}

// limit determines the number of strings in the returned slice.
func (s *SQLiteDB) getMostRecentlyUsedBrewingMethodNames(ctx context.Context, limit int) ([]string, error) {
	var names []string
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT m.name 
			FROM brewings as b
			INNER JOIN brewing_methods as m
				ON b.method_id = m.id
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve brewing method name rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
			}

			names = append(names, name)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getMostRecentlyUsedBrewingMethodNames transaction failed: %w", err)
	}

	return names, nil
}

// limit determines the number of strings in the returned slice.
func (s *SQLiteDB) getMostRecentlyUsedCoffeeGrinderNames(ctx context.Context, limit int) ([]string, error) {
	var names []string
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT g.name 
			FROM brewings as b
			INNER JOIN grinders as g
				ON b.grinder_id = g.id
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve coffee grinder name rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
			}

			names = append(names, name)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getMostRecentlyUsedCoffeeGrinderNames transaction failed: %w", err)
	}

	return names, nil
}

// limit determines the number of strings in the returned slice.
// Weight is in grams.
func (s *SQLiteDB) getMostRecentlyUsedCoffeeWeights(ctx context.Context, brewingMethodName string, coffeeGrinderName string, limit int) ([]int, error) {
	var weights []int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT b.coffee_grams
			FROM brewings as b
			INNER JOIN brewing_methods as m
				ON b.method_id = m.id
			INNER JOIN grinders as g
				ON b.grinder_id = g.id
			WHERE m.name = :brewingMethodName AND g.name = :coffeeGrinderName
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("brewingMethodName", brewingMethodName),
			sql.Named("coffeeGrinderName", coffeeGrinderName),
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve coffee weight rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var weight int
			if err := rows.Scan(&weight); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
			}

			weights = append(weights, weight)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getMostRecentlyUsedCoffeeWeights transaction failed: %w", err)
	}

	return weights, nil
}

// limit determines the number of strings in the returned slice.
// Weight is in grams.
func (s *SQLiteDB) getMostRecentlyUsedWaterWeights(ctx context.Context, brewingMethodName string, coffeeGrinderName string, limit int) ([]int, error) {
	var weights []int
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT b.water_grams
			FROM brewings as b
			INNER JOIN brewing_methods as m
				ON b.method_id = m.id
			INNER JOIN grinders as g
				ON b.grinder_id = g.id
			WHERE m.name = :brewingMethodName AND g.name = :coffeeGrinderName
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("brewingMethodName", brewingMethodName),
			sql.Named("coffeeGrinderName", coffeeGrinderName),
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve coffee weight rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var weight int
			if err := rows.Scan(&weight); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
			}

			weights = append(weights, weight)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getMostRecentlyUsedCoffeeWeights transaction failed: %w", err)
	}

	return weights, nil
}

// limit determines the number of strings in the returned slice.
func (s *SQLiteDB) getRoastersByCoffeeName(ctx context.Context, name string, limit int) ([]string, error) {
	var roasters []string
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT roaster 
			FROM coffees
			WHERE name = :name
			ORDER BY id DESC
			LIMIT :limit
		`,
			sql.Named("name", name),
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to retrieve coffee roaster rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var roaster string
			if err := rows.Scan(&roaster); err != nil {
				return fmt.Errorf("buna: input_suggestions: failed to scan row: %w", err)
			}

			roasters = append(roasters, roaster)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: input_suggestions: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: input_suggestions: getRoastersByCoffeeName transaction failed: %w", err)
	}

	return roasters, nil
}

func removeStrDuplicates(strings []string) []string {
	keys := make(map[string]bool)
	var res []string

	for _, entry := range strings {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			res = append(res, entry)
		}
	}

	return res
}
