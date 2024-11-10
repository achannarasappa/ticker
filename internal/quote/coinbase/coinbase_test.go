package coinbase_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/quote/coinbase"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Coinbase Quote", func() {
	Describe("GetAssetQuotes", func() {
		It("should make a request to get crypto quotes and transform the response", func() {
			responseFixture := `{
				"products": [
					{
						"base_display_symbol": "BTC-USD",
						"base_name": "Bitcoin",
						"price": "50000.00",
						"price_percentage_change_24h": "2.0408163265306123",
						"volume_24h": "1500.50",
						"high_24h": "51000.00",
						"low_24h": "49000.00",
						"open_24h": "49000.00",
						"display_name": "Bitcoin",
						"status": "online",
						"quote_currency_id": "USD",
						"product_venue": "Coinbase"
					}
				]
			}`
			responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"BTC-USD"})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"Name":   Equal("Bitcoin"),
					"Symbol": Equal("BTC-USD"),
					"Class":  Equal(c.AssetClassCryptocurrency),
					"Currency": g.MatchFields(g.IgnoreExtras, g.Fields{
						"FromCurrencyCode": Equal("USD"),
					}),
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(50000.00),
						"PricePrevClose": Equal(49000.00),
						"PriceOpen":      Equal(49000.00),
						"PriceDayHigh":   Equal(51000.00),
						"PriceDayLow":    Equal(49000.00),
						"Change":         Equal(1020.4081632653063),
						"ChangePercent":  Equal(2.0408163265306123),
					}),
					"QuoteExtended": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Volume": Equal(1500.50),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoinbase),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Name":                    Equal("Coinbase"),
						"State":                   Equal(c.ExchangeStateOpen),
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
					"Meta": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsVariablePrecision": BeTrue(),
					}),
				}),
			}))
		})

		When("the crypto asset is offline", func() {
			It("should mark the asset as inactive", func() {
				responseFixture := `{
				"products": [
					{
						"base_display_symbol": "BTC-USD",
						"base_name": "Bitcoin",
						"price": "50000.00",
						"price_percentage_change_24h": "2.0408163265306123",
						"volume_24h": "1500.50",
						"high_24h": "51000.00", 
						"low_24h": "49000.00",
						"open_24h": "49000.00",
						"display_name": "Bitcoin",
						"status": "offline",
						"quote_currency_id": "USD",
						"product_venue": "Coinbase"
					}
				]
			}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"BTC-USD"})
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
							"IsActive": BeFalse(),
						}),
					}),
				}))
			})
		})

		When("the request fails", func() {

			It("should return an empty slice", func() {

				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl,
					httpmock.NewStringResponder(500, `{"error": "Internal Server Error"}`))

				output := GetAssetQuotes(*client, []string{"BTC-USD"})
				Expect(output).To(BeEmpty())
			})
		})

	})
})
