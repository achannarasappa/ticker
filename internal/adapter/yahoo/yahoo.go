package yahoo

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/go-resty/resty/v2"
)

// ResponseQuote represents a quote of a single security from the API response
type ResponseQuote struct {
	ShortName                  string  `json:"shortName"`
	Symbol                     string  `json:"symbol"`
	MarketState                string  `json:"marketState"`
	Currency                   string  `json:"currency"`
	ExchangeName               string  `json:"fullExchangeName"`
	ExchangeDelay              float64 `json:"exchangeDataDelayedBy"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayRange      string  `json:"regularMarketDayRange"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketVolume        float64 `json:"regularMarketVolume"`
	PostMarketChange           float64 `json:"postMarketChange"`
	PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
	PreMarketChange            float64 `json:"preMarketChange"`
	PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
	PreMarketPrice             float64 `json:"preMarketPrice"`
	FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
	QuoteType                  string  `json:"quoteType"`
	MarketCap                  float64 `json:"marketCap"`
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
			Code: responseQuote.Currency,
		},
		QuotePrice: c.QuotePrice{
			Price:          responseQuote.RegularMarketPrice,
			PricePrevClose: responseQuote.RegularMarketPreviousClose,
			PriceOpen:      responseQuote.RegularMarketOpen,
			PriceDayHigh:   responseQuote.RegularMarketDayHigh,
			PriceDayLow:    responseQuote.RegularMarketDayLow,
			Change:         responseQuote.RegularMarketChange,
			ChangePercent:  responseQuote.RegularMarketChangePercent,
		},
		QuoteExtended: c.QuoteExtended{
			FiftyTwoWeekHigh: responseQuote.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:  responseQuote.FiftyTwoWeekLow,
			MarketCap:        responseQuote.MarketCap,
			Volume:           responseQuote.RegularMarketVolume,
		},
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

	if responseQuote.MarketState == "POST" && responseQuote.PostMarketPrice == 0.0 {
		assetQuote.Exchange.IsRegularTradingSession = false
		return assetQuote
	}

	if responseQuote.MarketState == "PRE" && responseQuote.PreMarketPrice == 0.0 {
		assetQuote.Exchange.IsActive = false
		assetQuote.Exchange.IsRegularTradingSession = false
		return assetQuote
	}

	if responseQuote.MarketState == "POST" {
		assetQuote.QuotePrice.Price = responseQuote.PostMarketPrice
		assetQuote.QuotePrice.Change = (responseQuote.PostMarketChange + responseQuote.RegularMarketChange)
		assetQuote.QuotePrice.ChangePercent = responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent
		assetQuote.Exchange.IsRegularTradingSession = false
		return assetQuote
	}

	if responseQuote.MarketState == "PRE" {
		assetQuote.QuotePrice.Price = responseQuote.PreMarketPrice
		assetQuote.QuotePrice.Change = responseQuote.PreMarketChange
		assetQuote.QuotePrice.ChangePercent = responseQuote.PreMarketChangePercent
		assetQuote.Exchange.IsRegularTradingSession = false
		return assetQuote
	}

	if responseQuote.PostMarketPrice != 0.0 {
		assetQuote.QuotePrice.Price = responseQuote.PostMarketPrice
		assetQuote.QuotePrice.Change = (responseQuote.PostMarketChange + responseQuote.RegularMarketChange)
		assetQuote.QuotePrice.ChangePercent = responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent
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

// GetQuotes issues a HTTP request to retrieve quotes from the API and process the response
func GetAssetQuotes(client resty.Client, symbols []string) []c.AssetQuote {
	symbolsString := strings.Join(symbols, ",")
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
	res, _ := client.R().
		SetResult(Response{}).
		Get(url)

	return transformResponseQuotes((res.Result().(*Response)).QuoteResponse.Quotes)
}
