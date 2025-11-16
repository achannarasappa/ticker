package util

import (
	"math"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

func getPrecision(f float64) int {

	v := math.Abs(f)

	if v == 0.0 {
		return 2
	}

	if v >= 1000000 {
		return 0
	}

	if v < 10 {
		return 4
	}

	if v < 100 {
		return 3
	}

	if v >= 1000 && f < 0 {
		return 1
	}

	return 2
}

// addCommaDelimiters adds comma delimiters to the integer part of a number string
func addCommaDelimiters(s string) string {
	// Handle negative numbers
	isNegative := strings.HasPrefix(s, "-")
	if isNegative {
		s = s[1:]
	}

	// Split on decimal point
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = "." + parts[1]
	}

	// Add commas to integer part
	n := len(integerPart)
	if n <= 3 {
		if isNegative {
			return "-" + integerPart + decimalPart
		}

		return integerPart + decimalPart
	}

	var result strings.Builder
	for i, digit := range integerPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	if isNegative {
		return "-" + result.String() + decimalPart
	}

	return result.String() + decimalPart
}

// ConvertFloatToString formats a float as a string including handling large or small numbers
func ConvertFloatToString(f float64, isVariablePrecision bool) string {

	var unit string

	if !isVariablePrecision {
		return strconv.FormatFloat(f, 'f', 2, 64)
	}

	if f > 1000000000000 {
		f /= 1000000000000
		unit = " T"
	}

	if f > 1000000000 {
		f /= 1000000000
		unit = " B"
	}

	if f > 1000000 {
		f /= 1000000
		unit = " M"
	}

	prec := getPrecision(f)

	return strconv.FormatFloat(f, 'f', prec, 64) + unit
}

// ConvertFloatToStringWithCommas formats a float as a string with comma delimiters including handling large or small numbers
func ConvertFloatToStringWithCommas(f float64, isVariablePrecision bool) string {

	var unit string

	if !isVariablePrecision {
		formatted := strconv.FormatFloat(f, 'f', 2, 64)

		return addCommaDelimiters(formatted)
	}

	if f > 1000000000000 {
		f /= 1000000000000
		unit = " T"
	}

	if f > 1000000000 {
		f /= 1000000000
		unit = " B"
	}

	if f > 1000000 {
		f /= 1000000
		unit = " M"
	}

	prec := getPrecision(f)

	formatted := strconv.FormatFloat(f, 'f', prec, 64)

	return addCommaDelimiters(formatted) + unit
}

// ValueText formats a float as a styled string
func ValueText(value float64, styles c.Styles) string {
	if value <= 0.0 {
		return ""
	}

	return styles.Text(ConvertFloatToStringWithCommas(value, false))
}
