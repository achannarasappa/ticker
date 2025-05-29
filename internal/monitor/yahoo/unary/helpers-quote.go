package unary

import (
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

//nolint:gochecknoglobals
var (
	postMarketStatuses = map[string]bool{"POST": true, "POSTPOST": true}
)

// transformResponseQuote transforms a single quote returned by the API into an AssetQuote
func transformResponseQuote(responseQuote ResponseQuote) c.AssetQuote {

	assetClass := getAssetClass(responseQuote.QuoteType)
	isVariablePrecision := (assetClass == c.AssetClassCryptocurrency)

	assetQuote := c.AssetQuote{
		Name:   responseQuote.ShortName,
		Symbol: responseQuote.Symbol,
		Class:  assetClass,
		Currency: c.Currency{
			FromCurrencyCode: strings.ToUpper(responseQuote.Currency),
		},
		QuotePrice: c.QuotePrice{
			Price:          responseQuote.RegularMarketPrice.Raw,
			PricePrevClose: responseQuote.RegularMarketPreviousClose.Raw,
			PriceOpen:      responseQuote.RegularMarketOpen.Raw,
			PriceDayHigh:   responseQuote.RegularMarketDayHigh.Raw,
			PriceDayLow:    responseQuote.RegularMarketDayLow.Raw,
			Change:         responseQuote.RegularMarketChange.Raw,
			ChangePercent:  responseQuote.RegularMarketChangePercent.Raw,
		},
		QuoteExtended: c.QuoteExtended{
			FiftyTwoWeekHigh: responseQuote.FiftyTwoWeekHigh.Raw,
			FiftyTwoWeekLow:  responseQuote.FiftyTwoWeekLow.Raw,
			MarketCap:        responseQuote.MarketCap.Raw,
			Volume:           responseQuote.RegularMarketVolume.Raw,
		},
		QuoteSource: c.QuoteSourceYahoo,
		Exchange: c.Exchange{
			Name:                    responseQuote.ExchangeName,
			Delay:                   responseQuote.ExchangeDelay,
			State:                   c.ExchangeStateOpen,
			IsActive:                true,
			IsRegularTradingSession: true,
		},
		Meta: c.Meta{
			IsVariablePrecision: isVariablePrecision,
			SymbolInSourceAPI:   responseQuote.Symbol,
		},
	}

	if responseQuote.MarketState == "REGULAR" {
		return assetQuote
	}

	if _, exists := postMarketStatuses[responseQuote.MarketState]; exists && responseQuote.PostMarketPrice.Raw == 0.0 {
		assetQuote.Exchange.IsRegularTradingSession = false

		return assetQuote
	}

	if responseQuote.MarketState == "PRE" && responseQuote.PreMarketPrice.Raw == 0.0 {
		assetQuote.Exchange.IsActive = false
		assetQuote.Exchange.IsRegularTradingSession = false

		return assetQuote
	}

	if _, exists := postMarketStatuses[responseQuote.MarketState]; exists {
		assetQuote.QuotePrice.Price = responseQuote.PostMarketPrice.Raw
		assetQuote.QuotePrice.Change = (responseQuote.PostMarketChange.Raw + responseQuote.RegularMarketChange.Raw)
		assetQuote.QuotePrice.ChangePercent = responseQuote.PostMarketChangePercent.Raw + responseQuote.RegularMarketChangePercent.Raw
		assetQuote.Exchange.IsRegularTradingSession = false

		return assetQuote
	}

	if responseQuote.MarketState == "PRE" {
		assetQuote.QuotePrice.Price = responseQuote.PreMarketPrice.Raw
		assetQuote.QuotePrice.Change = responseQuote.PreMarketChange.Raw
		assetQuote.QuotePrice.ChangePercent = responseQuote.PreMarketChangePercent.Raw
		assetQuote.Exchange.IsRegularTradingSession = false

		return assetQuote
	}

	if responseQuote.PostMarketPrice.Raw != 0.0 {
		assetQuote.QuotePrice.Price = responseQuote.PostMarketPrice.Raw
		assetQuote.QuotePrice.Change = (responseQuote.PostMarketChange.Raw + responseQuote.RegularMarketChange.Raw)
		assetQuote.QuotePrice.ChangePercent = responseQuote.PostMarketChangePercent.Raw + responseQuote.RegularMarketChangePercent.Raw
		assetQuote.Exchange.IsActive = false
		assetQuote.Exchange.IsRegularTradingSession = false

		return assetQuote
	}

	assetQuote.Exchange.IsActive = false
	assetQuote.Exchange.IsRegularTradingSession = false

	return assetQuote

}

// transformResponseQuotes transforms the quotes returned by the API into a slice of AssetQuote
func transformResponseQuotes(responseQuotes []ResponseQuote) ([]c.AssetQuote, map[string]*c.AssetQuote) {
	quotes := make([]c.AssetQuote, 0, len(responseQuotes))
	quotesBySymbol := make(map[string]*c.AssetQuote, len(responseQuotes))

	for _, responseQuote := range responseQuotes {
		quote := transformResponseQuote(responseQuote)
		quotes = append(quotes, quote)
		quotesBySymbol[quote.Symbol] = &quote
	}

	return quotes, quotesBySymbol
}

// getAssetClass determines the asset class based on the quote type returned by the API
func getAssetClass(assetClass string) c.AssetClass {

	if assetClass == "CRYPTOCURRENCY" {
		return c.AssetClassCryptocurrency
	}

	return c.AssetClassStock

}
