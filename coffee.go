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

	fmt.Println("Adding new coffee (Enter # to quit):")
	fmt.Print("Enter coffee name: ")
	name, quit := validateStrInput(quitStr, false)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter roaster/producer name: ")
	roaster, quit := validateStrInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter region: ")
	region, quit := validateStrInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter variety: ")
	variety, quit := validateStrInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter method: ")
	method, quit := validateStrInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Is decaf (true or false): ")
	decaf, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
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
		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee: transaction failed: %w", err)
	}
	return nil
}
