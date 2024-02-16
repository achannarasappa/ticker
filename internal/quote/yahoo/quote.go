package yahoo

import (
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

//nolint:gochecknoglobals
var (
	postMarketStatuses = map[string]bool{"POST": true, "POSTPOST": true}
)

// ResponseQuote represents a quote of a single security from the API response
type ResponseQuote struct {
	ShortName                  string              `json:"shortName"`
	Symbol                     string              `json:"symbol"`
	MarketState                string              `json:"marketState"`
	Currency                   string              `json:"currency"`
	ExchangeName               string              `json:"fullExchangeName"`
	ExchangeDelay              float64             `json:"exchangeDataDelayedBy"`
	RegularMarketChange        ResponseFieldFloat  `json:"regularMarketChange"`
	RegularMarketChangePercent ResponseFieldFloat  `json:"regularMarketChangePercent"`
	RegularMarketPrice         ResponseFieldFloat  `json:"regularMarketPrice"`
	RegularMarketPreviousClose ResponseFieldFloat  `json:"regularMarketPreviousClose"`
	RegularMarketOpen          ResponseFieldFloat  `json:"regularMarketOpen"`
	RegularMarketDayRange      ResponseFieldString `json:"regularMarketDayRange"`
	RegularMarketDayHigh       ResponseFieldFloat  `json:"regularMarketDayHigh"`
	RegularMarketDayLow        ResponseFieldFloat  `json:"regularMarketDayLow"`
	RegularMarketVolume        ResponseFieldFloat  `json:"regularMarketVolume"`
	PostMarketChange           ResponseFieldFloat  `json:"postMarketChange"`
	PostMarketChangePercent    ResponseFieldFloat  `json:"postMarketChangePercent"`
	PostMarketPrice            ResponseFieldFloat  `json:"postMarketPrice"`
	PreMarketChange            ResponseFieldFloat  `json:"preMarketChange"`
	PreMarketChangePercent     ResponseFieldFloat  `json:"preMarketChangePercent"`
	PreMarketPrice             ResponseFieldFloat  `json:"preMarketPrice"`
	FiftyTwoWeekHigh           ResponseFieldFloat  `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            ResponseFieldFloat  `json:"fiftyTwoWeekLow"`
	QuoteType                  string              `json:"quoteType"`
	MarketCap                  ResponseFieldFloat  `json:"marketCap"`
}

type ResponseFieldFloat struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

type ResponseFieldString struct {
	Raw string `json:"raw"`
	Fmt string `json:"fmt"`
}

func getAssetClass(assetClass string) c.AssetClass {

	if assetClass == "CRYPTOCURRENCY" {
		return c.AssetClassCryptocurrency
	}

	return c.AssetClassStock

}

// Response represents the container object from the API response
type Response struct {
	QuoteResponse struct {
		Quotes []ResponseQuote `json:"result"`
		Error  interface{}     `json:"error"`
	} `json:"quoteResponse"`
}

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

func transformResponseQuotes(responseQuotes []ResponseQuote) []c.AssetQuote {

	quotes := make([]c.AssetQuote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformResponseQuote(responseQuote))
	}

	return quotes

}

// GetAssetQuotes issues a HTTP request to retrieve quotes from the API and process the response
func GetAssetQuotes(client resty.Client, symbols []string) func() []c.AssetQuote {
	return func() []c.AssetQuote {
		symbolsString := strings.Join(symbols, ",")

		res, _ := client.R().
			SetResult(Response{}).
			SetQueryParam("fields", "shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap").
			SetQueryParam("symbols", symbolsString).
			Get("/v7/finance/quote")

		return transformResponseQuotes((res.Result().(*Response)).QuoteResponse.Quotes) //nolint:forcetypeassert
	}
}
