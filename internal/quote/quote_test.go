package quote_test

import (
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "ticker-tape/internal/quote"
)

var _ = Describe("Quote", func() {
	Describe("GetQuotes", func() {
		It("should make a request to get stock quotes and transform the response", func() {
			var client = resty.New()
			httpmock.ActivateNonDefault(client.GetClient())
			responseFixture := `{
				"quoteResponse": {
					"result": [
						{
							"marketState": "REGULAR",
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

			output := GetQuotes(*client)([]string{"NET"})
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
						PreMarketChange:            1.0399933,
						PreMarketChangePercent:     1.2238094,
						PreMarketPrice:             86.02,
					},
					Price:                   84.98,
					Change:                  3.0800018,
					ChangePercent:           3.7606857,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			Expect(output).To(Equal(expected))
			httpmock.DeactivateAndReset()
		})
	})
})
