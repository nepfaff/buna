package buna

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
)

type brewing struct {
	date                                   string
	coffeeName                             string
	coffeeRoaster                          string
	brewingMethodName                      string
	roastDate                              string
	grinderName                            string
	grindSetting                           int
	totalBrewingTimeSec                    int
	coffeeGrams                            int
	waterGrams                             int
	v60FilterType                          string
	rating                                 int
	recommendedGrindSettingAdjustment      string
	recommendedCoffeeWeightAdjustmentGrams int
	notes                                  string
}

func addBrewing(ctx context.Context, db DB) error {
	fmt.Println("Adding new coffee brewing (Enter # to quit):")
	brewingDate, quit := getDateInput(quitStr, false, "Enter brewing ?: ", []date{
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day()},
		{year: time.Now().Year(), month: int(time.Now().Month()), day: time.Now().Day() - 1},
	})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

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

	brewingMethodName, quit, err := getBrewingMethodNameWithSuggestions(ctx, db, quitStr)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get brewing method name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	roastDate, quit, err := getCoffeeRoastDateWithSuggestions(ctx, db, quitStr, coffeeName)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee roast date: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	grinderName, quit, err := getCoffeeGrinderNameWithSuggestions(ctx, db, quitStr)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee grinder name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

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

	coffeeGrams, quit, err := getCoffeeWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee weight: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	waterGrams, quit, err := getWaterWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get water weight: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	v60FilterType, quit := getV60FilterTypeWithSuggestions(quitStr)
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

	notes, quit := getNotes(quitStr, true, "brewing")
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	brewing := brewing{
		date:                                   createDateString(brewingDate),
		coffeeName:                             coffeeName,
		coffeeRoaster:                          coffeeRoaster,
		brewingMethodName:                      brewingMethodName,
		roastDate:                              createDateString(roastDate),
		grinderName:                            grinderName,
		grindSetting:                           grindSetting,
		totalBrewingTimeSec:                    totalBrewingTimeSec,
		coffeeGrams:                            coffeeGrams,
		waterGrams:                             waterGrams,
		v60FilterType:                          v60FilterType,
		rating:                                 rating,
		recommendedGrindSettingAdjustment:      recommendedGrindSettingAdjustment,
		recommendedCoffeeWeightAdjustmentGrams: recommendedCoffeeWeightAdjustmentGrams,
		notes:                                  notes,
	}

	if err := db.insertBrewing(ctx, brewing); err != nil {
		return fmt.Errorf("buna: brewing: failed to insert coffee brewing: %w", err)
	}

	fmt.Println("Added coffee brewing successfully")
	return nil
}

func retrieveBrewing(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Retrieve brewing suggestions",
		1: "Retrieve brewing ordered by last added",
		2: "Retrieve brewing ordered by rating",
	}

	fmt.Println("Retrieving brewing (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: brewing: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if err := runRetrieveBrewingSelection(ctx, selection, db); err != nil {
		return fmt.Errorf("buna: brewing: failed to run the retrieve selection: %w", err)
	}

	return nil
}

func runRetrieveBrewingSelection(ctx context.Context, selection int, db DB) error {
	switch selection {
	case 0:
		if err := displayBrewingSuggestions(ctx, db); err != nil {
			return fmt.Errorf("buna: brewing: failed to display brewing suggestions: %w", err)
		}
	case 1:
		if err := displayBrewingsByLastAdded(ctx, db); err != nil {
			return fmt.Errorf("buna: brewing: failed to display brewings by last added: %w", err)
		}
	case 2:
	case 3:
	case 4:
	case 5:
	case 6:
	case 7:
	default:
		return errors.New("buna: brewing: invalid retrieve selection")
	}
	return nil
}

