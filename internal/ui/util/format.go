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

func ValueChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return ""
	}

	return QuoteChangeText(change, changePercent)
}

func QuoteChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return StyleNeutralFaded("  " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return StylePricePositive(changePercent)("↑ " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	return StylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change) + " (" + ConvertFloatToString(changePercent) + "%)")
}
