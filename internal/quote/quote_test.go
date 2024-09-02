package quote_test

import (
	"net/http"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/quote"
	. "github.com/achannarasappa/ticker/v4/test/http"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/onsi/gomega/gstruct"
)

func mockResponseCurrencyGOOG() {
	response := `{
		"quoteResponse": {
			"result": [
				{
					"currency": "USD",
					"symbol": "GOOG"
				}
			],
			"error": null
		}
	}
	`
	responseURL := `=~\/finance\/quote.*symbols\=GOOG`
	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")
		return resp, nil
	})
}

func mockResponseCurrencyEURUSD() {
	response := `{
		"quoteResponse": {
			"result": [
				{
					"quoteType": "CURRENCY",
					"quoteSourceName": "Delayed Quote",
					"currency": "EUR",
					"regularMarketPrice": {"raw": 0.8891,"fmt": "0.8891"},
					"sourceInterval": 15,
					"exchangeDataDelayedBy": 0,
					"exchange": "CCY",
					"fullExchangeName": "CCY",
					"symbol": "USDEUR=X"
				}
			],
			"error": null
		}
	}
	
	`
	responseURL := `=~\/finance\/quote.*symbols\=USDEUR.*X`

	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")
		return resp, nil
	})
}

var _ = Describe("Quote", func() {

	var (
		dep c.Dependencies
	)

	BeforeEach(func() {
		dep = c.Dependencies{
			HttpClients: c.DependenciesHttpClients{
				Default: client,
				Yahoo:   client,
			},
		}
		MockResponseYahooQuotes()
		MockResponseCoingeckoQuotes()
		MockResponseCoincapQuotes()
	})

	Describe("GetAssetGroupQuote", func() {

		It("should get price quotes for each asset based on it's data source", func() {

			input := c.AssetGroup{
				SymbolsBySource: []c.AssetGroupSymbolsBySource{
					{
						Source: c.QuoteSourceYahoo,
						Symbols: []c.Symbol{
							{ID: "GOOG", Name: "GOOG"},
							{ID: "RBLX", Name: "RBLX"},
						},
					},
					{
						Source: c.QuoteSourceCoingecko,
						Symbols: []c.Symbol{
							{ID: "BITCOIN.CG", Name: "bitcoin"},
						},
					},
					{
						Source: c.QuoteSourceCoinCap,
						Symbols: []c.Symbol{
							{ID: "ELROND.CC", Name: "elrond"},
						},
					},
					{
						Source: c.QuoteSourceUserDefined,
						Symbols: []c.Symbol{
							{ID: "CASH", Name: "CASH"},
							{ID: "PRIVATESHARES", Name: "PRIVATESHARES"},
						},
					},
				},
			}
			output := GetAssetGroupQuote(dep)(input)

			idFn := func(e interface{}) string { return e.(c.AssetQuote).Symbol }

			Expect(output).To(g.MatchFields(g.IgnoreExtras, g.Fields{
				"AssetQuotes": g.MatchElements(idFn, g.IgnoreExtras, g.Elements{
					"GOOG": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuoteSource": Equal(c.QuoteSourceYahoo),
					}),
					"BTC": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuoteSource": Equal(c.QuoteSourceCoingecko),
					}),
					"EGLD": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuoteSource": Equal(c.QuoteSourceCoinCap),
					}),
				}),
			}))

		})

	})

	Describe("GetAssetGroupsCurrencyRates", func() {

		It("should get currency conversion rates for each type of data source", func() {

			mockResponseCurrencyEURUSD()
			mockResponseCurrencyGOOG()
			input := []c.AssetGroup{
				{
					SymbolsBySource: []c.AssetGroupSymbolsBySource{
						{
							Source: c.QuoteSourceYahoo,
							Symbols: []c.Symbol{
								{ID: "GOOG", Name: "GOOG"},
							},
						},
					},
				},
			}
			output, _ := GetAssetGroupsCurrencyRates(client, input, "EUR")
			Expect(output).To(g.MatchAllKeys(g.Keys{
				"USD": g.MatchFields(g.IgnoreExtras, g.Fields{
					"FromCurrency": Equal("USD"),
					"ToCurrency":   Equal("EUR"),
					"Rate":         Equal(0.8891),
				}),
			}))

		})

	})

})
