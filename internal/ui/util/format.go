package util

import "strconv"

func ConvertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
