package buna

import (
	"context"
	"database/sql"
	"fmt"
)

// limit determines the number of strings in the returned slice.
func (s *SQLiteDB) getMostRecentBrewedCoffeeNames(ctx context.Context, limit int) ([]string, error) {
	var names []string
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT DISTINCT c.name 
			FROM brewings as b
			INNER JOIN coffees as c
				ON b.coffee_id = c.id
			ORDER BY b.id DESC
			LIMIT :limit
		`,
			sql.Named("limit", limit),
		)
		if err != nil {
			return fmt.Errorf("buna: coffee: failed to retrieve coffee name rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("buna: coffee: failed to scan row: %w", err)
			}

			names = append(names, name)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: coffee: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: coffee: getMostRecentBrewedCoffeeNames transaction failed: %w", err)
	}

	return names, nil
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
			return fmt.Errorf("buna: coffee: failed to retrieve brewing method name rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return fmt.Errorf("buna: coffee: failed to scan row: %w", err)
			}

			names = append(names, name)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: coffee: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: coffee: getMostRecentlyUsedBrewingMethodNames transaction failed: %w", err)
	}

	return names, nil
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
			return fmt.Errorf("buna: coffee: failed to retrieve coffee roaster rows: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var roaster string
			if err := rows.Scan(&roaster); err != nil {
				return fmt.Errorf("buna: coffee: failed to scan row: %w", err)
			}

			roasters = append(roasters, roaster)
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("buna: coffee: failed to scan last row: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("buna: coffee: getRoastersByCoffeeName transaction failed: %w", err)
	}

	return roasters, nil
}
