package buna

import (
	"context"
	"database/sql"
	"fmt"
)

type coffee struct {
	name    string
	roaster string
	region  string
	variety string
	method  string
	decaf   bool
}

func addCoffee(ctx context.Context, db DB) error {
	quitStr := "#"
	quitMsg := "Quit"

	fmt.Println("Adding new coffee (Enter # to quit):")
	fmt.Print("Enter coffee name: ")
	name, quit := validateStrInput(quitStr, false, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter roaster/producer name: ")
	roaster, quit := validateStrInput(quitStr, false, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter region: ")
	region, quit := validateStrInput(quitStr, true, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter variety: ")
	variety, quit := validateStrInput(quitStr, true, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter method: ")
	method, quit := validateStrInput(quitStr, true, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Is decaf (true or false): ")
	decaf, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	coffee := coffee{
		name:    name,
		roaster: roaster,
		region:  region,
		variety: variety,
		method:  method,
		decaf:   decaf,
	}

	if err := db.insertCoffee(ctx, coffee); err != nil {
		return fmt.Errorf("buna: coffee: failed to insert coffee: %w", err)
	}

	fmt.Println("Added coffee successfully")
	return nil
}

func (s *SQLiteDB) insertCoffee(ctx context.Context, coffee coffee) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO coffees(name, roaster, region, variety, method, decaf)
			VALUES (:name, :roaster, :region, :variety, :method, :decaf)
		`,
			sql.Named("name", coffee.name),
			sql.Named("roaster", coffee.roaster),
			sql.Named("region", coffee.region),
			sql.Named("variety", coffee.variety),
			sql.Named("method", coffee.method),
			sql.Named("decaf", coffee.decaf),
		); err != nil {
			return fmt.Errorf("buna: coffee: failed to insert coffee into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE coffees
			SET roaster = NULLIF(roaster, ""),
				region = NULLIF(region, ""),
				variety = NULLIF(variety, ""),
				method = NULLIF(method, ""),
				decaf = NULLIF(decaf, "")
			WHERE name = :name
		`,
			sql.Named("name", coffee.name),
		); err != nil {
			return fmt.Errorf("buna: coffee: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee: insertCoffee transaction failed: %w", err)
	}
	return nil
}

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
