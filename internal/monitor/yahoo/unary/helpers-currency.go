package unary

// minorCurrency represents the scaling ('minor unit') for a minor currency.
// Convert major to minor by multiplying by 10^n and minor to major by dividing by 10^n.
type minorCurrency struct {
	MajorCurrencyCode string
	MinorCurrencyCode string
	MinorUnit         float64
}

// MinorUnitForCurrencyCode reports whether the given major (ISO 4217) currency code has a
// tradable minor form (e.g. GBP -> GBp). When it does, the minor currency code and its minor
// unit (the power of ten separating the major and minor units) are returned.
//
// Only currencies traded on major exchanges are considered. The complete ISO 4217 list is
// retained as a reference in helpers-currency.go below this function.
//
//nolint:gochecknoglobals
func MinorUnitForCurrencyCode(majorCurrency string) (bool, string, float64) {

	var minorCurrencyCodeByMajorCurrencyCode = map[string]minorCurrency{
		"AUD": {"AUD", "AUd", 2},
		"CAD": {"CAD", "CAd", 2},
		"CHF": {"CHF", "CHf", 2},
		"CNY": {"CNY", "CNy", 2},
		"EUR": {"EUR", "EUr", 2},
		"GBP": {"GBP", "GBp", 2},
		"HKD": {"HKD", "HKd", 2},
		"INR": {"INR", "INr", 2},
		"TWD": {"TWD", "TWd", 2},
		"USD": {"USD", "USd", 2},
		"ZAR": {"ZAR", "ZAr", 2},
	}

	if mc, ok := minorCurrencyCodeByMajorCurrencyCode[majorCurrency]; ok {
		return true, mc.MinorCurrencyCode, mc.MinorUnit
	}

	return false, "", 0

}
