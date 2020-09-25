package buna

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
)

type brewingMethod struct {
	name string
}

func addBrewingMethod(ctx context.Context, db DB) error {
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

func retrieveBrewingMethod(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Retrieve brewing methods ordered by last added",
	}

	fmt.Println("Retrieving brewing methods (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: brewing_method: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: brewing_method: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveBrewingMethodSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: brewing_method: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveBrewingMethodSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
		if err := displayBrewingMethodsByLastAdded(ctx, db); err != nil {
			return fmt.Errorf("buna: brewing_method: failed to display brewing methods by last added: %w", err)
		}
	default:
		return errors.New("buna: brewing_method: invalid retrieve selection")
	}
	return nil
}

// Promts user for an optional limit.
func displayBrewingMethodsByLastAdded(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 20
	const maxDisplayAmount = 60

	fmt.Println("Displaying brewing methods by last added (Enter # to quit):")

	fmt.Print("Enter a limit for the number of brewing methods to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if limit == 0 {
		limit = defaultDisplayAmount
	}

	brewingMethods, err := db.getBrewingMethodsByLastAdded(ctx, limit)
	if err != nil {
		return fmt.Errorf("buna: brewing_method: failed to get brewing methods by last added: %w", err)
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{"Name"})

	for _, brewingMethod := range brewingMethods {
		t.AppendRow(table.Row{brewingMethod.name})
		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: brewing_method: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}
