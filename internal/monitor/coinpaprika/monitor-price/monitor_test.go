package monitorPriceCoinPaprika_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/achannarasappa/ticker/v5/internal/cache"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	monitorPriceCoinPaprika "github.com/achannarasappa/ticker/v5/internal/monitor/coinpaprika/monitor-price"

	"github.com/onsi/gomega/ghttp"
	"github.com/spf13/afero"
)

var responseTickerBitcoinFixture = map[string]interface{}{
	"id":     "btc-bitcoin",
	"name":   "Bitcoin",
	"symbol": "BTC",
	"rank":   1,
	"quotes": map[string]interface{}{
		"USD": map[string]interface{}{
			"price":              50000.0,
			"volume_24h":         21359052876.635258,
			"market_cap":         1264515966773.0,
			"percent_change_24h": 2.5,
		},
	},
}

var responseSearchBitcoinFixture = map[string]interface{}{
	"currencies": []map[string]interface{}{
		{
			"id":        "btc-bitcoin",
			"name":      "Bitcoin",
			"symbol":    "BTC",
			"rank":      1,
			"is_active": true,
			"type":      "coin",
		},
	},
}

var _ = Describe("Monitor CoinPaprika", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		server.RouteToHandler("GET", "/v1/search/",
			ghttp.RespondWithJSONEncoded(http.StatusOK, responseSearchBitcoinFixture),
		)
		server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
			ghttp.RespondWithJSONEncoded(http.StatusOK, responseTickerBitcoinFixture),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewMonitorPriceCoinPaprika", func() {
		It("should return a new monitor", func() {
			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			})
			Expect(monitor).NotTo(BeNil())
		})
	})

	Describe("SetSymbols", func() {
		It("should resolve ticker symbols to coin ids and get quotes", func() {
			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

			err := monitor.SetSymbols([]string{"BTC"}, 0)
			Expect(err).NotTo(HaveOccurred())

			assetQuotes, err := monitor.GetAssetQuotes()

			Expect(err).NotTo(HaveOccurred())
			Expect(assetQuotes).To(HaveLen(1))
			Expect(assetQuotes[0].Name).To(Equal("Bitcoin"))
			Expect(assetQuotes[0].Symbol).To(Equal("BTC.CP"))
			Expect(assetQuotes[0].QuotePrice.Price).To(Equal(50000.0))
			Expect(assetQuotes[0].QuoteSource).To(Equal(c.QuoteSourceCoinPaprika))
		})

		It("should request currency rates for USD once", func() {
			chanRequestCurrencyRates := make(chan []string, 2)

			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanRequestCurrencyRates: chanRequestCurrencyRates,
			}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

			err := monitor.SetSymbols([]string{"BTC"}, 0)
			Expect(err).NotTo(HaveOccurred())

			Expect(chanRequestCurrencyRates).To(Receive(Equal([]string{"USD"})))

			err = monitor.SetSymbols([]string{"BTC"}, 1)
			Expect(err).NotTo(HaveOccurred())

			Expect(chanRequestCurrencyRates).NotTo(Receive())
		})

		When("a symbol cannot be resolved to a coin id", func() {
			It("should skip the symbol", func() {
				server.RouteToHandler("GET", "/v1/search/",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
							if r.URL.Query().Get("q") == "BTC" {
								json.NewEncoder(w).Encode(responseSearchBitcoinFixture)

								return
							}

							json.NewEncoder(w).Encode(map[string]interface{}{"currencies": []map[string]interface{}{}})
						},
					),
				)

				monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

				err := monitor.SetSymbols([]string{"BTC", "NOTACOIN"}, 0)
				Expect(err).NotTo(HaveOccurred())

				assetQuotes, err := monitor.GetAssetQuotes()

				Expect(err).NotTo(HaveOccurred())
				Expect(assetQuotes).To(HaveLen(1))
				Expect(assetQuotes[0].Symbol).To(Equal("BTC.CP"))
			})
		})

		When("a cache is set", func() {
			It("should resolve coin ids from the cache without a search request", func() {
				sharedCache := cache.New(afero.NewMemMapFs(), "/cache/ticker/cache.json", true)

				monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
					Cache:                    sharedCache,
				}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

				err := monitor.SetSymbols([]string{"BTC"}, 0)
				Expect(err).NotTo(HaveOccurred())

				searchRequestCount := 0
				for _, request := range server.ReceivedRequests() {
					if request.URL.Path == "/v1/search/" {
						searchRequestCount++
					}
				}

				// Create a second monitor with the same cache which should not make a new search request
				monitorSecond := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
					Cache:                    sharedCache,
				}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

				err = monitorSecond.SetSymbols([]string{"BTC"}, 0)
				Expect(err).NotTo(HaveOccurred())

				searchRequestCountAfter := 0
				for _, request := range server.ReceivedRequests() {
					if request.URL.Path == "/v1/search/" {
						searchRequestCountAfter++
					}
				}

				Expect(searchRequestCount).To(Equal(1))
				Expect(searchRequestCountAfter).To(Equal(1))
			})
		})
	})

	Describe("GetAssetQuotes", func() {
		When("the ignoreCache flag is set to true", func() {
			It("should make a new request for quotes", func() {
				requestCount := 0

				server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, _ *http.Request) {
							requestCount++
							json.NewEncoder(w).Encode(responseTickerBitcoinFixture)
						},
					),
				)

				monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

				err := monitor.SetSymbols([]string{"BTC"}, 0)
				Expect(err).NotTo(HaveOccurred())

				requestCountAfterSetSymbols := requestCount

				_, err = monitor.GetAssetQuotes(true)
				Expect(err).NotTo(HaveOccurred())
				Expect(requestCount).To(Equal(requestCountAfterSetSymbols + 1))

				_, err = monitor.GetAssetQuotes()
				Expect(err).NotTo(HaveOccurred())
				Expect(requestCount).To(Equal(requestCountAfterSetSymbols + 1))
			})
		})
	})

	Describe("Start", func() {
		It("should start the monitor and send updates when the price changes", func() {
			price := 50000.0

			server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
				ghttp.CombineHandlers(
					func(w http.ResponseWriter, _ *http.Request) {
						price += 100.0
						response := map[string]interface{}{
							"id":     "btc-bitcoin",
							"name":   "Bitcoin",
							"symbol": "BTC",
							"quotes": map[string]interface{}{
								"USD": map[string]interface{}{
									"price":              price,
									"volume_24h":         21359052876.635258,
									"market_cap":         1264515966773.0,
									"percent_change_24h": 2.5,
								},
							},
						}
						json.NewEncoder(w).Encode(response)
					},
				),
			)

			chanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 5)

			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanError:                make(chan error, 5),
				ChanUpdateAssetQuote:     chanUpdateAssetQuote,
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

			err := monitor.SetSymbols([]string{"BTC"}, 0)
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			var update c.MessageUpdate[c.AssetQuote]
			Eventually(chanUpdateAssetQuote).Should(Receive(&update))
			Expect(update.ID).To(Equal("BTC.CP"))
			Expect(update.Data.QuotePrice.Price).To(BeNumerically(">", 50000.0))
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {
				monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

				err := monitor.Start()
				Expect(err).NotTo(HaveOccurred())

				err = monitor.Start()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Stop", func() {
		It("should stop the monitor", func() {
			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is not started", func() {
			It("should return an error", func() {
				monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
					Ctx:                      context.Background(),
					UnaryURL:                 server.URL(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("SetCurrencyRates", func() {
		It("should set the currency rates and update the asset quotes", func() {
			monitor := monitorPriceCoinPaprika.NewMonitorPriceCoinPaprika(monitorPriceCoinPaprika.Config{
				Ctx:                      context.Background(),
				UnaryURL:                 server.URL(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinPaprika.WithRefreshInterval(time.Millisecond*100))

			err := monitor.SetSymbols([]string{"BTC"}, 0)
			Expect(err).NotTo(HaveOccurred())

			err = monitor.SetCurrencyRates(c.CurrencyRates{
				"USD": c.CurrencyRate{
					FromCurrency: "USD",
					ToCurrency:   "EUR",
					Rate:         0.9,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			assetQuotes, err := monitor.GetAssetQuotes()

			Expect(err).NotTo(HaveOccurred())
			Expect(assetQuotes).To(HaveLen(1))
			Expect(assetQuotes[0].Currency.Rate).To(Equal(0.9))
			Expect(assetQuotes[0].Currency.ToCurrencyCode).To(Equal("EUR"))
		})
	})
})
