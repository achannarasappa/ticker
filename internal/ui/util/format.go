package util

import (
	"math"
	"strconv"

	c "github.com/achannarasappa/ticker/internal/common"
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

	return 2
}

func ConvertFloatToString(f float64, isVariablePrecision bool) string {

	if !isVariablePrecision {
		return strconv.FormatFloat(f, 'f', 2, 64)
	}

	prec := getPrecision(f)

	return strconv.FormatFloat(f, 'f', prec, 64)
}

func ValueText(value float64, styles c.Styles) string {
	if value <= 0.0 {
		return ""
	}

	return styles.Text(ConvertFloatToString(value, false))
}
