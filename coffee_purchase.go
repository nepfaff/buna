package buna

import (
	"context"
	"database/sql"
	"fmt"
)

type coffeePurchase struct {
	coffeeName    string
	coffeeRoaster string
	boughtDate    string
	roastDate     string
}

func addCoffeePurchase(ctx context.Context, db DB) error {
	quitStr := "#"
	quitMsg := "Quit"

	fmt.Println("Adding new coffee purchase (Enter # to quit):")
	fmt.Print("Enter coffee name: ")
	name, quit := validateStrInput(quitStr, false, []string{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter roaster/producer name: ")
	roaster, quit := validateStrInput(quitStr, false, []string{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	boughtDate, quit := getDateInput(quitStr, false, "Enter ? of purchase or ? of arrival if bought online: ")
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	roastDate, quit := getDateInput(quitStr, true, "Enter roast ?: ")
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	coffeePurchase := coffeePurchase{
		coffeeName:    name,
		coffeeRoaster: roaster,
		boughtDate:    createDateString(boughtDate),
		roastDate:     createDateString(roastDate),
	}

	if err := db.insertCoffeePurchase(ctx, coffeePurchase); err != nil {
		return fmt.Errorf("buna: coffee_purchase: failed to insert coffee_purchase: %w", err)
	}

	fmt.Println("Added coffee pruchase successfully")
	return nil
}

func (s *SQLiteDB) insertCoffeePurchase(ctx context.Context, coffeePurchase coffeePurchase) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var coffeeID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM coffees
			WHERE name = :coffeeName AND (roaster = :coffeeRoaster OR (:coffeeRoaster = "" AND roaster IS NULL))
		`,
			sql.Named("coffeeName", coffeePurchase.coffeeName),
			sql.Named("coffeeRoaster", coffeePurchase.coffeeRoaster),
		).Scan(&coffeeID); err != nil {
			fmt.Println("Unable to link the purchased coffee to an existing coffee. Please create a new coffee first and then try again.")
			return nil
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO purchases(coffee_id, bought_date, roast_date)
			VALUES (:coffeeID, :boughtDate, :roastDate)
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("boughtDate", coffeePurchase.boughtDate),
			sql.Named("roastDate", coffeePurchase.roastDate),
		); err != nil {
			return fmt.Errorf("buna: coffee_purchase: failed to insert coffee purchase into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE purchases
			SET roast_date = NULLIF(roast_date, "0-00-00")
			WHERE coffee_id = :coffeeID AND bought_date = :boughtDate
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("boughtDate", coffeePurchase.boughtDate),
		); err != nil {
			return fmt.Errorf("buna: coffee_purchase: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee_purchase: transaction failed: %w", err)
	}
	return nil
}
