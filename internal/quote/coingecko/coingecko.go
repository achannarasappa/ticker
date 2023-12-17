package coingecko

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

type ResponseQuotes []ResponseQuote

type ResponseQuote struct {
	Id                           string  `json:"id"` //nolint:golint,stylecheck,revive
	Symbol                       string  `json:"symbol"`
	Name                         string  `json:"name"`
	Image                        string  `json:"image"`
	CurrentPrice                 float64 `json:"current_price"`
	MarketCap                    float64 `json:"market_cap"`
	MarketCapRank                float64 `json:"market_cap_rank"`
	FullyDilutedValuation        float64 `json:"fully_diluted_valuation"`
	TotalVolume                  float64 `json:"total_volume"`
	High24h                      float64 `json:"high_24h"`
	Low24h                       float64 `json:"low_24h"`
	PriceChange24h               float64 `json:"price_change_24h"`
	PriceChangePercentage24h     float64 `json:"price_change_percentage_24h"`
	MarketCapChange24h           float64 `json:"market_cap_change_24h"`
	MarketCapChangePercentage24h float64 `json:"market_cap_change_percentage_24h"`
	CirculatingSupply            float64 `json:"circulating_supply"`
	TotalSupply                  float64 `json:"total_supply"`
	MaxSupply                    float64 `json:"max_supply"`
	Ath                          float64 `json:"ath"`
	AthChangePercentage          float64 `json:"ath_change_percentage"`
	AthDate                      string  `json:"ath_date"`
	Atl                          float64 `json:"atl"`
	AtlChangePercentage          float64 `json:"atl_change_percentage"`
	AtlDate                      string  `json:"atl_date"`
	LastUpdated                  string  `json:"last_updated"`
}

func transformResponseToAssetQuotes(responseQuotes *ResponseQuotes) []c.AssetQuote {

	assetQuotes := make([]c.AssetQuote, 0)

	for _, responseQuote := range *responseQuotes {

		assetQuote := c.AssetQuote{
			Name:   responseQuote.Name,
			Symbol: strings.ToUpper(responseQuote.Symbol),
			Class:  c.AssetClassCryptocurrency,
			Currency: c.Currency{
				FromCurrencyCode: "USD",
			},
			QuotePrice: c.QuotePrice{
				Price:          responseQuote.CurrentPrice,
				PricePrevClose: responseQuote.CurrentPrice - responseQuote.PriceChange24h,
				PriceOpen:      0.0,
				PriceDayHigh:   responseQuote.High24h,
				PriceDayLow:    responseQuote.Low24h,
				Change:         responseQuote.PriceChange24h,
				ChangePercent:  responseQuote.PriceChangePercentage24h,
			},
			QuoteExtended: c.QuoteExtended{
				FiftyTwoWeekHigh: responseQuote.Ath,
				FiftyTwoWeekLow:  responseQuote.Atl,
				MarketCap:        responseQuote.MarketCap,
				Volume:           responseQuote.TotalVolume,
			},
			QuoteSource: c.QuoteSourceCoingecko,
			Exchange: c.Exchange{
				Name:                    "Crypto Aggregate",
				Delay:                   0,
				State:                   c.ExchangeStateOpen,
				IsActive:                true,
				IsRegularTradingSession: true,
			},
			Meta: c.Meta{
				IsVariablePrecision: true,
			},
		}

		assetQuotes = append(assetQuotes, assetQuote)

	}

	return assetQuotes

}

func GetAssetQuotes(client resty.Client, symbols []string) []c.AssetQuote {
	symbolsString := strings.Join(symbols, ",")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=250&page=1&sparkline=false", symbolsString)
	res, _ := client.R().
		SetResult(ResponseQuotes{}).
		Get(url)

	out := (res.Result().(*ResponseQuotes)) //nolint:forcetypeassert

	assetQuotes := transformResponseToAssetQuotes(out)

	return assetQuotes
}
