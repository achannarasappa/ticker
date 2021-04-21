package util

import (
	"math"
	"strconv"

	c "github.com/achannarasappa/ticker/internal/common"
)

func getPrecision(f float64) int {
	v := math.Abs(f)

	if v == 0.0 {
		return 1
	}

	if v >= 10000 {
		return 0
	}

	if v >= 1000 {
		return 1
	}

	if v >= 100 {
		return 2
	}

	if v >= 10 {
		return 3
	}

	if v >= 1 {
		return 4
	}
	
	return 5
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

func PriceToString(f float64) string {
	prec := getPrecision(f)
	return strconv.FormatFloat(f, 'f', prec, 64)
}

func ConvertPercent(f float64) string {
	return "("+strconv.FormatFloat(f, 'f', 2, 64)+"%)"
}

func ConvertMktcap(f float64) string {
	if(f>1e12) {
		return strconv.FormatFloat(f/1e12, 'f', 2, 64)+"T"
	}

	if(f>1e9) {
		return strconv.FormatFloat(f/1e9, 'f', 2, 64)+"B"
	}

	if(f>1e6) {
		return strconv.FormatFloat(f/1e6, 'f', 2, 64)+"M"
	}

	if(f>0) {
		return strconv.FormatFloat(f, 'f', 2, 64)
	}

	return ""
}
