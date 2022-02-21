package coingecko

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/go-resty/resty/v2"
)

type ResponseQuotes []ResponseQuote

type ResponseQuote struct {
	Id                           string  `json:"id"`                               //"bitcoin",
	Symbol                       string  `json:"symbol"`                           //"btc",
	Name                         string  `json:"name"`                             //"Bitcoin",
	Image                        string  `json:"image"`                            //"https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1547033579",
	CurrentPrice                 float64 `json:"current_price"`                    //42306,
	MarketCap                    float64 `json:"market_cap"`                       //802234977601,
	MarketCapRank                float64 `json:"market_cap_rank"`                  //1,
	FullyDilutedValuation        float64 `json:"fully_diluted_valuation"`          //888680014113,
	TotalVolume                  float64 `json:"total_volume"`                     //12256877478,
	High24h                      float64 `json:"high_24h"`                         //42751,
	Low24h                       float64 `json:"low_24h"`                          //41992,
	PriceChange24h               float64 `json:"price_change_24h"`                 //307.47,
	PriceChangePercentage24h     float64 `json:"price_change_percentage_24h"`      //0.73209,
	MarketCapChange24h           float64 `json:"market_cap_change_24h"`            //6782357519,
	MarketCapChangePercentage24h float64 `json:"market_cap_change_percentage_24h"` //0.85264,
	CirculatingSupply            float64 `json:"circulating_supply"`               //18957256,
	TotalSupply                  float64 `json:"total_supply"`                     //21000000,
	MaxSupply                    float64 `json:"max_supply"`                       //21000000,
	Ath                          float64 `json:"ath"`                              //69045,
	AthChangePercentage          float64 `json:"ath_change_percentage"`            //-38.70919,
	AthDate                      string  `json:"ath_date"`                         //"2021-11-10T14:24:11.849Z",
	Atl                          float64 `json:"atl"`                              //67.81,
	AtlChangePercentage          float64 `json:"atl_change_percentage"`            //62307.78644,
	AtlDate                      string  `json:"atl_date"`                         //"2013-07-06T00:00:00.000Z",
	LastUpdated                  string  `json:"last_updated"`                     //"2022-02-13T22:05:58.681Z"
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
				PricePrevClose: responseQuote.CurrentPrice + responseQuote.PriceChange24h,
				PriceOpen:      responseQuote.CurrentPrice + responseQuote.PriceChange24h,
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

	out := (res.Result().(*ResponseQuotes))

	assetQuotes := transformResponseToAssetQuotes(out)

	return assetQuotes
}
