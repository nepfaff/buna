package buna

import (
	"context"
	"fmt"
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
