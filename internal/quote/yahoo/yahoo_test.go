package yahoo_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/quote/yahoo"
	. "github.com/achannarasappa/ticker/v4/test/http"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Yahoo Quote", func() {
	Describe("GetAssetQuotes", func() {
		It("should make a request to get stock quotes and transform the response", func() {
			responseFixture := `{
				"quoteResponse": {
					"result": [
						{
							"marketState": "REGULAR",
							"shortName": "Cloudflare, Inc.",
							"preMarketChange": { "raw": 1.0399933, "fmt": "1.0399933"},
							"preMarketChangePercent": { "raw": 1.2238094, "fmt": "1.2238094"},
							"preMarketPrice": { "raw": 86.03, "fmt": "86.03"},
							"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
							"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
							"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
							"regularMarketPreviousClose": { "raw": 84.00, "fmt": "84.00"},
							"regularMarketOpen": { "raw": 85.22, "fmt": "85.22"},
							"regularMarketDayHigh": { "raw": 90.00, "fmt": "90.00"},
							"regularMarketDayLow": { "raw": 80.00, "fmt": "80.00"},
							"postMarketChange": { "raw": 1.37627, "fmt": "1.37627"},
							"postMarketChangePercent": { "raw": 1.35735, "fmt": "1.35735"},
							"postMarketPrice": { "raw": 86.56, "fmt": "86.56"},
							"symbol": "NET"
						}
					],
					"error": null
				}
			}`
			responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"NET"})()
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(84.98),
						"PricePrevClose": Equal(84.00),
						"PriceOpen":      Equal(85.22),
						"PriceDayHigh":   Equal(90.00),
						"PriceDayLow":    Equal(80.00),
						"Change":         Equal(3.0800018),
						"ChangePercent":  Equal(3.7606857),
					}),
					"QuoteSource": Equal(c.QuoteSourceYahoo),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
				}),
			}))
		})

		When("the market is in a pre-market trading session", func() {
			It("should return the pre-market price", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "PRE",
								"shortName": "Cloudflare, Inc.",
								"preMarketChange": { "raw": 1.0399933, "fmt": "1.0399933"},
								"preMarketChangePercent": { "raw": 1.2238094, "fmt": "1.2238094"},
								"preMarketPrice": { "raw": 86.03, "fmt": "86.03"},
								"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
								"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
								"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price":         Equal(86.03),
							"Change":        Equal(1.0399933),
							"ChangePercent": Equal(1.2238094),
						}),
						"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
							"IsActive":                BeTrue(),
							"IsRegularTradingSession": BeFalse(),
						}),
					}),
				}))
			})

			When("there is no pre-market price", func() {
				It("should return the regular market price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "PRE",
									"shortName": "Cloudflare, Inc.",
									"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
									"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
									"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price":         Equal(84.98),
								"Change":        Equal(3.0800018),
								"ChangePercent": Equal(3.7606857),
							}),
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeFalse(),
								"IsRegularTradingSession": BeFalse(),
							}),
						}),
					}))
				})
			})
		})

		When("the market is in a post-market trading session", func() {
			It("should return the post-market price added to the regular market price", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "POST",
								"shortName": "Cloudflare, Inc.",
								"postMarketChange": { "raw": 1.0399933, "fmt": "1.0399933"},
								"postMarketChangePercent": { "raw": 1.2238094, "fmt": "1.2238094"},
								"postMarketPrice": { "raw": 86.02, "fmt": "86.02"},
								"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
								"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
								"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price":         Equal(86.02),
							"Change":        Equal(4.1199951),
							"ChangePercent": Equal(4.9844951),
						}),
						"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
							"IsActive":                BeTrue(),
							"IsRegularTradingSession": BeFalse(),
						}),
					}),
				}))
			})

			When("there is no post-market price", func() {
				It("should return the regular market price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "POST",
									"shortName": "Cloudflare, Inc.",
									"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
									"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
									"regularMarketTime": { "raw": 1623777601, "fmt": "4:00PM EDT"},
									"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
									"regularMarketPreviousClose": { "raw": 81.9, "fmt": "81.9"},
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price":         Equal(84.98),
								"Change":        Equal(3.0800018),
								"ChangePercent": Equal(3.7606857),
							}),
						}),
					}))
				})
			})
		})

		When("the market is closed", func() {
			It("should return the post-market price added to the regular market price", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "CLOSED",
								"shortName": "Cloudflare, Inc.",
								"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
								"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
								"regularMarketTime": { "raw": 1623777601, "fmt": "4:00PM EDT" },
								"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
								"regularMarketPreviousClose": { "raw": 81.9, "fmt": "81.9"},
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price":         Equal(84.98),
							"Change":        Equal(3.0800018),
							"ChangePercent": Equal(3.7606857),
						}),
						"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
							"IsActive":                BeFalse(),
							"IsRegularTradingSession": BeFalse(),
						}),
					}),
				}))
			})

			When("there is a post market price", func() {
				It("should show a closed state but with the post market change and price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "CLOSED",
									"shortName": "Cloudflare, Inc.",
									"postMarketChange": { "raw": 1.0399933, "fmt": "1.0399933"},
									"postMarketChangePercent": { "raw": 1.2238094, "fmt": "1.2238094"},
									"postMarketPrice": { "raw": 86.02, "fmt": "86.02"},
									"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
									"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
									"regularMarketTime": { "raw": 1623777601, "fmt": "4:00PM EDT" },
									"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
									"regularMarketPreviousClose": { "raw": 81.9, "fmt": "81.9"},
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := `=~\/finance\/quote.*symbols\=NET.*`
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price":         Equal(86.02),
								"Change":        Equal(4.1199951),
								"ChangePercent": Equal(4.9844951),
							}),
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeFalse(),
								"IsRegularTradingSession": BeFalse(),
							}),
						}),
					}))
				})
			})
		})

		When("the quote is for a cryptocurrency", func() {
			It("should should set the asset class to cryptocurrency", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "PRE",
								"shortName": "Cloudflare, Inc.",
								"preMarketChange": { "raw": 1.0399933, "fmt": "1.0399933"},
								"preMarketChangePercent": { "raw": 1.2238094, "fmt": "1.2238094"},
								"preMarketPrice": { "raw": 86.03, "fmt": "86.03"},
								"regularMarketChange": { "raw": 3.0800018, "fmt": "3.0800018"},
								"regularMarketChangePercent": { "raw": 3.7606857, "fmt": "3.7606857"},
								"regularMarketPrice": { "raw": 84.98, "fmt": "84.98"},
								"symbol": "BTC-USD",
								"quoteType": "CRYPTOCURRENCY"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=BTC\-USD`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"BTC-USD"})()
				Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
					"0": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Class": Equal(c.AssetClassCryptocurrency),
					}),
				}))
			})
		})
	})

	Describe("GetCurrencyRates", func() {
		It("should get the currency exchange rate", func() {

			MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
			MockResponse(ResponseParameters{Symbol: "EURUSD%3DX", Currency: "USD", Price: 1.2})
			output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "USD")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(c.CurrencyRates{
				"EUR": c.CurrencyRate{
					FromCurrency: "EUR",
					ToCurrency:   "USD",
					Rate:         1.2,
				},
			}))

		})

		When("target currency is not set", func() {
			It("defaults to USD", func() {

				MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
				MockResponse(ResponseParameters{Symbol: "EURUSD%3DX", Currency: "USD", Price: 1.2})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{
					"EUR": c.CurrencyRate{
						FromCurrency: "EUR",
						ToCurrency:   "USD",
						Rate:         1.2,
					},
				}))

			})
		})

		When("target currency is the same as all watchlist currencies", func() {
			It("returns an empty currency exchange rate list", func() {

				MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to get the currencies of each symbol fails", func() {
			It("returns error", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": { "raw": 1.2, "fmt": "1.2"},
								"currency": "EUR",
								"symbol": "EURUSD=X"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=EURUSD\=X.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).To(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to get the currencies of each symbol does not include a currency", func() {
			It("should return an empty currency rate", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": { "raw": 160.0, "fmt": "160.0"},
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=VOW3.DE.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to the exchange rate fails", func() {
			It("returns error", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": { "raw": 160.0, "fmt": "160.0"},
								"currency": "EUR",
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := `=~\/finance\/quote.*symbols\=VOW3.DE.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "USD")
				Expect(err).To(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})
	})

})
