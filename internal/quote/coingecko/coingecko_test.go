package coingecko_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/quote/coingecko"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Coingecko", func() {

	Describe("GetAssetQuotes", func() {

		It("should make a request to get crypto quotes and transform the response", func() {
			responseFixture := `[
				{
						"ath": 69045,
						"ath_change_percentage": -43.4461,
						"ath_date": "2021-11-10T14:24:11.849Z",
						"atl": 67.81,
						"atl_change_percentage": 57484.55501,
						"atl_date": "2013-07-06T00:00:00.000Z",
						"circulating_supply": 18964093.0,
						"current_price": 39045,
						"fully_diluted_valuation": 819997729028,
						"high_24h": 40090,
						"id": "bitcoin",
						"image": "https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1547033579",
						"last_updated": "2022-02-21T01:24:23.221Z",
						"low_24h": 38195,
						"market_cap": 740500628242,
						"market_cap_change_24h": -17241577635.956177,
						"market_cap_change_percentage_24h": -2.27539,
						"market_cap_rank": 1,
						"max_supply": 21000000.0,
						"name": "Bitcoin",
						"price_change_24h": -978.048909591315,
						"price_change_percentage_24h": -2.44373,
						"roi": null,
						"symbol": "btc",
						"total_supply": 21000000.0,
						"total_volume": 16659222262
				}
		]`
			responseUrl := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=bitcoin&order=market_cap_desc&per_page=250&page=1&sparkline=false"
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"bitcoin"})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(39045.0),
						"PricePrevClose": Equal(40023.048909591314),
						"PriceOpen":      Equal(0.0),
						"PriceDayHigh":   Equal(40090.0),
						"PriceDayLow":    Equal(38195.0),
						"Change":         Equal(-978.048909591315),
						"ChangePercent":  Equal(-2.44373),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoingecko),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
				}),
			}))
		})

	})

})
