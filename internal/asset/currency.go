package asset

import (
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
)

func convertAssetQuotePriceCurrency(currencyRateByUse currency.CurrencyRateByUse, quotePrice c.QuotePrice) c.QuotePrice {
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

func convertAssetQuoteExtendedCurrency(currencyRateByUse currency.CurrencyRateByUse, quoteExtended c.QuoteExtended) c.QuoteExtended {
	return c.QuoteExtended{
		FiftyTwoWeekHigh: quoteExtended.FiftyTwoWeekHigh * currencyRateByUse.QuotePrice,
		FiftyTwoWeekLow:  quoteExtended.FiftyTwoWeekLow * currencyRateByUse.QuotePrice,
		MarketCap:        quoteExtended.MarketCap * currencyRateByUse.QuotePrice,
		Volume:           quoteExtended.Volume,
	}
}

func convertAssetHoldingCurrency(currencyRateByUse currency.CurrencyRateByUse, holding c.Holding) c.Holding {
	return c.Holding{
		Value:     holding.Value * currencyRateByUse.QuotePrice,
		Cost:      holding.Cost * currencyRateByUse.PositionCost,
		Quantity:  holding.Quantity,
		UnitValue: holding.UnitValue * currencyRateByUse.QuotePrice,
		UnitCost:  holding.UnitCost * currencyRateByUse.PositionCost,
		DayChange: c.HoldingChange{
			Amount:  holding.DayChange.Amount * currencyRateByUse.QuotePrice,
			Percent: holding.DayChange.Percent,
		},
		TotalChange: c.HoldingChange{
			Amount:  holding.TotalChange.Amount * currencyRateByUse.QuotePrice,
			Percent: holding.TotalChange.Percent,
		},
		Weight: holding.Weight,
	}
}
