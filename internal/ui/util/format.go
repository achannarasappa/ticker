package util

import "strconv"

func ConvertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func ValueText(value float64) string {
	if value <= 0.0 {
		return ""
	}

	return StyleNeutral(ConvertFloatToString(value))
}
