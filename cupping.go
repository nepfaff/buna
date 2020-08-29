package buna

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"
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

		coffeeName, quit, err := getCoffeeNameWithSuggestions(ctx, db, quitStr)
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
