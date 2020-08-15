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

	fmt.Println("Adding new coffee purchase (Enter # to quit):")
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

	fmt.Print("Enter year of purchase or year of arrival if bought online: ")
	boughtYear, quit := validateYearInput(quitStr, false)
	if quit {
		fmt.Println("Quit")
		return nil
	}
	fmt.Print("Enter month of purchase or month of arrival if bought online: ")
	boughtMonth, quit := validateMonthInput(quitStr, false)
	if quit {
		fmt.Println("Quit")
		return nil
	}
	fmt.Print("Enter day of purchase or day of arrival if bought online: ")
	boughtDay, quit := validateDayInput(quitStr, false, boughtMonth)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	fmt.Print("Enter roast year: ")
	roastYear, quit := validateYearInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}
	fmt.Print("Enter roast month: ")
	roastMonth, quit := validateMonthInput(quitStr, true)
	if quit {
		fmt.Println("Quit")
		return nil
	}
	fmt.Print("Enter roast day: ")
	roastDay, quit := validateDayInput(quitStr, true, roastMonth)
	if quit {
		fmt.Println("Quit")
		return nil
	}

	coffeePurchase := coffeePurchase{
		coffeeName:    name,
		coffeeRoaster: roaster,
		boughtDate:    createDateString(boughtYear, boughtMonth, boughtDay),
		roastDate:     createDateString(roastYear, roastMonth, roastDay),
	}

	if err := db.insertCoffeePurchase(ctx, coffeePurchase); err != nil {
		return fmt.Errorf("buna: coffee_purchase: failed to insert coffee_purchase: %w", err)
	}
	return nil
}

func (s *SQLiteDB) insertCoffeePurchase(ctx context.Context, coffeePurchase coffeePurchase) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var coffeeID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM coffees
			WHERE name = :coffeeName AND roaster = :coffeeRoaster
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

		return nil
	}); err != nil {
		return fmt.Errorf("buna: coffee_purchase: transaction failed: %w", err)
	}
	return nil
}
