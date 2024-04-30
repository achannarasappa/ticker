package coincap_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/quote/coincap"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("CoinCap Quote", func() {
	Describe("GetAssetQuotes", func() {
		It("should make a request to get stock quotes and transform the response", func() {
			responseFixture := `{
				"data": [
					{
						"id": "bitcoin",
						"rank": "1",
						"symbol": "BTC",
						"name": "Bitcoin",
						"supply": "19685775.0000000000000000",
						"maxSupply": "21000000.0000000000000000",
						"marketCapUsd": "1248489381324.9592799671502700",
						"volumeUsd24Hr": "7744198446.5431034815177485",
						"priceUsd": "63420.8905326287270868",
						"changePercent24Hr": "1.3622077494913284",
						"vwap24Hr": "62988.1090433238215198",
						"explorer": "https://blockchain.info/"
					}
				],
				"timestamp": 1714453771801
			}`
			responseUrl := `=~\/v2\/assets.*ids\=bitcoin.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"bitcoin"})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(63420.89053262873),
						"PricePrevClose": Equal(64284.8148182606),
						"PriceOpen":      Equal(0.0),
						"PriceDayHigh":   Equal(0.0),
						"PriceDayLow":    Equal(0.0),
						"Change":         Equal(863.9242856318742),
						"ChangePercent":  Equal(1.3622077494913285),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoinCap),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
				}),
			}))
		})
	})
})
