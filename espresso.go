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

func addEspressoDialingIn(ctx context.Context, db DB) error {
	fmt.Println("Adding new espresso dialing in (Enter # to quit):")
	dialingInDate, quit := getDateInput(quitStr, false, "Enter dialing in ?: ", []date{
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day()},
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day() - 1},
	})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	coffeeName, quit, err := getCoffeeNameWithSuggestions(ctx, db, quitStr, false)
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get coffee name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	coffeeRoaster, quit, err := getCoffeeRoasterWithSuggestions(ctx, db, quitStr, coffeeName)
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get coffee roaster: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	brewingMethodName, quit, err := getBrewingMethodNameWithSuggestions(ctx, db, quitStr, false)
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get brewing method name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	roastDate, quit, err := getCoffeeRoastDateWithSuggestions(ctx, db, quitStr, coffeeName)
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get coffee roast date: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	grinderName, quit, err := getCoffeeGrinderNameWithSuggestions(ctx, db, quitStr, false)
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get coffee grinder name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var (
		finishedDialingIn bool
		previousEspressos []brewing
	)
	espressoCount := 1
	for !finishedDialingIn {
		fmt.Printf("Entering %v. espresso (Enter # to save the previous espressos and quit):\n", espressoCount)

		grindSetting, quit := getCoffeeGrindSettingWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		totalBrewingTimeSec, quit := getTotalCoffeeBrewingTimeSecWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		coffeeGrams, quit, err := getCoffeeWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName, false)
		if err != nil {
			return fmt.Errorf("buna: espresso: failed to get coffee weight: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		waterGrams, quit, err := getWaterWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName, false)
		if err != nil {
			return fmt.Errorf("buna: espresso: failed to get water weight: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		rating, quit := getCoffeeRatingWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		recommendedGrindSettingAdjustment, quit := getRecommendedGrindSettingAdjustmentWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		recommendedCoffeeWeightAdjustmentGrams, quit := getRecommendedCoffeeWeightAdjustmentGramsWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		notes, quit := getNotes(quitStr, true, "espresso")
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		espresso := brewing{
			date:                                   createDateString(dialingInDate),
			coffeeName:                             coffeeName,
			coffeeRoaster:                          coffeeRoaster,
			brewingMethodName:                      brewingMethodName,
			roastDate:                              createDateString(roastDate),
			grinderName:                            grinderName,
			grindSetting:                           grindSetting,
			totalBrewingTimeSec:                    totalBrewingTimeSec,
			coffeeGrams:                            coffeeGrams,
			waterGrams:                             waterGrams,
			v60FilterType:                          "",
			rating:                                 rating,
			recommendedGrindSettingAdjustment:      recommendedGrindSettingAdjustment,
			recommendedCoffeeWeightAdjustmentGrams: recommendedCoffeeWeightAdjustmentGrams,
			notes:                                  notes,
		}

		if err := db.insertBrewing(ctx, espresso); err != nil {
			return fmt.Errorf("buna: espresso: failed to insert espresso brewing: %w", err)
		}
		previousEspressos = append(previousEspressos, espresso)

		// Display espresso that was just entered
		if err := displayPreviousDialingInEspressos([]brewing{espresso}); err != nil {
			return fmt.Errorf("buna: espresso: failed to display disaling in espressos that was just entered: %w", err)
		}

		options := map[int]string{
			0: "Enter next espresso",
			1: "Display all previous espressos from this dialing in and enter next one",
			2: "Finish dialing in",
		}

		if err := displayIntOptions(options); err != nil {
			return fmt.Errorf("buna: espresso: failed to display int dialing in options: %w", err)
		}

		selection, quit, err := getIntSelection(options, quitStr)
		if err != nil {
			return fmt.Errorf("buna: espresso: failed to get int dialing in selection: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		switch selection {
		case 0:
			continue
		case 1:
			if err := displayPreviousDialingInEspressos(previousEspressos); err != nil {
				return fmt.Errorf("buna: espresso: failed to display previous disaling in espressos: %w", err)
			}
			continue
		case 2:
			finishedDialingIn = true
		default:
			return errors.New("buna: espresso: invalid dialing in selection")
		}
	}

	fmt.Println("Added espresso dialing in successfully")
	return nil
}

func displayPreviousDialingInEspressos(espressos []brewing) error {
	const maxNoteFieldWidth = 70

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Grind\nSetting",
		"Time\n(s)",
		"Coffee\nWeight\n(g)",
		"Water\nWeight\n(g)",
		"Recommended\nGrind\nAdjustment",
		"Recommended\nCoffee\nAdjustment (g)",
		"Notes",
		"Rating",
	})

	for _, espresso := range espressos {
		notes := splitTextIntoField(espresso.notes, maxNoteFieldWidth)

		row := table.Row{
			espresso.grindSetting,
			espresso.totalBrewingTimeSec,
			espresso.coffeeGrams,
			espresso.waterGrams,
			espresso.recommendedGrindSettingAdjustment,
			espresso.recommendedCoffeeWeightAdjustmentGrams,
			notes,
			espresso.rating,
		}

		t.AppendRow(row)
		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: espresso: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}
