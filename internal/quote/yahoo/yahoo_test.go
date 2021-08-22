package yahoo_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/quote/yahoo"
	. "github.com/achannarasappa/ticker/test/http"
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
							"preMarketChange": 1.0399933,
							"preMarketChangePercent": 1.2238094,
							"preMarketPrice": 86.03,
							"regularMarketChange": 3.0800018,
							"regularMarketChangePercent": 3.7606857,
							"regularMarketPrice": 84.98,
							"regularMarketPreviousClose": 84.00,
							"regularMarketOpen": 85.22,
							"regularMarketDayHigh": 90.00,
							"regularMarketDayLow": 80.00,
							"postMarketChange": 1.37627,
							"postMarketChangePercent": 1.35735,
							"postMarketPrice": 86.56,
							"symbol": "NET"
						}
					],
					"error": null
				}
			}`
			responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})

			output := GetAssetQuotes(*client, []string{"NET"})()
			Expect(output[0].QuotePrice.Price).To(Equal(84.98))
			Expect(output[0].QuotePrice.PricePrevClose).To(Equal(84.00))
			Expect(output[0].QuotePrice.PriceOpen).To(Equal(85.22))
			Expect(output[0].QuotePrice.PriceDayHigh).To(Equal(90.00))
			Expect(output[0].QuotePrice.PriceDayLow).To(Equal(80.00))
			Expect(output[0].QuotePrice.Change).To(Equal(3.0800018))
			Expect(output[0].QuotePrice.ChangePercent).To(Equal(3.7606857))
			Expect(output[0].Exchange.IsActive).To(BeTrue())
			Expect(output[0].Exchange.IsRegularTradingSession).To(BeTrue())
		})

		When("the market is in a pre-market trading session", func() {
			It("should return the pre-market price", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "PRE",
								"shortName": "Cloudflare, Inc.",
								"preMarketChange": 1.0399933,
								"preMarketChangePercent": 1.2238094,
								"preMarketPrice": 86.03,
								"regularMarketChange": 3.0800018,
								"regularMarketChangePercent": 3.7606857,
								"regularMarketPrice": 84.98,
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output[0].QuotePrice.Price).To(Equal(86.03))
				Expect(output[0].QuotePrice.Change).To(Equal(1.0399933))
				Expect(output[0].QuotePrice.ChangePercent).To(Equal(1.2238094))
				Expect(output[0].Exchange.IsActive).To(BeTrue())
				Expect(output[0].Exchange.IsRegularTradingSession).To(BeFalse())
			})

			When("there is no pre-market price", func() {
				It("should return the regular market price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "PRE",
									"shortName": "Cloudflare, Inc.",
									"regularMarketChange": 3.0800018,
									"regularMarketChangePercent": 3.7606857,
									"regularMarketPrice": 84.98,
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					Expect(output[0].QuotePrice.Price).To(Equal(84.98))
					Expect(output[0].QuotePrice.Change).To(Equal(3.0800018))
					Expect(output[0].QuotePrice.ChangePercent).To(Equal(3.7606857))
					Expect(output[0].Exchange.IsActive).To(Equal(false))
					Expect(output[0].Exchange.IsRegularTradingSession).To(Equal(false))
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
								"postMarketChange": 1.0399933,
								"postMarketChangePercent": 1.2238094,
								"postMarketPrice": 86.02,
								"regularMarketChange": 3.0800018,
								"regularMarketChangePercent": 3.7606857,
								"regularMarketPrice": 84.98,
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output[0].QuotePrice.Price).To(Equal(86.02))
				Expect(output[0].QuotePrice.Change).To(Equal(4.1199951))
				Expect(output[0].QuotePrice.ChangePercent).To(Equal(4.9844951))
				Expect(output[0].Exchange.IsActive).To(BeTrue())
				Expect(output[0].Exchange.IsRegularTradingSession).To(BeFalse())
			})

			When("there is no post-market price", func() {
				It("should return the regular market price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "POST",
									"shortName": "Cloudflare, Inc.",
									"regularMarketChange": 3.0800018,
									"regularMarketChangePercent": 3.7606857,
									"regularMarketTime": 1608832801,
									"regularMarketPrice": 84.98,
									"regularMarketPreviousClose": 81.9,
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					expectedPrice := 84.98
					expectedChange := 3.0800018
					expectedChangePercent := 3.7606857
					Expect(output[0].QuotePrice.Price).To(Equal(expectedPrice))
					Expect(output[0].QuotePrice.Change).To(Equal(expectedChange))
					Expect(output[0].QuotePrice.ChangePercent).To(Equal(expectedChangePercent))
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
								"regularMarketChange": 3.0800018,
								"regularMarketChangePercent": 3.7606857,
								"regularMarketTime": 1608832801,
								"regularMarketPrice": 84.98,
								"regularMarketPreviousClose": 81.9,
								"symbol": "NET"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

				output := GetAssetQuotes(*client, []string{"NET"})()
				Expect(output[0].QuotePrice.Price).To(Equal(84.98))
				Expect(output[0].QuotePrice.Change).To(Equal(3.0800018))
				Expect(output[0].QuotePrice.ChangePercent).To(Equal(3.7606857))
				Expect(output[0].Exchange.IsActive).To(Equal(false))
				Expect(output[0].Exchange.IsRegularTradingSession).To(Equal(false))
			})

			When("there is a post market price", func() {
				It("should show a closed state but with the post market change and price", func() {
					responseFixture := `{
						"quoteResponse": {
							"result": [
								{
									"marketState": "CLOSED",
									"shortName": "Cloudflare, Inc.",
									"postMarketChange": 1.0399933,
									"postMarketChangePercent": 1.2238094,
									"postMarketPrice": 86.02,
									"regularMarketChange": 3.0800018,
									"regularMarketChangePercent": 3.7606857,
									"regularMarketTime": 1608832801,
									"regularMarketPrice": 84.98,
									"regularMarketPreviousClose": 81.9,
									"symbol": "NET"
								}
							],
							"error": null
						}
					}`
					responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=NET"
					httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
						resp := httpmock.NewStringResponse(200, responseFixture)
						resp.Header.Set("Content-Type", "application/json")
						return resp, nil
					})

					output := GetAssetQuotes(*client, []string{"NET"})()
					Expect(output[0].QuotePrice.Price).To(Equal(86.02))
					Expect(output[0].QuotePrice.Change).To(Equal(4.1199951))
					Expect(output[0].QuotePrice.ChangePercent).To(Equal(4.9844951))
					Expect(output[0].Exchange.IsActive).To(Equal(false))
					Expect(output[0].Exchange.IsRegularTradingSession).To(Equal(false))
				})
			})
		})
	})

	Describe("GetCurrencyRates", func() {
		It("should get the currency exchange rate", func() {

			MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
			MockResponse(ResponseParameters{Symbol: "EURUSD=X", Currency: "USD", Price: 1.2})
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
				MockResponse(ResponseParameters{Symbol: "EURUSD=X", Currency: "USD", Price: 1.2})
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
								"regularMarketPrice": 1.2,
								"currency": "EUR",
								"symbol": "EURUSD=X"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=EURUSD=X"
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
								"regularMarketPrice": 160.0,
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=VOW3.DE"
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
								"regularMarketPrice": 160.0,
								"currency": "EUR",
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=VOW3.DE"
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
