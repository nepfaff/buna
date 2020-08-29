package buna

import (
	"context"
	"fmt"
	"time"
)

type coffeePurchase struct {
	coffeeName    string
	coffeeRoaster string
	boughtDate    string
	roastDate     string
}

func addCoffeePurchase(ctx context.Context, db DB) error {
	fmt.Println("Adding new coffee purchase (Enter # to quit):")

	fmt.Print("Do you want to create a new coffee first? (true or false): ")
	createCoffee, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var name, roaster string
	if createCoffee {
		addedCoffee, err := addCoffee(ctx, db)
		if err != nil {
			return fmt.Errorf("buna: coffee_purchases: failed to create new coffee: %w", err)
		}

		name = addedCoffee.name
		roaster = addedCoffee.roaster

		fmt.Println("\nAdding new coffee purchase for the just added coffee (Enter # to quit):")
	} else {
		fmt.Print("Enter coffee name: ")
		name, quit = validateStrInput(quitStr, false, nil, nil)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		fmt.Print("Enter roaster/producer name: ")
		roaster, quit = validateStrInput(quitStr, false, nil, nil)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}
	}

	boughtDate, quit := getDateInput(quitStr, false, "Enter ? of purchase or ? of arrival if bought online: ", []date{
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day()},
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day() - 1},
	})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	roastDate, quit := getDateInput(quitStr, true, "Enter roast ?: ", []date{})
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
