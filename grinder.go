package buna

import (
	"context"
	"fmt"
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
