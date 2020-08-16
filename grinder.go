package buna

import (
	"context"
	"database/sql"
	"fmt"
)

type grinder struct {
	name            string
	company         string
	maxGrindSetting int
}

func addGrinder(ctx context.Context, db DB) error {
	quitStr := "#"
	quitMsg := "Quit"

	fmt.Println("Adding new coffee grinder (Enter # to quit):")
	fmt.Print("Enter grinder name: ")
	name, quit := validateStrInput(quitStr, false, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter grinder's company name: ")
	company, quit := validateStrInput(quitStr, true, nil, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter the maximum grind setting (Integer): ")
	maxGrindSetting, quit := validateIntInput(quitStr, true, 0, 100, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	grinder := grinder{
		name:            name,
		company:         company,
		maxGrindSetting: maxGrindSetting,
	}

	if err := db.insertGrinder(ctx, grinder); err != nil {
		return fmt.Errorf("buna: grinder: failed to insert coffee grinder: %w", err)
	}

	fmt.Println("Added coffee grinder successfully")
	return nil
}

func (s *SQLiteDB) insertGrinder(ctx context.Context, grinder grinder) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO grinders(name, company, max_grind_setting)
			VALUES (:name, :company, :maxGrindSetting)
		`,
			sql.Named("name", grinder.name),
			sql.Named("company", grinder.company),
			sql.Named("maxGrindSetting", grinder.maxGrindSetting),
		); err != nil {
			return fmt.Errorf("buna: grinder: failed to insert coffee grinder into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE grinders
			SET company = NULLIF(company, ""),
				max_grind_setting = NULLIF(max_grind_setting, 0)
			WHERE name = :name
		`,
			sql.Named("name", grinder.name),
		); err != nil {
			return fmt.Errorf("buna: grinder: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: grinder: transaction failed: %w", err)
	}
	return nil
}
