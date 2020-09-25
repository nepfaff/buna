package buna

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
)

type grinder struct {
	name            string
	company         string
	maxGrindSetting int
}

func addGrinder(ctx context.Context, db DB) error {
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

func retrieveGrinder(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Retrieve grinders ordered by last added",
	}

	fmt.Println("Retrieving grinders (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: grinder: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: grinder: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveGrinderSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: grinder: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveGrinderSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
		if err := displayGrindersByLastAdded(ctx, db); err != nil {
			return fmt.Errorf("buna: grinder: failed to display grinders by last added: %w", err)
		}
	default:
		return errors.New("buna: grinder: invalid retrieve selection")
	}
	return nil
}

// Promts user for an optional limit.
func displayGrindersByLastAdded(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 20
	const maxDisplayAmount = 60

	fmt.Println("Displaying grinders by last added (Enter # to quit):")

	fmt.Print("Enter a limit for the number of grinders to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if limit == 0 {
		limit = defaultDisplayAmount
	}

	grinders, err := db.getGrindersByLastAdded(ctx, limit)
	if err != nil {
		return fmt.Errorf("buna: grinder: failed to get grinders by last added: %w", err)
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Name",
		"Company",
		"Max Grind Setting",
	})

	for _, grinder := range grinders {
		t.AppendRow(table.Row{
			grinder.name,
			grinder.company,
			grinder.maxGrindSetting,
		})
		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: grinder: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}
