package buna

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

type selection struct {
	category category
	index    int
}

type category int

const (
	create category = iota
	retrieve
	// statistics
	control
)

var (
	categoryRefs = map[category]string{
		create:   "A",
		retrieve: "B",
		control:  "E",
	}
	options = map[category]map[int]string{
		create: map[int]string{
			0: "New brewing",
			1: "New cupping",
			2: "New coffee purchase",
			3: "New coffee",
			4: "New brewing method",
			5: "New grinder",
		},
		retrieve: map[int]string{
			0: "Retrive brewing",
			1: "Retrieve cupping",
			2: "Retrieve coffee purchase",
			3: "Retrieve coffee",
			4: "Retrieve brewing method",
			5: "Retrieve grinder",
		},
		control: map[int]string{
			0: "Quit",
			1: "Display options",
		},
	}
)

func Run(ctx context.Context, db DB) error {
	displayOptions()

	var selection selection
	for {
		selection = getSelection()

		// Check for Quit option
		if selection.category == control && selection.index == 0 {
			break
		}

		if err := runSelection(ctx, selection, db); err != nil {
			return fmt.Errorf("buna: ui: failed to run the selection: %w", err)
		}
	}

	fmt.Println("Bye, keep enjoying your coffee!")
	return nil
}

func displayOptions() {
	t := table.NewWriter()

	var header table.Row
	for i := 0; i < len(options); i++ {
		header = append(header, "Option")
		header = append(header, "Description")
	}
	t.AppendHeader(header)

	rows := make([]table.Row, getLongestCategoryLength())
	for i := range rows {
		for j := 0; j < len(options); j++ {
			cat := category(j)
			if _, ok := options[cat][i]; ok {
				rows[i] = append(rows[i], categoryRefs[cat]+strconv.Itoa(i))
				rows[i] = append(rows[i], options[cat][i])
			}
		}
	}
	t.AppendRows(rows)

	t.SetOutputMirror(os.Stdout)
	t.Render()
}

func getSelection() selection {
	retry := func() {
		fmt.Println("Invalid option. The following options are available:")
		displayOptions()
	}

	for {
		fmt.Print("Enter option: ")
		var input string
		fmt.Scanln(&input)
		input = strings.ToUpper(input)

		if len(input) != 2 {
			retry()
			continue
		}

		cat, err := getCategoryByString(input[:1])
		if err != nil {
			retry()
			continue
		}

		idx, err := strconv.Atoi(input[1:])
		if err != nil {
			retry()
			continue
		}

		if _, ok := options[cat][idx]; !ok {
			retry()
			continue
		}

		return selection{
			category: cat,
			index:    idx,
		}
	}
}

func runSelection(ctx context.Context, selection selection, db DB) error {
	switch selection.category {
	case create:
		switch selection.index {
		case 0:
		case 1:
		case 2:
			if err := addCoffeePurchase(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee purchase: %w", err)
			}
		case 3:
			if err := addCoffee(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee: %w", err)
			}
		case 4:
			if err := addBrewingMethod(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee brewing method: %w", err)
			}
		case 5:
			if err := addGrinder(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee grinder: %w", err)
			}
		}
	case retrieve:
		switch selection.index {
		case 0:
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
		}
	case control:
		switch selection.index {
		case 0:
			// Special case
			// Already handled in Run()
		case 1:
			displayOptions()
		}
	default:
		return errors.New("buna: ui: invalid category")
	}
	return nil
}

func getLongestCategoryLength() int {
	var longest int
	for _, cat := range options {
		current := len(cat)
		if current > longest {
			longest = current
		}
	}

	return longest
}

func getCategoryByString(str string) (category, error) {
	for cat, val := range categoryRefs {
		if val == str {
			return cat, nil
		}
	}
	return 0, errors.New("buna: ui: string does not correspond to category")
}
