package util

import (
	"math"
	"strconv"

	c "github.com/achannarasappa/ticker/v4/internal/common"
)

func getPrecision(f float64) int {

	v := math.Abs(f)

	if v == 0.0 {
		return 2
	}

	if v >= 10000 {
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

// ValueText formats a float as a styled string
func ValueText(value float64, styles c.Styles) string {
	if value <= 0.0 {
		return ""
	}

	return styles.Text(ConvertFloatToString(value, false))
}
