package buna

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
)

type cupping struct {
	date          string
	durationMin   int
	cuppedCoffees []cuppedCoffee
	notes         string
}

type cuppedCoffee struct {
	name    string
	roaster string
	rank    int
	notes   string
}

func addCupping(ctx context.Context, db DB) error {
	fmt.Println("Adding new cupping (Enter # to quit):")

	cuppingDate, quit := getDateInput(quitStr, false, "Enter cupping ?: ", []date{
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day()},
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day() - 1},
	})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter cupping duration in minutes: ")
	cuppingDurationMin, quit := validateIntInput(quitStr, false, 1, math.MaxInt64, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	cuppingNotes, quit := getNotes(quitStr, false, "general cupping")
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter number of coffees in this cupping: ")
	coffeeNumber, quit := validateIntInput(quitStr, false, 2, 30, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	cuppedCoffees := make([]cuppedCoffee, coffeeNumber)
	for i := 0; i < coffeeNumber; i++ {
		fmt.Println("\nAdding " + strconv.Itoa(i+1) + ". cupped coffee (Enter # to quit):")

		coffeeName, quit, err := getCoffeeNameWithSuggestions(ctx, db, quitStr, false)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee name: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		coffeeRoaster, quit, err := getCoffeeRoasterWithSuggestions(ctx, db, quitStr, coffeeName)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee roaster: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		fmt.Print("Enter this coffees rank (1 = highest): ")
		coffeeRank, quit := validateIntInput(quitStr, false, 1, coffeeNumber, nil)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		coffeeNotes, quit := getNotes(quitStr, false, "cupped coffee")
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		cuppedCoffees[i] = cuppedCoffee{
			name:    coffeeName,
			roaster: coffeeRoaster,
			rank:    coffeeRank,
			notes:   coffeeNotes,
		}
	}

	newCupping := cupping{
		date:          createDateString(cuppingDate),
		durationMin:   cuppingDurationMin,
		cuppedCoffees: cuppedCoffees,
		notes:         cuppingNotes,
	}

	if err := db.insertCupping(ctx, newCupping); err != nil {
		return fmt.Errorf("buna: cupping: failed to insert cupping: %w", err)
	}

	fmt.Println("Added cupping successfully")
	return nil
}

func retrieveCupping(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Retrieve cuppings ordered by last added",
	}

	fmt.Println("Retrieving cuppings (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: cupping: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: cupping: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveCuppingSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: cupping: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveCuppingSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
		if err := displayCuppingsByLastAdded(ctx, db); err != nil {
			return fmt.Errorf("buna: cupping: failed to display cuppings by last added: %w", err)
		}
	default:
		return errors.New("buna: cupping: invalid retrieve selection")
	}
	return nil
}

// Promts user for an optional limit.
func displayCuppingsByLastAdded(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 3
	const maxDisplayAmount = 10
	const maxNoteFieldWidth = 100

	fmt.Println("Displaying cuppings by last added (Enter # to quit):")
	fmt.Print("Enter a limit for the number of cuppings to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if limit == 0 {
		limit = defaultDisplayAmount
	}

	cuppings, err := db.getCuppingsByLastAdded(ctx, limit)
	if err != nil {
		return fmt.Errorf("buna: cupping: failed to get cuppings by last added: %w", err)
	}

	if len(cuppings) == 0 {
		fmt.Println("No cuppings to display")
		return nil
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: cupping: failed to get terminal width: %w", err)
	}

	for _, cupping := range cuppings {
		// Cupping table
		t := table.NewWriter()

		t.AppendHeader(table.Row{"Date", "Duration (min)", "General notes"})

		cuppingNotes := splitTextIntoField(cupping.notes, maxNoteFieldWidth)

		t.AppendRow(table.Row{cupping.date, cupping.durationMin, cuppingNotes})

		t.SetAllowedRowLength(terminalWidth)
		t.SetOutputMirror(os.Stdout)
		t.Render()

		// Cupped coffees table
		t = table.NewWriter()

		t.AppendHeader(table.Row{"Coffee name", "Rank (1 = best)", "Coffee notes"})

		for _, cuppedCoffee := range cupping.cuppedCoffees {
			cuppedCoffeeNotes := splitTextIntoField(cuppedCoffee.notes, maxNoteFieldWidth)

			t.AppendRow(table.Row{cuppedCoffee.name, cuppedCoffee.rank, cuppedCoffeeNotes})
			t.AppendSeparator()
		}

		t.SetAllowedRowLength(terminalWidth)
		t.SetOutputMirror(os.Stdout)
		t.Render()
		fmt.Println()
	}

	return nil
}
