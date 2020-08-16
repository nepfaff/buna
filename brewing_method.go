package buna

import (
	"context"
	"database/sql"
	"fmt"
)

type brewingMethod struct {
	name string
}

func addBrewingMethod(ctx context.Context, db DB) error {
	quitStr := "#"
	quitMsg := "Quit"

	fmt.Println("Adding new coffee brewing method (Enter # to quit):")
	fmt.Print("Enter brewing method name: ")
	name, quit := validateStrInput(quitStr, false, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	brewingMethod := brewingMethod{
		name: name,
	}

	if err := db.insertBrewingMethod(ctx, brewingMethod); err != nil {
		return fmt.Errorf("buna: brewing_method: failed to insert brewingMethod: %w", err)
	}

	fmt.Println("Added coffee brewing method successfully")
	return nil
}

func (s *SQLiteDB) insertBrewingMethod(ctx context.Context, brewingMethod brewingMethod) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO brewing_methods(name)
			VALUES (:name)
		`,
			sql.Named("name", brewingMethod.name),
		); err != nil {
			return fmt.Errorf("buna: brewing_method: failed to insert coffee brewing method into db: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: brewing_method: transaction failed: %w", err)
	}
	return nil
}
