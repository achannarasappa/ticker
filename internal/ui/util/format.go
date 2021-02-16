package util

import "strconv"

func ConvertFloatToString(f float64) string {
	var prec = 2
	if f < 10 && f > -10 {
		prec = 4
	} else if f < 100 && f > -100 {
		prec = 3
	}

	return strconv.FormatFloat(f, 'f', prec, 64)
}

func ValueText(value float64) string {
	if value <= 0.0 {
		return ""
	}

	return StyleNeutral(ConvertFloatToString(value))
}
