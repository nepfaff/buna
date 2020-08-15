package buna

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Returns a 'true' boolean if quit
func validateStrInput(quitStr string, isOptional bool) (string, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input == quitStr {
		return "", true
	}

	if !isOptional && input == "" {
		fmt.Print("A value is required. Please try again: ")
		return validateStrInput(quitStr, isOptional)
	}

	return input, false
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
	return validateYearMonthInput(quitStr, isOptional, true)
}

// Considers a month to be an integer value x such that 1 <= x <= 12.
// Returns a 'true' boolean if quit.
func validateMonthInput(quitStr string, isOptional bool) (int, bool) {
	return validateYearMonthInput(quitStr, isOptional, false)
}

// Helper function for validateYearInput() and validateMonthInput().
// isYear must be true for validateYearInput() and false for validateMonthInput()
func validateYearMonthInput(quitStr string, isOptional bool, isYear bool) (int, bool) {
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
		return validateYearMonthInput(quitStr, isOptional, isYear)
	}

	var min, max int
	if isYear {
		min = 2020
		max = time.Now().Year()
	} else {
		min = 1
		max = 12
	}

	num, err := strconv.Atoi(input)
	if err != nil || num < min || num > max {
		fmt.Print("Input invalid. Please try again: ")
		return validateYearMonthInput(quitStr, isOptional, isYear)
	}

	return num, false
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

// Creates a date sring in the format "YYYY-MM-DD".
// Expects the inputs to be valid
func createDateString(year int, month int, day int) string {
	yearStr := strconv.Itoa(year)

	monthStr := strconv.Itoa(month)
	if len(monthStr) < 2 {
		monthStr = "0" + monthStr
	}

	dayStr := strconv.Itoa(day)
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
