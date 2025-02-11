package unary_test

// import (
// 	"net/http"
// 	"time"

// 	c "github.com/achannarasappa/ticker/v4/internal/common"
// 	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
// 	. "github.com/achannarasappa/ticker/v4/test/http"
// 	"github.com/jarcoal/httpmock"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	g "github.com/onsi/gomega/gstruct"
// )

// func mockResponseCurrencyGOOG() {
// 	response := `{
// 		"quoteResponse": {
// 			"result": [
// 				{
// 					"currency": "USD",
// 					"symbol": "GOOG"
// 				}
// 			],
// 			"error": null
// 		}
// 	}
// 	`
// 	responseURL := `=~\/finance\/quote.*symbols\=GOOG`
// 	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
// 		resp := httpmock.NewStringResponse(200, response)
// 		resp.Header.Set("Content-Type", "application/json")
// 		return resp, nil
// 	})
// }

// func mockResponseCurrencyEURUSD() {
// 	response := `{
// 		"quoteResponse": {
// 			"result": [
// 				{
// 					"quoteType": "CURRENCY",
// 					"quoteSourceName": "Delayed Quote",
// 					"currency": "EUR",
// 					"regularMarketPrice": {"raw": 0.8891,"fmt": "0.8891"},
// 					"sourceInterval": 15,
// 					"exchangeDataDelayedBy": 0,
// 					"exchange": "CCY",
// 					"fullExchangeName": "CCY",
// 					"symbol": "USDEUR=X"
// 				}
// 			],
// 			"error": null
// 		}
// 	}

// 	`
// 	responseURL := `=~\/finance\/quote.*symbols\=USDEUR.*X`

// 	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
// 		resp := httpmock.NewStringResponse(200, response)
// 		resp.Header.Set("Content-Type", "application/json")
// 		return resp, nil
// 	})
// }

// var (
// 	monitors c.MonitorsDependencies
// )

// var _ = Describe("Quote", func() {

// 	BeforeEach(func() {
// 		coinbase := monitorCoinbase.NewMonitorCoinbase(
// 			monitorCoinbase.Config{
// 				Client:   *client,
// 				OnUpdate: func() {},
// 			},
// 			monitorCoinbase.WithRefreshInterval(time.Duration(0)),
// 		)
// 		var monitor c.Monitor = coinbase
// 		monitors = c.MonitorsDependencies{
// 			Coinbase: &monitor,
// 			HttpClients: c.DependenciesHttpClients{
// 				Default: client,
// 				Yahoo:   client,
// 			},
// 		}
// 		MockResponseCoinbaseQuotes()
// 	})

// 	Describe("GetAssetQuotes", func() {

// 		It("should get price quotes from Coinbase", func() {

// 			(*monitors.Coinbase).SetSymbols([]string{"ADA-USD", "ADA-31JAN25-CDE"})
// 			output := (*monitors.Coinbase).GetAssetQuotes(true)

// 			Expect(output).To(HaveLen(2))

// 			Expect(output).To(ContainElement(g.MatchFields(g.IgnoreExtras, g.Fields{
// 				"Symbol": Equal("ADA"),
// 				"Name":   Equal("Cardano"),
// 				"Class":  Equal(c.AssetClassCryptocurrency),
// 				"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
// 					"Price":          Equal(50000.00),
// 					"ChangePercent":  Equal(2.0408163265306123),
// 					"PriceDayHigh":   BeZero(),
// 					"PriceDayLow":    BeZero(),
// 					"PriceOpen":      BeZero(),
// 					"PricePrevClose": BeZero(),
// 					"Change":         BeNumerically("~", 1020, 1.0),
// 				}),
// 				"QuoteExtended": g.MatchFields(g.IgnoreExtras, g.Fields{
// 					"Volume":    Equal(1500.50),
// 					"MarketCap": BeZero(),
// 				}),
// 				"Currency": g.MatchFields(g.IgnoreExtras, g.Fields{
// 					"FromCurrencyCode": Equal("USD"),
// 					"ToCurrencyCode":   Equal(""),
// 				}),
// 				"QuoteSource": Equal(c.QuoteSourceCoinbase),
// 			})))

// 		})

// 		When("there are no symbols", func() {
// 			It("should return an empty slice", func() {
// 				output := (*monitors.Coinbase).GetAssetQuotes(true)
// 				Expect(output).To(BeEmpty())
// 			})
// 		})

// 		When("the quotes are cached", func() {
// 			It("should return the cached quotes", func() {
// 				(*monitors.Coinbase).SetSymbols([]string{"ADA-USD", "ADA-31JAN25-CDE"})
// 				(*monitors.Coinbase).GetAssetQuotes(true)

// 				output := (*monitors.Coinbase).GetAssetQuotes()
// 				Expect(output).To(HaveLen(2))
// 			})
// 		})

// 	})

// })