// Promts user for an optional limit and whether the brewing notes should be included.
func displayBrewingsByLastAdded(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 15
	const maxDisplayAmount = 50

	fmt.Println("Displaying brewings by last added (Enter # to quit):")
	fmt.Print("Enter a limit for the number of brewings to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Show brewing notes (true or false): ")
	showNotes, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	if limit == 0 {
		if showNotes {
			limit = defaultDisplayAmount / 3
		} else {
			limit = defaultDisplayAmount
		}
	}

	brewings, err := db.getBrewingsByLastAdded(ctx, limit)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffees by last added: %w", err)
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Date",
		"Coffee\nName",
		"Method",
		"Grind\nSetting",
		"Time\n(s)",
		"Coffee\nWeight\n(g)",
		"Water\nWeight\n(g)",
		"Rating",
		"Recommended\nGrind\nAdjustment",
		"Recommended\nCoffee\nAdjustment (g)",
		"V60\nFilter\nType",
		"Grinder",
		"Coffee\nRoaster",
		"Roast Date",
	})

	for _, brewing := range brewings {
		grinder := strings.ReplaceAll(brewing.grinderName, "(", "\n(")

		row := table.Row{
			brewing.date,
			brewing.coffeeName,
			brewing.brewingMethodName,
			brewing.grindSetting,
			brewing.totalBrewingTimeSec,
			brewing.coffeeGrams,
			brewing.waterGrams,
			brewing.rating,
			brewing.recommendedGrindSettingAdjustment,
			brewing.recommendedCoffeeWeightAdjustmentGrams,
			brewing.v60FilterType,
			grinder,
			brewing.coffeeRoaster,
			brewing.roastDate,
		}

		t.AppendRow(row)

		if showNotes {
			re := regexp.MustCompile(`([.?!:;]) ([A-Z])`)
			notes := "\n" + string(re.ReplaceAll([]byte(brewing.notes), []byte("${1}\n${2}")))

			t.AppendRow(table.Row{"\nNotes:", notes})
		}

		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}

func displayBrewingSuggestions(ctx context.Context, db DB) error {
	const defaultDisplayAmount = 6
	const maxDisplayAmount = 20
	const maxNoteFieldWidth = 50

	fmt.Println("Displaying brewing suggestions (Enter # to quit):")

	fmt.Print("Enter a limit for the number of suggestions to display: ")
	limit, quit := validateIntInput(quitStr, true, 1, maxDisplayAmount, []int{})
	if quit {
		fmt.Println(quitMsg)
		return nil
	}
	if limit == 0 {
		limit = defaultDisplayAmount
	}

	brewingMethodName, quit, err := getBrewingMethodNameWithSuggestions(ctx, db, quitStr)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get brewing method name: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var v60FilterType string
	if brewingMethodName == "v60" || brewingMethodName == "V60" {
		v60FilterType, quit = getV60FilterTypeWithSuggestions(quitStr)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}
	}

	fmt.Print("Show optional options (true or false): ")
	showOptionalOptions, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var (
		coffeeName, coffeeRoaster, grinderName string
		coffeeGrams, waterGrams                int
	)
	if showOptionalOptions {
		coffeeName, quit, err = getCoffeeNameWithSuggestions(ctx, db, quitStr)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee name: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		if coffeeName != "" {
			coffeeRoaster, quit, err = getCoffeeRoasterWithSuggestions(ctx, db, quitStr, coffeeName)
			if err != nil {
				return fmt.Errorf("buna: brewing: failed to get coffee roaster: %w", err)
			}
			if quit {
				fmt.Println(quitMsg)
				return nil
			}
		}

		grinderName, quit, err = getCoffeeGrinderNameWithSuggestions(ctx, db, quitStr)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee grinder name: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		coffeeGrams, quit, err = getCoffeeWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee weight: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		waterGrams, quit, err = getWaterWeightWithSuggestions(ctx, db, quitStr, brewingMethodName, grinderName)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get water weight: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}
	}

	brewingFilter := brewing{
		date:                                   "",
		coffeeName:                             coffeeName,
		coffeeRoaster:                          coffeeRoaster,
		brewingMethodName:                      brewingMethodName,
		roastDate:                              "",
		grinderName:                            grinderName,
		grindSetting:                           0,
		totalBrewingTimeSec:                    0,
		coffeeGrams:                            coffeeGrams,
		waterGrams:                             waterGrams,
		v60FilterType:                          v60FilterType,
		rating:                                 0,
		recommendedGrindSettingAdjustment:      "",
		recommendedCoffeeWeightAdjustmentGrams: 0,
		notes:                                  "",
	}

	suggestions, err := db.getBrewingSuggestions(ctx, limit, brewingFilter)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get brewing suggestions: %w", err)
	}

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
		"V60\nFilter\nType",
		"Coffee\nName",
		"Date",
		"Grinder",
	})

	for _, suggestion := range suggestions {
		notes := splitTextIntoField(suggestion.notes, maxNoteFieldWidth)
		grinder := strings.ReplaceAll(suggestion.grinderName, "(", "\n(")

		row := table.Row{
			suggestion.grindSetting,
			suggestion.totalBrewingTimeSec,
			suggestion.coffeeGrams,
			suggestion.waterGrams,
			suggestion.recommendedGrindSettingAdjustment,
			suggestion.recommendedCoffeeWeightAdjustmentGrams,
			notes,
			suggestion.rating,
			suggestion.v60FilterType,
			suggestion.coffeeName,
			suggestion.date,
			grinder,
		}

		t.AppendRow(row)
		t.AppendSeparator()
	}

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}
