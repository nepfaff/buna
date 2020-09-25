package buna

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
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
			return fmt.Errorf("buna: coffee_purchases: failed to create new coffee_purchases: %w", err)
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

func retrieveCoffeePurchase(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Retrieve coffee purchases ordered by last added",
	}

	fmt.Println("Retrieving coffee purchase (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: coffee_purchases: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: coffee_purchases: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveCoffeePurchaseSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: coffee_purchases: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveCoffeePurchaseSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
		if err := displayCoffeePurchasesByLastAdded(ctx, db); err != nil {
			return fmt.Errorf("buna: coffee_purchases: failed to display coffee purchases by last added: %w", err)
		}
	default:
		return errors.New("buna: coffee_purchases: invalid retrieve selection")
	}
	return nil
}

// Promts user for an optional limit.
func displayCoffeePurchasesByLastAdded(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 20
	const maxDisplayAmount = 60

	fmt.Println("Displaying coffee purchases by last added (Enter # to quit):")

	fmt.Print("Enter a limit for the number of coffee purchases to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if limit == 0 {
		limit = defaultDisplayAmount
	}

	coffeePurchases, err := db.getCoffeePurchasesByLastAdded(ctx, limit)
	if err != nil {
		return fmt.Errorf("buna: coffee_purchases: failed to get coffee purchases by last added: %w", err)
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Coffee\nName",
		"Coffee\nRoaster",
		"Bought\nDate",
		"Roast\nDate",
	})

	for _, coffeePurchase := range coffeePurchases {
		row := table.Row{
			coffeePurchase.coffeeName,
			coffeePurchase.coffeeRoaster,
			coffeePurchase.boughtDate,
			coffeePurchase.roastDate,
		}

		t.AppendRow(row)
		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: coffee_purchases: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}
