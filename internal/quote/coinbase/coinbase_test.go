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
						"base_display_symbol": "BTC",
						"product_type": "SPOT",
						"product_id": "BTC-USD",
						"base_name": "Bitcoin",
						"price": "50000.00",
						"price_percentage_change_24h": "2.0408163265306123",
						"volume_24h": "1500.50",
						"display_name": "Bitcoin",
						"status": "online",
						"quote_currency_id": "USD",
						"product_venue": "CBE"
					}
				]
			}`
			responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"BTC-USD"}, []string{})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"Name":   Equal("Bitcoin"),
					"Symbol": Equal("BTC"),
					"Class":  Equal(c.AssetClassCryptocurrency),
					"Currency": g.MatchFields(g.IgnoreExtras, g.Fields{
						"FromCurrencyCode": Equal("USD"),
					}),
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(50000.00),
						"PricePrevClose": Equal(0.00),
						"PriceOpen":      Equal(0.00),
						"PriceDayHigh":   Equal(0.00),
						"PriceDayLow":    Equal(0.00),
						"Change":         Equal(1020.4081632653063),
						"ChangePercent":  Equal(2.0408163265306123),
					}),
					"QuoteExtended": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Volume": Equal(1500.50),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoinbase),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Name":                    Equal("CBE"),
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

		When("the market for a crypto asset is closed", func() {
			It("should mark the asset as inactive", func() {
				responseFixture := `{
				"products": [
					{
						"base_display_symbol": "BTC",
						"product_type": "SPOT",
						"product_id": "BTC-USD",
						"base_name": "Bitcoin",
						"price": "50000.00",
						"price_percentage_change_24h": "2.0408163265306123",
						"volume_24h": "1500.50",
						"display_name": "Bitcoin",
						"status": "",
						"quote_currency_id": "USD",
						"product_venue": "CBE"
					}
				]
			}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"BTC-USD"}, []string{})
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

				output := GetAssetQuotes(*client, []string{"BTC-USD"}, []string{})
				Expect(output).To(BeEmpty())
			})
		})

		When("the asset is a futures contract", func() {
			It("should return the futures contract quote", func() {

				responseFixture := `{
					"products": [
						{
							"product_id": "BIT-31JAN25-CDE",
							"price": "97345",
							"price_percentage_change_24h": "-3.14412218297597",
							"volume_24h": "93744",
							"base_name": "",
							"status": "",
							"product_type": "FUTURE",
							"quote_currency_id": "USD",
							"fcm_trading_session_details": {
								"is_session_open": true
							},
							"base_display_symbol": "",
							"product_venue": "FCM",
							"future_product_details": {
								"venue": "cde",
								"contract_code": "BIT",
								"contract_expiry": "2025-01-31T16:00:00Z",
								"contract_root_unit": "BTC",
								"group_description": "Nano Bitcoin Futures",
								"contract_expiry_timezone": "Europe/London"
							}
						}
					]
				}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BIT\-31JAN25\-CDE.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"BIT-31JAN25-CDE"}, []string{})
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Class":  Equal(c.AssetClassFuturesContract),
						"Symbol": Equal("BIT-31JAN25-CDE"),
						"Name":   Equal("Nano Bitcoin Futures"),
						"QuoteFutures": g.MatchFields(g.IgnoreExtras, g.Fields{
							"SymbolUnderlying": Equal("BTC"),
							"IndexPrice":       Equal(0.00),
							"Basis":            Equal(0.00),
						}),
						"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Name":                    Equal("FCM"),
							"State":                   Equal(c.ExchangeStateOpen),
							"IsActive":                BeTrue(),
							"IsRegularTradingSession": BeTrue(),
						}),
					}),
				}))
			})

			When("the futures contract's underlying asset has a quote", func() {
				It("should return the properties based off of the underlying asset's quote", func() {

					responseFixture := `{
						"products": [
							{
								"product_id": "BIT-31JAN25-CDE",
								"price": "97345",
								"price_percentage_change_24h": "-3.14412218297597",
								"volume_24h": "93744",
								"base_name": "",
								"status": "",
								"product_type": "FUTURE",
								"quote_currency_id": "USD",
								"fcm_trading_session_details": {
									"is_session_open": true
								},
								"base_display_symbol": "",
								"product_venue": "FCM",
								"future_product_details": {
									"venue": "cde",
									"contract_code": "BIT",
									"contract_expiry": "2025-01-31T16:00:00Z",
									"contract_root_unit": "BTC",
									"group_description": "Nano Bitcoin Futures",
									"contract_expiry_timezone": "Europe/London"
								}
							},
							{
								"base_display_symbol": "BTC",
								"product_type": "SPOT",
								"product_id": "BTC-USD",
								"base_name": "Bitcoin",
								"price": "50000.00",
								"price_percentage_change_24h": "2.0408163265306123",
								"volume_24h": "1500.50",
								"display_name": "Bitcoin",
								"status": "",
								"quote_currency_id": "USD",
								"product_venue": "CBE"
							}
						]
					}`
					responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BIT\-31JAN25\-CDE.*product_ids\=BTC\-USD.*`
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"BIT-31JAN25-CDE"}, []string{"BTC-USD"})
					Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Class":  Equal(c.AssetClassFuturesContract),
							"Symbol": Equal("BIT-31JAN25-CDE"),
							"Name":   Equal("Nano Bitcoin Futures"),
							"QuoteFutures": g.MatchFields(g.IgnoreExtras, g.Fields{
								"SymbolUnderlying": Equal("BTC"),
								"IndexPrice":       Equal(50000.00),
								"Basis":            Equal(-0.4863629359494581),
							}),
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Name":                    Equal("FCM"),
								"State":                   Equal(c.ExchangeStateOpen),
								"IsActive":                BeTrue(),
								"IsRegularTradingSession": BeTrue(),
							}),
						}),
					}))

				})
			})
		})

	})

	Describe("GetUnderlyingAssetSymbols", func() {
		It("should return the underlying asset symbols for a futures contract", func() {
			responseFixture := `{
				"products": [
					{
						"product_id": "BIT-31JAN25-CDE",
						"product_type": "FUTURE",
						"future_product_details": {
							"contract_root_unit": "BTC"
						}
					}
				]
			}`
			responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BIT\-31JAN25\-CDE.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output, err := GetUnderlyingAssetSymbols(*client, []string{"BIT-31JAN25-CDE"})
			Expect(output).To(Equal([]string{"BTC-USD"}))
			Expect(err).To(BeNil())
		})

		When("the asset is not a futures contract", func() {
			It("should ignore the asset", func() {
				responseFixture := `{
					"products": [
						{
							"product_id": "BTC-USD",
							"product_type": "SPOT"
						}
					]
				}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output, err := GetUnderlyingAssetSymbols(*client, []string{"BTC-USD"})
				Expect(output).To(BeEmpty())
				Expect(err).To(BeNil())
			})
		})

	})
})
