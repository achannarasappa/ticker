package coincap

import (
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

// Quote represents a quote of a single security from the API response
type Quote struct {
	ShortName                  string `json:"name"`
	Symbol                     string `json:"symbol"`
	RegularMarketChangePercent string `json:"changePercent24Hr"`
	RegularMarketPrice         string `json:"priceUsd"`
	RegularMarketVolume        string `json:"volumeUsd24Hr"`
	MarketCap                  string `json:"marketCapUsd"`
}

// Response represents the container object from the API response
type Response struct {
	Data []Quote `json:"data"`
}

func transformQuote(responseQuote Quote) c.AssetQuote {

	price, _ := strconv.ParseFloat(responseQuote.RegularMarketPrice, 64)
	changePercent, _ := strconv.ParseFloat(responseQuote.RegularMarketChangePercent, 64)
	pricePrevClose := (1 + changePercent/100) * price
	marketCap, _ := strconv.ParseFloat(responseQuote.MarketCap, 64)
	volume, _ := strconv.ParseFloat(responseQuote.RegularMarketVolume, 64)

	assetQuote := c.AssetQuote{
		Name:   responseQuote.ShortName,
		Symbol: responseQuote.Symbol,
		Class:  c.AssetClassCryptocurrency,
		Currency: c.Currency{
			FromCurrencyCode: "USD",
		},
		QuotePrice: c.QuotePrice{
			Price:          price,
			PricePrevClose: pricePrevClose,
			PriceOpen:      0.0,
			PriceDayHigh:   0.0,
			PriceDayLow:    0.0,
			Change:         pricePrevClose - price,
			ChangePercent:  changePercent,
		},
		QuoteExtended: c.QuoteExtended{
			FiftyTwoWeekHigh: 0.0,
			FiftyTwoWeekLow:  0.0,
			MarketCap:        marketCap,
			Volume:           volume,
		},
		QuoteSource: c.QuoteSourceCoinCap,
		Exchange: c.Exchange{
			Name:                    "Crypto Aggregate via CoinCap",
			Delay:                   0,
			State:                   c.ExchangeStateOpen,
			IsActive:                true,
			IsRegularTradingSession: true,
		},
		Meta: c.Meta{
			IsVariablePrecision: true,
		},
	}

	return assetQuote

}

func transformQuotes(responseQuotes []Quote) []c.AssetQuote {

	quotes := make([]c.AssetQuote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformQuote(responseQuote))
	}

	return quotes

}

// GetAssetQuotes issues a HTTP request to retrieve quotes from the API and process the response
func GetAssetQuotes(client resty.Client, symbols []string) []c.AssetQuote {
	symbolsString := strings.Join(symbols, ",")

	res, _ := client.R().
		SetResult(Response{}).
		SetQueryParam("ids", strings.ToLower(symbolsString)).
		Get("https://api.coincap.io/v2/assets")

	return transformQuotes((res.Result().(*Response)).Data) //nolint:forcetypeassert

}
