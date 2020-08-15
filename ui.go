package buna

import (
	"context"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/table"
)

type category int

const (
	create category = iota
	retrieve
	// statistics
)

var categoryRefs = map[category]string{
	create:   "A",
	retrieve: "B",
}

var options = map[category]map[int]string{
	create: map[int]string{
		1: "New brewing",
		2: "New cupping",
		3: "New coffee purchase",
		4: "New coffee",
		5: "New brewing method",
		6: "New grinder",
	},
	retrieve: map[int]string{
		1: "Retrive brewing",
		2: "Retrieve cupping",
		3: "Retrieve coffee purchase",
		4: "Retrieve coffee",
		5: "Retrieve brewing method",
		6: "Retrieve grinder",
	},
}

func Run(ctx context.Context, db DB) error {
	displayOptions()
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
