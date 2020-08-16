package buna

import (
	"context"
	"database/sql"
	"errors"
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

func retrieveCoffee(ctx context.Context, db DB) error {
	quitStr := "#"
	quitMsg := "Quit"

	options := map[int]string{
		0: "Retrieve coffee by name",
		1: "Retrieve coffees ordered by last added",
		2: "Retrieve coffees ordered alphabetically",
		3: "Retrieve coffees by origin",
		4: "Retrieve coffees by roaster",
		5: "Retrieve coffees ordered alphabetically",
		6: "Retrieve coffees by processing method",
		7: "Retrieve decaf coffees ordered by last added",
	}

	fmt.Println("Retrieving coffee (Enter # to quit):")
	displayIntOptions(options)

	selection, quit := getIntSelection(options, quitStr)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveCoffeeSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: coffee: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveCoffeeSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
	case 1:
	case 2:
	case 3:
	case 4:
	case 5:
	case 6:
	case 7:
	default:
		return errors.New("buna: coffee: invalid retrieve selection")
	}
	return nil
}
