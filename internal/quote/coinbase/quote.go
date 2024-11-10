package coinbase

import (
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

// ResponseQuote represents a quote of a single product from the Coinbase API
type ResponseQuote struct {
	Symbol         string `json:"base_display_symbol"`
	ShortName      string `json:"base_name"`
	Price          string `json:"price"`
	PriceChange24H string `json:"price_percentage_change_24h"`
	Volume24H      string `json:"volume_24h"`
	High24H        string `json:"high_24h"`
	Low24H         string `json:"low_24h"`
	Open24H        string `json:"open_24h"`
	DisplayName    string `json:"display_name"`
	MarketState    string `json:"status"`
	Currency       string `json:"quote_currency_id"`
	ExchangeName   string `json:"product_venue"`
}

// Response represents the container object from the API response
type Response struct {
	Products []ResponseQuote `json:"products"`
}

func transformResponseQuote(responseQuote ResponseQuote) c.AssetQuote {
	price, _ := strconv.ParseFloat(responseQuote.Price, 64)
	priceOpen, _ := strconv.ParseFloat(responseQuote.Open24H, 64)
	priceDayHigh, _ := strconv.ParseFloat(responseQuote.High24H, 64)
	priceDayLow, _ := strconv.ParseFloat(responseQuote.Low24H, 64)
	volume, _ := strconv.ParseFloat(responseQuote.Volume24H, 64)
	changePercent, _ := strconv.ParseFloat(responseQuote.PriceChange24H, 64)

	// Calculate absolute price change from percentage change
	change := price * (changePercent / 100)

	return c.AssetQuote{
		Name:   responseQuote.ShortName,
		Symbol: responseQuote.Symbol,
		Class:  c.AssetClassCryptocurrency,
		Currency: c.Currency{
			FromCurrencyCode: strings.ToUpper(responseQuote.Currency),
		},
		QuotePrice: c.QuotePrice{
			Price:          price,
			PricePrevClose: priceOpen, // Using Open24H as previous close
			PriceOpen:      priceOpen,
			PriceDayHigh:   priceDayHigh,
			PriceDayLow:    priceDayLow,
			Change:         change,
			ChangePercent:  changePercent,
		},
		QuoteExtended: c.QuoteExtended{
			Volume: volume,
		},
		QuoteSource: c.QuoteSourceCoinbase,
		Exchange: c.Exchange{
			Name:                    responseQuote.ExchangeName,
			State:                   c.ExchangeStateOpen,
			IsActive:                responseQuote.MarketState == "online",
			IsRegularTradingSession: true, // Crypto markets are always in regular session
		},
		Meta: c.Meta{
			IsVariablePrecision: true,
		},
	}
}

func transformResponseQuotes(responseQuotes []ResponseQuote) []c.AssetQuote {
	quotes := make([]c.AssetQuote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformResponseQuote(responseQuote))
	}

	return quotes
}

// GetAssetQuotes issues a HTTP request to retrieve quotes from the Coinbase API
func GetAssetQuotes(client resty.Client, symbols []string) []c.AssetQuote {

	symbolsString := strings.Join(symbols, "&product_ids=")

	res, _ := client.R().
		SetResult(Response{}).
		SetQueryParam("product_ids", symbolsString).
		Get("https://api.coinbase.com/api/v3/brokerage/market/products")

	return transformResponseQuotes(res.Result().(*Response).Products) //nolint:forcetypeassert
}
