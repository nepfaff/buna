package buna

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type date struct {
	year  int
	month int
	day   int
}

// Returns a 'true' boolean if quit
// Optional strings default to "".
// Pass an empty slice for options if want to allow any string.
// Otherwise, only strings that appear in options will be accepted (+ "" if isOptional is true).
// If suggestions is empty and options is not empty, options will be used as suggestions.
func validateStrInput(quitStr string, isOptional bool, options []string, suggestions []string) (string, bool) {
	if len(suggestions) == 0 && len(options) > 0 {
		suggestions = options
	}

	suggestionNum := len(suggestions)
	if suggestionNum > 0 {
		fmt.Println("\nSelect one of the following (integer) or enter 'm' for manual entry:")
		for i, suggestion := range suggestions {
			fmt.Printf("%v. %v\n", i+1, suggestion)
		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		if input == quitStr {
			return "", true
		}

		if input == "m" {
			fmt.Println("Skipping to manual entry.")
			fmt.Print("Input: ")
		} else {
			num, err := strconv.Atoi(input)
			if err != nil || num > suggestionNum || num <= 0 {
				fmt.Println("Not a valid option. Skipping to manual entry")
				fmt.Print("Input: ")
			} else {
				return suggestions[num-1], false
			}
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input == quitStr {
		return "", true
	}

	if input == "" {
		if isOptional {
			return "", false
		}

		fmt.Print("A value is required. Please try again: ")
		return validateStrInput(quitStr, isOptional, options, nil)
	}

	if len(options) > 0 {
		for _, option := range options {
			if input == option {
				return input, false
			}
		}

		fmt.Print("Not a valid option. Please try again: ")
		return validateStrInput(quitStr, isOptional, options, nil)
	}

	return input, false
}

// Returns a 'true' boolean if quit.
// Optional integers default to 0.
// The integer bounds are specified using min and max.
func validateIntInput(quitStr string, isOptional bool, min int, max int) (int, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input == quitStr {
		return 0, true
	}

	if input == "" {
		if isOptional {
			return 0, false
		}

		fmt.Print("A value is required. Please try again: ")
		return validateIntInput(quitStr, isOptional, min, max)
	}

	num, err := strconv.Atoi(input)
	if err != nil || num < min || num > max {
		fmt.Print("Input invalid. Please try again: ")
		return validateIntInput(quitStr, isOptional, min, max)
	}

	return num, false
}

// Second return boolean is 'true' if quit.
// Optional booleans default to 'false'.
func validateBoolInput(quitStr string, isOptional bool) (bool, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input == quitStr {
		return false, true
	}

	if isOptional && input == "" {
		return false, false
	}

	inputBool, err := strconv.ParseBool(input)
	if err != nil {
		fmt.Print("Invalid value. Please try again: ")
		return validateBoolInput(quitStr, isOptional)
	}

	return inputBool, false
}

// Considers a year to be an integer value x such that 2020 <= x <= time.Year.
// Returns a 'true' boolean if quit.
func validateYearInput(quitStr string, isOptional bool) (int, bool) {
	return validateIntInput(quitStr, isOptional, 2020, time.Now().Year())
}

// Considers a month to be an integer value x such that 1 <= x <= 12.
// Returns a 'true' boolean if quit.
func validateMonthInput(quitStr string, isOptional bool) (int, bool) {
	return validateIntInput(quitStr, isOptional, 1, 12)
}

// Considers a day to be an integer value x such that 1 <= x <= (max day in month, 29 for Feb).
// Returns a 'true' boolean if quit.
func validateDayInput(quitStr string, isOptional bool, month int) (int, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input == quitStr {
		return 0, true
	}

	if input == "" {
		if isOptional {
			return 0, false
		}

		fmt.Print("A value is required. Please try again: ")
		return validateDayInput(quitStr, isOptional, month)
	}

	var max int
	switch month {
	case 1:
		max = 31
	case 2:
		max = 29
	case 3:
		max = 31
	case 4:
		max = 30
	case 5:
		max = 31
	case 6:
		max = 30
	case 7:
		max = 31
	case 8:
		max = 31
	case 9:
		max = 30
	case 10:
		max = 31
	case 11:
		max = 30
	case 12:
		max = 31
	default:
		fmt.Println("buna: input_util: invalid month passed into day validator")
		return 0, false
	}

	day, err := strconv.Atoi(input)
	if err != nil || day <= 0 || day > max {
		fmt.Print("Day invalid. Please try again: ")
		return validateDayInput(quitStr, isOptional, month)
	}

	return day, false
}

// Used to get a date input by promting the user for year, month and day separately.
// The user is not prompted to enter month and day if date is optional and no value is entered for year
// (same for day if no value is entered for month).
// inputMsg is used as the message for the user. All '?' characters are replaced by year, month or day.
// inputMsg must contain at least one '?' and should end with ": ", for it to make sense to the user.
// Returns a 'true' boolean if quit.
func getDateInput(quitStr string, isOptional bool, inputMsg string, suggestions []date) (date, bool) {
	suggestionNum := len(suggestions)
	if suggestionNum > 0 {
		dateMsg := strings.ReplaceAll(inputMsg, "?", "date")
		fmt.Println(dateMsg)
		fmt.Println("Select one of the following (integer) or enter 'm' for manual entry:")
		for i, suggestion := range suggestions {
			fmt.Printf("%v. %v\n", i+1, createDateString(suggestion))
		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		if input == quitStr {
			return date{}, true
		}

		if input == "m" {
			fmt.Println("Skipping to manual entry")
		} else {
			num, err := strconv.Atoi(input)
			if err != nil || num > suggestionNum || num <= 0 {
				fmt.Println("Not a valid option. Skipping to manual entry")
			} else {
				return suggestions[num-1], false
			}
		}
	}

	yearMsg := strings.ReplaceAll(inputMsg, "?", "year")
	fmt.Print(yearMsg)
	year, quit := validateYearInput(quitStr, isOptional)
	if quit {
		return date{}, true
	}

	if year == 0 {
		return date{}, false
	}

	monthMsg := strings.ReplaceAll(inputMsg, "?", "month")
	fmt.Print(monthMsg)
	month, quit := validateMonthInput(quitStr, isOptional)
	if quit {
		return date{}, true
	}

	if month == 0 {
		return date{}, false
	}

	dayMsg := strings.ReplaceAll(inputMsg, "?", "day")
	fmt.Print(dayMsg)
	day, quit := validateDayInput(quitStr, isOptional, month)
	if quit {
		return date{}, true
	}

	return date{
		year:  year,
		month: month,
		day:   day,
	}, false
}

// Creates a date sring in the format "YYYY-MM-DD".
// Expects the inputs to be valid.
func createDateString(date date) string {
	yearStr := strconv.Itoa(date.year)

	monthStr := strconv.Itoa(date.month)
	if len(monthStr) < 2 {
		monthStr = "0" + monthStr
	}

	dayStr := strconv.Itoa(date.day)
	if len(dayStr) < 2 {
		dayStr = "0" + dayStr
	}

	var sb strings.Builder
	sb.WriteString(yearStr)
	sb.WriteString("-")
	sb.WriteString(monthStr)
	sb.WriteString("-")
	sb.WriteString(dayStr)

	return sb.String()
}
