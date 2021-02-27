package quote_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/quote"
)

var _ = Describe("Quote", func() {
	Describe("GetQuotes", func() {
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

			inputCtx := c.Context{}
			output := GetQuotes(inputCtx, *client, []string{"NET"})()
			Expect(output[0].Price).To(Equal(84.98))
			Expect(output[0].PricePrevClose).To(Equal(84.00))
			Expect(output[0].PriceOpen).To(Equal(85.22))
			Expect(output[0].PriceDayHigh).To(Equal(90.00))
			Expect(output[0].PriceDayLow).To(Equal(80.00))
			Expect(output[0].Change).To(Equal(3.0800018))
			Expect(output[0].ChangePercent).To(Equal(3.7606857))
			Expect(output[0].IsActive).To(BeTrue())
			Expect(output[0].IsRegularTradingSession).To(BeTrue())
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

				inputCtx := c.Context{}
				output := GetQuotes(inputCtx, *client, []string{"NET"})()
				Expect(output[0].Price).To(Equal(86.03))
				Expect(output[0].Change).To(Equal(1.0399933))
				Expect(output[0].ChangePercent).To(Equal(1.2238094))
				Expect(output[0].IsActive).To(BeTrue())
				Expect(output[0].IsRegularTradingSession).To(BeFalse())
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

					inputCtx := c.Context{}
					output := GetQuotes(inputCtx, *client, []string{"NET"})()
					Expect(output[0].Price).To(Equal(84.98))
					Expect(output[0].Change).To(Equal(3.0800018))
					Expect(output[0].ChangePercent).To(Equal(3.7606857))
					Expect(output[0].IsActive).To(Equal(false))
					Expect(output[0].IsRegularTradingSession).To(Equal(false))
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

				inputCtx := c.Context{}
				output := GetQuotes(inputCtx, *client, []string{"NET"})()
				Expect(output[0].Price).To(Equal(86.02))
				Expect(output[0].Change).To(Equal(4.1199951))
				Expect(output[0].ChangePercent).To(Equal(4.9844951))
				Expect(output[0].IsActive).To(BeTrue())
				Expect(output[0].IsRegularTradingSession).To(BeFalse())
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

					inputCtx := c.Context{}
					output := GetQuotes(inputCtx, *client, []string{"NET"})()
					expectedPrice := 84.98
					expectedChange := 3.0800018
					expectedChangePercent := 3.7606857
					Expect(output[0].Price).To(Equal(expectedPrice))
					Expect(output[0].Change).To(Equal(expectedChange))
					Expect(output[0].ChangePercent).To(Equal(expectedChangePercent))
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

				inputCtx := c.Context{}
				output := GetQuotes(inputCtx, *client, []string{"NET"})()
				Expect(output[0].Price).To(Equal(84.98))
				Expect(output[0].Change).To(Equal(3.0800018))
				Expect(output[0].ChangePercent).To(Equal(3.7606857))
				Expect(output[0].IsActive).To(Equal(false))
				Expect(output[0].IsRegularTradingSession).To(Equal(false))
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

					inputCtx := c.Context{}
					output := GetQuotes(inputCtx, *client, []string{"NET"})()
					Expect(output[0].Price).To(Equal(86.02))
					Expect(output[0].Change).To(Equal(4.1199951))
					Expect(output[0].ChangePercent).To(Equal(4.9844951))
					Expect(output[0].IsActive).To(Equal(false))
					Expect(output[0].IsRegularTradingSession).To(Equal(false))
				})
			})
		})
	})
})
