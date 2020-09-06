package buna

import (
	"context"
	"errors"
	"fmt"
)

func findTotalCountInDB(ctx context.Context, db DB) error {
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
