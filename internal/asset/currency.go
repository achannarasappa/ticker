package asset

import (
	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// currencyRateByUse represents the currency conversion rate for each use case
type currencyRateByUse struct { //nolint:golint,revive
	ToCurrencyCode string
	QuotePrice     float64
	PositionCost   float64
	SummaryValue   float64
	SummaryCost    float64
}

// getCurrencyRateByUse reads currency rates from the context and sets the conversion rate for each use case
func getCurrencyRateByUse(ctx c.Context, fromCurrency string, toCurrency string, rate float64) currencyRateByUse {

	if rate == 0 {
		return currencyRateByUse{
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
		return currencyRateByUse{
			ToCurrencyCode: fromCurrency,
			QuotePrice:     1.0,
			PositionCost:   1.0,
			SummaryValue:   rate,
			SummaryCost:    currencyRateCost,
		}
	}

	// Convert all quotes and positions to target currency and implicitly convert summary currency (i.e. no conversion since underlying values are already converted)
	if ctx.Config.Currency != "" {
		return currencyRateByUse{
			ToCurrencyCode: toCurrency,
			QuotePrice:     rate,
			PositionCost:   currencyRateCost,
			SummaryValue:   1.0,
			SummaryCost:    1.0,
		}
	}

	// Convert only the summary currency to the default currency (USD) when currency conversion is not enabled
	return currencyRateByUse{
		ToCurrencyCode: toCurrency,
		QuotePrice:     1.0,
		PositionCost:   1.0,
		SummaryValue:   rate,
		SummaryCost:    currencyRateCost,
	}
}

func convertAssetQuotePriceCurrency(currencyRateByUse currencyRateByUse, quotePrice c.QuotePrice) c.QuotePrice {
	return c.QuotePrice{
		Price:          quotePrice.Price * currencyRateByUse.QuotePrice,
		PricePrevClose: quotePrice.PricePrevClose * currencyRateByUse.QuotePrice,
		PriceOpen:      quotePrice.PriceOpen * currencyRateByUse.QuotePrice,
		PriceDayHigh:   quotePrice.PriceDayHigh * currencyRateByUse.QuotePrice,
		PriceDayLow:    quotePrice.PriceDayLow * currencyRateByUse.QuotePrice,
		Change:         quotePrice.Change * currencyRateByUse.QuotePrice,
		ChangePercent:  quotePrice.ChangePercent,
	}
}

func convertAssetQuoteExtendedCurrency(currencyRateByUse currencyRateByUse, quoteExtended c.QuoteExtended) c.QuoteExtended {
	return c.QuoteExtended{
		FiftyTwoWeekHigh: quoteExtended.FiftyTwoWeekHigh * currencyRateByUse.QuotePrice,
		FiftyTwoWeekLow:  quoteExtended.FiftyTwoWeekLow * currencyRateByUse.QuotePrice,
		MarketCap:        quoteExtended.MarketCap * currencyRateByUse.QuotePrice,
		Volume:           quoteExtended.Volume,
	}
}
