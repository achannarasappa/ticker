package currency

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
)

// CurrencyRateByUse represents the currency conversion rate for each use case
type CurrencyRateByUse struct { //nolint:golint,revive
	ToCurrencyCode string
	QuotePrice     float64
	PositionCost   float64
	SummaryValue   float64
	SummaryCost    float64
}

// GetCurrencyRateFromContext reads currency rates from the context and sets the conversion rate for each use case
func GetCurrencyRateFromContext(ctx c.Context, fromCurrency string, toCurrency string, rate float64) CurrencyRateByUse {

	if rate == 0 {
		return CurrencyRateByUse{
			ToCurrencyCode: fromCurrency,
			QuotePrice:     1.0,
			PositionCost:   1.0,
			SummaryValue:   1.0,
			SummaryCost:    1.0,
		}
	}

	currencyRateCost := rate

	if ctx.Config.CurrencyDisableUnitCostConversion {
		currencyRateCost = 1.0
	}

	// Convert only the summary currency to the configured currency
	if ctx.Config.Currency != "" && ctx.Config.CurrencyConvertSummaryOnly {
		return CurrencyRateByUse{
			ToCurrencyCode: fromCurrency,
			QuotePrice:     1.0,
			PositionCost:   1.0,
			SummaryValue:   rate,
			SummaryCost:    currencyRateCost,
		}
	}

	// Convert all quotes and positions to target currency and implicitly convert summary currency (i.e. no conversion since underlying values are already converted)
	if ctx.Config.Currency != "" {
		return CurrencyRateByUse{
			ToCurrencyCode: toCurrency,
			QuotePrice:     rate,
			PositionCost:   currencyRateCost,
			SummaryValue:   1.0,
			SummaryCost:    1.0,
		}
	}

	// Convert only the summary currency to the default currency (USD) when currency conversion is not enabled
	return CurrencyRateByUse{
		ToCurrencyCode: toCurrency,
		QuotePrice:     1.0,
		PositionCost:   1.0,
		SummaryValue:   rate,
		SummaryCost:    currencyRateCost,
	}

}
