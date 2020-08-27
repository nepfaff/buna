package buna

import (
	"context"
	"database/sql"
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

	fmt.Print("Enter coffee name: ")
	coffeeSuggestions, err := db.getCoffeeNameSuggestions(ctx, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee suggestions: %w", err)
	}
	coffeeName, quit := validateStrInput(quitStr, false, nil, coffeeSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter roaster/producer name: ")
	roasterSuggestions, err := db.getRoastersByCoffeeName(ctx, coffeeName, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get roaster suggestions: %w", err)
	}
	coffeeRoaster, quit := validateStrInput(quitStr, false, nil, roasterSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter brewing method name: ")
	brewingMethodSuggestions, err := db.getMostRecentlyUsedBrewingMethodNames(ctx, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get brewing method suggestions: %w", err)
	}
	brewingMethodName, quit := validateStrInput(quitStr, false, nil, brewingMethodSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	roastDateSuggestion, err := db.getLastCoffeeRoastDate(ctx, coffeeName)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get roast date suggestions: %w", err)
	}
	// Check if returned empty date
	var roastDateSuggestions []date
	if roastDateSuggestion.year != 0 {
		roastDateSuggestions = append(roastDateSuggestions, roastDateSuggestion)
	}
	roastDate, quit := getDateInput(quitStr, true, "Enter roast ?: ", roastDateSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter coffee grinder name: ")
	grinderSuggestions, err := db.getMostRecentlyUsedCoffeeGrinderNames(ctx, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee grinder suggestions: %w", err)
	}
	grinderName, quit := validateStrInput(quitStr, false, nil, grinderSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter grind setting: ")
	// This assumes that every grinder has settings in the range 0 to 50
	// An improvement would be to look up the possible grind settings using the grinder name
	grindSetting, quit := validateIntInput(quitStr, false, 0, 50, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter the total brewing time in seconds: ")
	totalBrewingTimeSec, quit := validateIntInput(quitStr, false, 10, 1800, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter the coffee weight used in grams: ")
	coffeeWeightSuggestion, err := db.getMostRecentlyUsedCoffeeWeights(ctx, brewingMethodName, grinderName, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get coffee weight suggestions: %w", err)
	}
	coffeeGrams, quit := validateIntInput(quitStr, false, 5, 100, coffeeWeightSuggestion)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter the water weight used in grams: ")
	waterWeightSuggestion, err := db.getMostRecentlyUsedWaterWeights(ctx, brewingMethodName, grinderName, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get water weight suggestions: %w", err)
	}
	waterGrams, quit := validateIntInput(quitStr, false, 20, 2000, waterWeightSuggestion)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter v60 filter type if applicable: ")
	v60FilterType, quit := validateStrInput(quitStr, true, []string{"eu", "jp"}, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter your rating for this brew (1 <= x <= 10): ")
	rating, quit := validateIntInput(quitStr, true, 1, 10, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter recommended grind setting adjustment: ")
	recommendedGrindSettingAdjustment, quit := validateStrInput(quitStr, true, []string{"lower", "higher"}, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter recommended coffee weight adjustment in grams (-20 <= x <= 20): ")
	recommendedCoffeeWeightAdjustmentGrams, quit := validateIntInput(quitStr, true, -20, 20, nil)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	fmt.Print("Enter some notes about this brewing: ")
	notes, quit := validateStrInput(quitStr, true, nil, nil)
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

func (s *SQLiteDB) insertBrewing(ctx context.Context, brewing brewing) error {
	if err := s.TransactContext(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var coffeeID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM coffees
			WHERE name = :coffeeName AND (roaster = :coffeeRoaster OR (:coffeeRoaster = "" AND roaster IS NULL))
		`,
			sql.Named("coffeeName", brewing.coffeeName),
			sql.Named("coffeeRoaster", brewing.coffeeRoaster),
		).Scan(&coffeeID); err != nil {
			fmt.Println("Unable to link this brewing to an existing coffee. Please create a new coffee first and then try again.")
			return nil
		}

		var methodID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM brewing_methods
			WHERE name = :brewingMethodName
		`,
			sql.Named("brewingMethodName", brewing.brewingMethodName),
		).Scan(&methodID); err != nil {
			fmt.Println("Unable to link this brewing to an existing brewing method. Please create a new brewing method first and then try again.")
			return nil
		}

		var grinderID int
		if err := tx.QueryRowContext(ctx, `
			SELECT id
			FROM grinders
			WHERE name = :grinderName
		`,
			sql.Named("grinderName", brewing.grinderName),
		).Scan(&grinderID); err != nil {
			fmt.Println("Unable to link this brewing to an existing coffee grinder. Please create a new coffee grinder first and then try again.")
			return nil
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO brewings(
				coffee_id,
				method_id,
				grinder_id,
				date,
				roast_date,
				grind_setting,
				total_brewing_time_sec,
				water_grams,
				coffee_grams,
				v60_filter_type,
				rating,
				recommended_grind_setting_adjustment,
				recommended_coffee_weight_adjustment_grams,
				notes
			)
			VALUES (
				:coffeeID,
				:methodID,
				:grinderID,
				:date,
				:roastDate,
				:grindSetting,
				:totalBrewingTimeSec,
				:waterGrams,
				:coffeeGrams,
				:v60FilterType,
				:rating,
				:recommendedGrindSettingAdjustment,
				:recommendedCoffeeWeightAdjustmentGrams,
				:notes
			)
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("methodID", methodID),
			sql.Named("grinderID", grinderID),
			sql.Named("date", brewing.date),
			sql.Named("roastDate", brewing.roastDate),
			sql.Named("grindSetting", brewing.grindSetting),
			sql.Named("totalBrewingTimeSec", brewing.totalBrewingTimeSec),
			sql.Named("waterGrams", brewing.waterGrams),
			sql.Named("coffeeGrams", brewing.coffeeGrams),
			sql.Named("v60FilterType", brewing.v60FilterType),
			sql.Named("rating", brewing.rating),
			sql.Named("recommendedGrindSettingAdjustment", brewing.recommendedGrindSettingAdjustment),
			sql.Named("recommendedCoffeeWeightAdjustmentGrams", brewing.recommendedCoffeeWeightAdjustmentGrams),
			sql.Named("notes", brewing.notes),
		); err != nil {
			return fmt.Errorf("buna: brewing: failed to insert coffee brewing into db: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE brewings
			SET roast_date = NULLIF(roast_date, "0-00-00"),
				v60_filter_type = NULLIF(v60_filter_type, ""),
				rating = NULLIF(rating, 0),
				recommended_grind_setting_adjustment = NULLIF(recommended_grind_setting_adjustment, "")
			WHERE coffee_id = :coffeeID AND date = :date
		`,
			sql.Named("coffeeID", coffeeID),
			sql.Named("date", brewing.date),
		); err != nil {
			return fmt.Errorf("buna: brewing: failed to set null values: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("buna: brewing: transaction failed: %w", err)
	}
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

	fmt.Print("Enter brewing method name: ")
	brewingMethodSuggestions, err := db.getMostRecentlyUsedBrewingMethodNames(ctx, 5)
	if err != nil {
		return fmt.Errorf("buna: brewing: failed to get brewing method suggestions: %w", err)
	}
	brewingMethodName, quit := validateStrInput(quitStr, false, nil, brewingMethodSuggestions)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var v60FilterType string
	if brewingMethodName == "v60" || brewingMethodName == "V60" {
		fmt.Print("Enter v60 filter type (skip if want to display all): ")
		v60FilterType, quit = validateStrInput(quitStr, true, []string{"eu", "jp"}, nil)
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
		fmt.Print("Enter coffee name: ")
		coffeeSuggestions, err := db.getCoffeeNameSuggestions(ctx, 5)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee suggestions: %w", err)
		}
		coffeeName, quit = validateStrInput(quitStr, true, nil, coffeeSuggestions)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		if coffeeName != "" {
			fmt.Print("Enter roaster/producer name: ")
			roasterSuggestions, err := db.getRoastersByCoffeeName(ctx, coffeeName, 5)
			if err != nil {
				return fmt.Errorf("buna: brewing: failed to get roaster suggestions: %w", err)
			}
			coffeeRoaster, quit = validateStrInput(quitStr, true, nil, roasterSuggestions)
			if quit {
				fmt.Println(quitMsg)
				return nil
			}
		}

		fmt.Print("Enter coffee grinder name: ")
		grinderSuggestions, err := db.getMostRecentlyUsedCoffeeGrinderNames(ctx, 5)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee grinder suggestions: %w", err)
		}
		grinderName, quit = validateStrInput(quitStr, true, nil, grinderSuggestions)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		fmt.Print("Enter the coffee weight used in grams: ")
		coffeeWeightSuggestion, err := db.getMostRecentlyUsedCoffeeWeights(ctx, brewingMethodName, grinderName, 5)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get coffee weight suggestions: %w", err)
		}
		coffeeGrams, quit = validateIntInput(quitStr, true, 5, 100, coffeeWeightSuggestion)
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		fmt.Print("Enter the water weight used in grams: ")
		waterWeightSuggestion, err := db.getMostRecentlyUsedWaterWeights(ctx, brewingMethodName, grinderName, 5)
		if err != nil {
			return fmt.Errorf("buna: brewing: failed to get water weight suggestions: %w", err)
		}
		waterGrams, quit = validateIntInput(quitStr, true, 20, 2000, waterWeightSuggestion)
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

func splitTextIntoField(text string, maxFieldWidth int) string {
	for i := maxFieldWidth; i < len(text); i += maxFieldWidth {
		for string(text[i]) != " " {
			i--
		}
		text = text[:i] + "\n" + text[i:]
	}

	return text
}
