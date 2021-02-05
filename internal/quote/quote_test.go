package quote_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

			output := GetQuotes(*client, []string{"NET"})()
			expected := []Quote{
				{
					ResponseQuote: ResponseQuote{
						ShortName:                  "Cloudflare, Inc.",
						Symbol:                     "NET",
						MarketState:                "REGULAR",
						RegularMarketChange:        3.0800018,
						RegularMarketChangePercent: 3.7606857,
						RegularMarketPrice:         84.98,
						RegularMarketPreviousClose: 81.9,
					},
					Price:                   84.98,
					Change:                  3.0800018,
					ChangePercent:           3.7606857,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			Expect(output).To(Equal(expected))
		})

		Context("when the market is in a pre-market trading session", func() {
			It("should return the pre-market price", func() {
				responseFixture := `{
					"quoteResponse": {
						"result": [
							{
								"marketState": "PRE",
								"shortName": "Cloudflare, Inc.",
								"preMarketChange": 1.0399933,
								"preMarketChangePercent": 1.2238094,
								"preMarketPrice": 86.02,
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

				output := GetQuotes(*client, []string{"NET"})()
				expected := []Quote{
					{
						ResponseQuote: ResponseQuote{
							ShortName:                  "Cloudflare, Inc.",
							Symbol:                     "NET",
							MarketState:                "PRE",
							RegularMarketChange:        3.0800018,
							RegularMarketChangePercent: 3.7606857,
							RegularMarketPrice:         84.98,
							RegularMarketPreviousClose: 81.9,
							PreMarketChange:            1.0399933,
							PreMarketChangePercent:     1.2238094,
							PreMarketPrice:             86.02,
						},
						Price:                   86.02,
						Change:                  1.0399933,
						ChangePercent:           1.2238094,
						IsActive:                true,
						IsRegularTradingSession: false,
					},
				}
				Expect(output).To(Equal(expected))
			})
		})

		Context("when the market is in a post-market trading session", func() {
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

				output := GetQuotes(*client, []string{"NET"})()
				expected := []Quote{
					{
						ResponseQuote: ResponseQuote{
							ShortName:                  "Cloudflare, Inc.",
							Symbol:                     "NET",
							MarketState:                "POST",
							RegularMarketChange:        3.0800018,
							RegularMarketChangePercent: 3.7606857,
							RegularMarketPrice:         84.98,
							RegularMarketPreviousClose: 81.9,
							PostMarketChange:           1.0399933,
							PostMarketChangePercent:    1.2238094,
							PostMarketPrice:            86.02,
						},
						Price:                   86.02,
						Change:                  4.1199951,
						ChangePercent:           4.9844951,
						IsActive:                true,
						IsRegularTradingSession: false,
					},
				}
				Expect(output).To(Equal(expected))
			})
		})

		Context("when the market is CLOSED", func() {
			It("should return the post-market price added to the regular market price", func() {
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

				output := GetQuotes(*client, []string{"NET"})()
				expected := []Quote{
					{
						ResponseQuote: ResponseQuote{
							ShortName:                  "Cloudflare, Inc.",
							Symbol:                     "NET",
							MarketState:                "CLOSED",
							RegularMarketChange:        3.0800018,
							RegularMarketChangePercent: 3.7606857,
							RegularMarketPrice:         84.98,
							RegularMarketPreviousClose: 81.9,
							PostMarketChange:           1.0399933,
							PostMarketChangePercent:    1.2238094,
							PostMarketPrice:            86.02,
						},
						Price:                   84.98,
						Change:                  0.0,
						ChangePercent:           0.0,
						IsActive:                false,
						IsRegularTradingSession: false,
					},
				}
				Expect(output).To(Equal(expected))
			})
		})
	})
})
