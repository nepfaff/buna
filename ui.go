package buna

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"golang.org/x/crypto/ssh/terminal"
)

type selection struct {
	category category
	index    int
}

type category int

const (
	create category = iota
	retrieve
	// edit
	statistics
	control
)

const (
	quitStr = "#"
	quitMsg = "Quit"
)

var (
	categoryRefs = map[category]string{
		create:     "A",
		retrieve:   "B",
		statistics: "C",
		control:    "E",
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
		statistics: map[int]string{
			0: "Total count",
			1: "Average brewing rating",
		},
		control: map[int]string{
			0: "Quit",
			1: "Clear screen",
			2: "Display options",
		},
	}
)

// Used for clearing the terminal screen
var clear map[string]func() error

func init() {
	clear = make(map[string]func() error)
	clear["linux"] = func() error {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("buna: ui: failed to run linux clear terminal command: %w", err)
		}
		return nil
	}
	clear["windows"] = func() error {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("buna: ui: failed to run windows clear terminal command: %w", err)
		}
		return nil
	}
}

func Run(ctx context.Context, db DB) error {
	if err := displayOptions(); err != nil {
		return fmt.Errorf("buna: ui: failed to display main options: %w", err)
	}

	var selection selection
	var err error
	for {
		selection, err = getSelection()
		if err != nil {
			return fmt.Errorf("buna: ui: failed to get main selection: %w", err)
		}

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

func displayOptions() error {
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
			} else {
				rows[i] = append(rows[i], "")
				rows[i] = append(rows[i], "")
			}
		}
	}
	t.AppendRows(rows)

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("buna: ui: failed to get terminal width: %w", err)
	}
	t.SetAllowedRowLength(terminalWidth)

	t.SetOutputMirror(os.Stdout)
	t.Render()

	return nil
}

func getSelection() (selection, error) {
	retry := func() error {
		fmt.Println("Invalid option. The following options are available:")
		if err := displayOptions(); err != nil {
			return fmt.Errorf("buna: ui: failed to display main options: %w", err)
		}

		return nil
	}

	for {
		fmt.Print("Enter main option: ")
		var input string
		fmt.Scanln(&input)
		input = strings.ToUpper(input)

		if len(input) != 2 {
			if err := retry(); err != nil {
				return selection{}, fmt.Errorf("buna: ui: failed to display main options (retry): %w", err)
			}
			continue
		}

		cat, err := getCategoryByString(input[:1])
		if err != nil {
			if err := retry(); err != nil {
				return selection{}, fmt.Errorf("buna: ui: failed to display main options (retry): %w", err)
			}
			continue
		}

		idx, err := strconv.Atoi(input[1:])
		if err != nil {
			if err := retry(); err != nil {
				return selection{}, fmt.Errorf("buna: ui: failed to display main options (retry): %w", err)
			}
			continue
		}

		if _, ok := options[cat][idx]; !ok {
			if err := retry(); err != nil {
				return selection{}, fmt.Errorf("buna: ui: failed to display main options (retry): %w", err)
			}
			continue
		}

		return selection{
			category: cat,
			index:    idx,
		}, nil
	}
}

func runSelection(ctx context.Context, selection selection, db DB) error {
	switch selection.category {
	case create:
		switch selection.index {
		case 0:
			if err := addBrewing(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee brewing: %w", err)
			}
		case 1:
			if err := addCupping(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new cupping: %w", err)
			}
		case 2:
			if err := addCoffeePurchase(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to create new coffee purchase: %w", err)
			}
		case 3:
			if _, err := addCoffee(ctx, db); err != nil {
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
		default:
			return errors.New("buna: ui: invalid create index")
		}
	case retrieve:
		switch selection.index {
		case 0:
			if err := retrieveBrewing(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to retrieve brewing: %w", err)
			}
		case 1:
			if err := retrieveCupping(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to retrieve cupping: %w", err)
			}
		case 2:
		case 3:
			if err := retrieveCoffee(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to retrieve coffee: %w", err)
			}
		case 4:
		case 5:
		default:
			return errors.New("buna: ui: invalid retrieve index")
		}
	case statistics:
		switch selection.index {
		case 0:
			if err := getTotalCountInDB(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to get total count in db: %w", err)
			}
		case 1:
			if err := getAverageBrewingRating(ctx, db); err != nil {
				return fmt.Errorf("buna: ui: failed to get average brewing rating: %w", err)
			}
		default:
			return errors.New("buna: ui: invalid statistics index")
		}
	case control:
		switch selection.index {
		case 0:
			// Special case
			// Already handled in Run()
		case 1:
			if err := clearTerminalScreen(); err != nil {
				return fmt.Errorf("buna: ui: failed to clear terminal screen: %w", err)
			}
		case 2:
			if err := displayOptions(); err != nil {
				return fmt.Errorf("buna: ui: failed to display main options: %w", err)
			}
		default:
			return errors.New("buna: ui: control index")
		}
	default:
		return errors.New("buna: ui: invalid category")
	}
	return nil
}

func clearTerminalScreen() error {
	clearFunc, ok := clear[runtime.GOOS]
	if ok {
		if err := clearFunc(); err != nil {
			return fmt.Errorf("buna: ui: failed to clear terminal screen: %w", err)
		}
	} else {
		return errors.New("buna: ui: unsupported operating system for clearing terminal screen")
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
