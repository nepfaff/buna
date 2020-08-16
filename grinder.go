package buna

import (
	"context"
	"database/sql"
	"fmt"
)

type grinder struct {
	name    string
	company string
}

func addGrinder(ctx context.Context, db DB) error {
	quitStr := "#"

	fmt.Println("Adding new coffee grinder (Enter # to quit):")
	fmt.Print("Enter grinder name: ")
	name, quit := validateStrInput(quitStr, false)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter grinder's company name: ")
	company, quit := validateStrInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	grinder := grinder{
		name:    name,
		company: company,
	}

	if err := db.insertGrinder(ctx, grinder); err != nil {
		return fmt.Errorf("buna: grinder: failed to insert coffee grinder: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertGrinder(ctx context.Context, grinder grinder) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO grinders(name, company)
			VALUES (:name, :company)
		`,
			sql.Named("name", grinder.name),
			sql.Named("company", grinder.company),
		); err != nil {
			return fmt.Errorf("buna: grinder: failed to insert coffee grinder into db: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: grinder: transaction failed: %w", err)
	}
	return nil
}
