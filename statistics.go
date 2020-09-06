package buna

import (
	"context"
	"errors"
	"fmt"
)

func getAverageBrewingRating(ctx context.Context, db DB) error {
	fmt.Println("Getting average brewing rating (Enter # to quit):")

	fmt.Print("Add filters (true or false): ")
	showOptionalOptions, quit := validateBoolInput(quitStr, true)
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	var (
		brewingMethodName, v60FilterType, coffeeName, coffeeRoaster, grinderName string
		err                                                                      error
	)
	if showOptionalOptions {
		brewingMethodName, quit, err = getBrewingMethodNameWithSuggestions(ctx, db, quitStr, true)
		if err != nil {
			return fmt.Errorf("buna: statistics: failed to get brewing method name: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		if brewingMethodName == "v60" || brewingMethodName == "V60" {
			v60FilterType, quit = getV60FilterTypeWithSuggestions(quitStr)
			if quit {
				fmt.Println(quitMsg)
				return nil
			}
		}

		coffeeName, quit, err = getCoffeeNameWithSuggestions(ctx, db, quitStr, true)
		if err != nil {
			return fmt.Errorf("buna: statistics: failed to get coffee name: %w", err)
		}
		if quit {
			fmt.Println(quitMsg)
			return nil
		}

		if coffeeName != "" {
			coffeeRoaster, quit, err = getCoffeeRoasterWithSuggestions(ctx, db, quitStr, coffeeName)
			if err != nil {
				return fmt.Errorf("buna: statistics: failed to get coffee roaster: %w", err)
			}
			if quit {
				fmt.Println(quitMsg)
				return nil
			}
		}

		grinderName, quit, err = getCoffeeGrinderNameWithSuggestions(ctx, db, quitStr, true)
		if err != nil {
			return fmt.Errorf("buna: statistics: failed to get coffee grinder name: %w", err)
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
		coffeeGrams:                            0,
		waterGrams:                             0,
		v60FilterType:                          v60FilterType,
		rating:                                 0,
		recommendedGrindSettingAdjustment:      "",
		recommendedCoffeeWeightAdjustmentGrams: 0,
		notes:                                  "",
	}

	averageRating, err := db.getAverageBrewingRating(ctx, brewingFilter)
	if err != nil {
		return fmt.Errorf("buna: statistics: failed to get the average brewing rating: %w", err)
	}
	if averageRating == 0 {
		fmt.Println("No brewings exist")
		return nil
	}

	fmt.Printf("The average brewing rating is %.1f/10\n", averageRating)

	return nil
}

func getTotalCountInDB(ctx context.Context, db DB) error {
	options := map[int]string{
		0: "Total brewings count",
		1: "Total coffees count",
		2: "Total cuppings count",
		3: "Total coffee purchases count",
		4: "Total brewing methods count",
		5: "Total coffee grinders count",
	}

	fmt.Println("Getting total count (Enter # to quit):")
	if err := displayIntOptions(options); err != nil {
		return fmt.Errorf("buna: statistics: failed to display int options: %w", err)
	}

	selection, quit, err := getIntSelection(options, quitStr)
	if err != nil {
		return fmt.Errorf("buna: statistics: failed to get int selection: %w", err)
	}
	if quit {
		fmt.Println(quitMsg)
		return nil
	}

	entity, err := getDBEntityFromSelection(ctx, selection)
	if err != nil {
		return fmt.Errorf("buna: statistics: failed to get the dbEntity selection: %w", err)
	}

	count, err := db.getTotalCount(ctx, entity)
	if err != nil {
		return fmt.Errorf("buna: statistics: failed to get the total count: %w", err)
	}

	fmt.Println("There are", count, dbEntityToName[entity], "in total")

	return nil
}

func getDBEntityFromSelection(ctx context.Context, selection int) (dbEntity, error) {
	var entity dbEntity
	switch selection {
	case 0:
		entity = brewings
	case 1:
		entity = coffees
	case 2:
		entity = cuppings
	case 3:
		entity = coffeePurchases
	case 4:
		entity = brewingMethods
	case 5:
		entity = grinders
	default:
		return 0, errors.New("buna: statistics: invalid dbEntity selection")
	}
	return entity, nil
}
