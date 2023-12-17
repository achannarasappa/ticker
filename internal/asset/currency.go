package asset

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/currency"
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
