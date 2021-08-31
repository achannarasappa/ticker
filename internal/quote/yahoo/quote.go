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

// Quote prices for some symbols come back in a currency fraction format. For example,
// UK prices are in GBp (which represent pennies, i.e. 1/100 of GBP) and hence conversion
// fails because Yahoo returns GBPUSD and not GBpUSD.
type CurrencyFraction struct {
	CurrencyCode string
	Ratio        float64
}

func guessCurrencyFraction(currencyFractionCode string) CurrencyFraction {
	switch currencyFractionCode {
	case "GBp":
		return CurrencyFraction{
			CurrencyCode: "GBP",
			Ratio:        0.01,
		}
	default:
		return CurrencyFraction{
			CurrencyCode: currencyFractionCode,
			Ratio:        1,
		}
	}
}

func transformResponseQuote(responseQuote ResponseQuote) c.AssetQuote {

	assetClass := getAssetClass(responseQuote.QuoteType)
	isVariablePrecision := (assetClass == c.AssetClassCryptocurrency)

	currencyFraction := guessCurrencyFraction(responseQuote.Currency)

	assetQuote := c.AssetQuote{
		Name:   responseQuote.ShortName,
		Symbol: responseQuote.Symbol,
		Class:  assetClass,
		Currency: c.Currency{
			FromCurrencyCode: currencyFraction.CurrencyCode,
		},
		QuotePrice: c.QuotePrice{
			Price:          responseQuote.RegularMarketPrice * currencyFraction.Ratio,
			PricePrevClose: responseQuote.RegularMarketPreviousClose * currencyFraction.Ratio,
			PriceOpen:      responseQuote.RegularMarketOpen * currencyFraction.Ratio,
			PriceDayHigh:   responseQuote.RegularMarketDayHigh * currencyFraction.Ratio,
			PriceDayLow:    responseQuote.RegularMarketDayLow * currencyFraction.Ratio,
			Change:         responseQuote.RegularMarketChange * currencyFraction.Ratio,
			ChangePercent:  responseQuote.RegularMarketChangePercent * currencyFraction.Ratio,
		},
		QuoteExtended: c.QuoteExtended{
			FiftyTwoWeekHigh: responseQuote.FiftyTwoWeekHigh * currencyFraction.Ratio,
			FiftyTwoWeekLow:  responseQuote.FiftyTwoWeekLow * currencyFraction.Ratio,
			MarketCap:        responseQuote.MarketCap * currencyFraction.Ratio,
			Volume:           responseQuote.RegularMarketVolume,
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

// GetAssetQuotes issues a HTTP request to retrieve quotes from the API and process the response
func GetAssetQuotes(client resty.Client, symbols []string) func() []c.AssetQuote {
	return func() []c.AssetQuote {
		symbolsString := strings.Join(symbols, ",")
		url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
		res, _ := client.R().
			SetResult(Response{}).
			Get(url)

		return transformResponseQuotes((res.Result().(*Response)).QuoteResponse.Quotes)
	}
}
