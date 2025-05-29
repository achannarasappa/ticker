package monitorPriceYahoo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	monitorPriceYahoo "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/monitor-price"
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Monitor Yahoo", func() {
	var (
		server   *ghttp.Server
		unaryAPI *unary.UnaryAPI
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		unaryAPI = unary.NewUnaryAPI(unary.Config{
			BaseURL: server.URL(),
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewMonitorPriceYahoo", func() {
		It("should return a new MonitorYahoo", func() {
			monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
				UnaryAPI:                 unaryAPI,
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			})
			Expect(monitor).NotTo(BeNil())
		})
	})

	Describe("GetAssetQuotes", func() {
		It("should return the asset quotes", func() {

			server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
				ghttp.CombineHandlers(
					func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query()
						fields := query.Get("fields")

						if fields == "regularMarketPrice,currency" {
							json.NewEncoder(w).Encode(currencyResponseFixture)
						} else {
							json.NewEncoder(w).Encode(responseQuote1Fixture)
						}
					},
				),
			)

			monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
				UnaryAPI:                 unaryAPI,
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceYahoo.WithRefreshInterval(time.Millisecond*100))

			monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

			assetQuotes, err := monitor.GetAssetQuotes()

			Expect(err).NotTo(HaveOccurred())
			Expect(len(assetQuotes)).To(Equal(2))
			Expect(assetQuotes[0].Symbol).To(Equal("NET"))
			Expect(assetQuotes[1].Symbol).To(Equal("GOOG"))
			Expect(assetQuotes[0].QuotePrice.Price).To(Equal(84.98))
			Expect(assetQuotes[1].QuotePrice.Price).To(Equal(166.25))
		})

		When("the http request fails", func() {
			It("should return an error", func() {

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
							query := r.URL.Query()
							fields := query.Get("fields")

							if fields == "regularMarketPrice,currency" {
								json.NewEncoder(w).Encode(currencyResponseFixture)
							} else {
								w.WriteHeader(http.StatusInternalServerError)
								w.Write([]byte(""))
							}
						},
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				assetQuotes, err := monitor.GetAssetQuotes(true)
				Expect(err).To(HaveOccurred())
				Expect(assetQuotes).To(BeEmpty())

			})
		})

		When("the ignoreCache flag is set to true", func() {
			It("should return the asset quotes from the cache", func() {
				var responseQuote1 unary.Response
				responseQuoteJSON, _ := json.Marshal(responseQuote1Fixture)
				json.Unmarshal(responseQuoteJSON, &responseQuote1)

				calledCount := 0

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
							if calledCount > 1 {
								responseQuote1.QuoteResponse.Quotes[0].RegularMarketPrice = unary.ResponseFieldFloat{
									Raw: 310.00,
									Fmt: "310.00",
								}
							}
							calledCount++

							json.NewEncoder(w).Encode(responseQuote1)
						},
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				monitor.SetSymbols([]string{"GOOG", "NET"}, 0)

				// First call to populate cache
				firstQuotes, err := monitor.GetAssetQuotes(true)
				Expect(err).NotTo(HaveOccurred())
				// Second call with ignoreCache=false should return cached data
				secondQuotes, err := monitor.GetAssetQuotes(false)
				Expect(err).NotTo(HaveOccurred())

				Expect(secondQuotes).To(HaveLen(2))
				Expect(secondQuotes[0].QuotePrice.Price).To(Equal(firstQuotes[0].QuotePrice.Price))
				Expect(secondQuotes[0].QuotePrice.ChangePercent).To(Equal(firstQuotes[0].QuotePrice.ChangePercent))
			})
		})
	})

	Describe("Start", func() {
		It("should start the monitor", func() {

			server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
				ghttp.CombineHandlers(
					func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query()
						fields := query.Get("fields")

						if fields == "regularMarketPrice,currency" {
							json.NewEncoder(w).Encode(currencyResponseFixture)
						} else {
							json.NewEncoder(w).Encode(responseQuote1Fixture)
						}
					},
				),
			)

			monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
				UnaryAPI:                 unaryAPI,
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceYahoo.WithRefreshInterval(time.Millisecond*100))

			monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
							query := r.URL.Query()
							fields := query.Get("fields")

							if fields == "regularMarketPrice,currency" {
								json.NewEncoder(w).Encode(currencyResponseFixture)
							} else {
								json.NewEncoder(w).Encode(responseQuote1Fixture)
							}
						},
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				err := monitor.Start()
				Expect(err).NotTo(HaveOccurred())

				err = monitor.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		When("the initial unary request for quotes fails", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote"),
						func(w http.ResponseWriter, r *http.Request) {
							query := r.URL.Query()
							fields := query.Get("fields")

							if fields == "regularMarketPrice,currency" {
								json.NewEncoder(w).Encode(currencyResponseFixture)
							} else {
								w.WriteHeader(http.StatusInternalServerError)
								w.Write([]byte(""))
							}
						},
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		When("the poller fails to start", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote"),
						func(w http.ResponseWriter, r *http.Request) {
							query := r.URL.Query()
							fields := query.Get("fields")

							if fields == "regularMarketPrice,currency" {
								json.NewEncoder(w).Encode(currencyResponseFixture)
							} else {
								w.WriteHeader(http.StatusInternalServerError)
								w.Write([]byte(""))
							}
						},
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("session refresh failed"))
			})
		})

		When("there is a polling asset update", func() {
			It("should send the updated asset quote to the channel", func() {

				var responseQuote1 unary.Response
				responseQuoteJSON, _ := json.Marshal(responseQuote1Fixture)
				json.Unmarshal(responseQuoteJSON, &responseQuote1)

				calledCount := 0

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
							query := r.URL.Query()
							fields := query.Get("fields")
							if fields == "regularMarketPrice,currency" {
								json.NewEncoder(w).Encode(currencyResponseFixture)
							} else if fields == "shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap" {
								if calledCount > 3 {

									quoteNewPrice := quoteCloudflareFixture
									quoteNewPrice.RegularMarketPrice = unary.ResponseFieldFloat{
										Raw: 310.00,
										Fmt: "310.00",
									}

									responseQuote1.QuoteResponse.Quotes = []unary.ResponseQuote{
										quoteNewPrice,
										quoteGoogleFixture,
									}
								}
								calledCount++

								json.NewEncoder(w).Encode(responseQuote1)
							}
						},
					),
				)

				// Create a channel to receive updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				// Create a monitor with a short refresh interval for testing
				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					ChanUpdateAssetQuote:     updateChan,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 100),
				}, monitorPriceYahoo.WithRefreshInterval(100*time.Millisecond))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)
				monitor.Start()

				// Wait for the update to be received on the channel
				var receivedQuote c.AssetQuote
				Eventually(func() float64 {
					select {
					case update := <-updateChan:
						receivedQuote = update.Data
						return receivedQuote.QuotePrice.Price
					default:
						return 0
					}
				}, 2*time.Second).Should(Equal(310.00))

				Expect(receivedQuote.Symbol).To(Equal("NET"))

				monitor.Stop()
			})

			When("the price has not changed", func() {
				It("should not send updates to the channel", func() {

					// Set up initial request handler
					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
						),
					)

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

					// Create a monitor with a short refresh interval
					monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
						UnaryAPI:                 unaryAPI,
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceYahoo.WithRefreshInterval(100*time.Millisecond))

					monitor.SetSymbols([]string{"NET", "GOOG"}, 0)
					monitor.Start()

					// Check that no updates are sent to the channel
					Consistently(func() bool {
						select {
						case <-updateChan:
							return true
						default:
							return false
						}
					}, 500*time.Millisecond).Should(BeFalse())

					monitor.Stop()
				})
			})

			When("the product does not exist in the cache", func() {
				It("should not send updates for that product", func() {
					responseQuote1 := unary.Response{
						QuoteResponse: unary.ResponseQuoteResponse{
							Quotes: []unary.ResponseQuote{
								quoteCloudflareFixture,
								quoteGoogleFixture,
							},
							Error: nil,
						},
					}

					receivedSymbols := make(map[string]bool)
					calledCount := 0

					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							func(w http.ResponseWriter, r *http.Request) {
								if calledCount > 1 {
									responseQuote1.QuoteResponse.Quotes = []unary.ResponseQuote{
										quoteCloudflareFixture,
										quoteGoogleFixture,
										quoteMetaFixture,
									}
								}
								calledCount++

								json.NewEncoder(w).Encode(responseQuote1)
							},
						),
					)

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

					monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
						UnaryAPI:                 unaryAPI,
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceYahoo.WithRefreshInterval(100*time.Millisecond))

					monitor.SetSymbols([]string{"GOOG", "NET"}, 0)
					monitor.Start()

					// Check for updates and record which symbols we receive
					go func() {
						for update := range updateChan {
							receivedSymbols[update.Data.Symbol] = true
						}
					}()

					Consistently(func() bool {
						_, hasFB := receivedSymbols["FB"]
						return hasFB
					}, 2*time.Second).Should(BeFalse())

					monitor.Stop()
				})
			})
		})

		When("there is a currency rate update", func() {
			It("should replace the currency rate cache", func() {
				var err error

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				// Create channels for currency rate updates and asset quote updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				// Create and start the monitor
				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					ChanUpdateAssetQuote:     updateChan,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceYahoo.WithRefreshInterval(100*time.Millisecond))

				monitor.SetSymbols([]string{"NET"}, 0)
				err = monitor.Start()

				// Send currency rates update
				currencyRates := c.CurrencyRates{
					"USD": c.CurrencyRate{
						FromCurrency: "USD",
						ToCurrency:   "EUR",
						Rate:         0.85,
					},
				}
				err = monitor.SetCurrencyRates(currencyRates)
				Expect(err).NotTo(HaveOccurred())

				quotes, err := monitor.GetAssetQuotes()
				Expect(err).NotTo(HaveOccurred())
				Expect(quotes[0].Currency.Rate).To(Equal(0.85))
				Expect(quotes[0].Currency.FromCurrencyCode).To(Equal("USD"))
				Expect(quotes[0].Currency.ToCurrencyCode).To(Equal("EUR"))

				monitor.Stop()
			})

			When("there is an error getting asset quotes and replacing the cache after new currency rates are recieved", func() {
				It("should send a message to the error channel", func() {

					var err error
					var respondWithError bool = false

					server.RouteToHandler("GET", "/v7/finance/quote",
						func(w http.ResponseWriter, r *http.Request) {
							if respondWithError {
								w.Write([]byte("invalid"))
							} else {
								json.NewEncoder(w).Encode(responseQuote1Fixture)
							}
						},
					)

					// Create channels for currency rate updates and asset quote updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)
					errorChan := make(chan error, 1)

					// Create and start the monitor
					monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
						UnaryAPI:                 unaryAPI,
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
						ChanError:                errorChan,
					}, monitorPriceYahoo.WithRefreshInterval(100*time.Millisecond))

					err = monitor.SetSymbols([]string{"NET"}, 0)
					Expect(err).NotTo(HaveOccurred())

					err = monitor.Start()
					Expect(err).NotTo(HaveOccurred())

					// Send currency rates update
					currencyRates := c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         0.85,
						},
					}

					respondWithError = true
					err = monitor.SetCurrencyRates(currencyRates)

					Expect(err).To(MatchError(ContainSubstring("failed to decode response")))

					monitor.Stop()
				})
			})
		})
	})

	Describe("Stop", func() {
		It("should stop the monitor", func() {

			server.RouteToHandler("GET", "/v7/finance/quote",
				ghttp.CombineHandlers(
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
				UnaryAPI:                 unaryAPI,
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceYahoo.WithRefreshInterval(10*time.Second))

			monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Stop()
			Expect(err).NotTo(HaveOccurred())

		})

		When("the monitor is not started", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})

	Describe("SetSymbols", func() {

		When("there are no symbols set", func() {

			It("should not request a currency mapping", func() {

				var calledGetCurrencyMap bool = false

				server.RouteToHandler("GET", "/v7/finance/quote",
					func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query()
						fields := query.Get("fields")

						if fields == "regularMarketPrice,currency" {
							calledGetCurrencyMap = true
							json.NewEncoder(w).Encode(currencyResponseFixture)
						} else {
							json.NewEncoder(w).Encode(responseQuote1Fixture)
						}
					},
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				err := monitor.SetSymbols([]string{}, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(calledGetCurrencyMap).To(BeFalse())

				monitor.Start()

			})
		})

		When("there is an error getting the currency mapping", func() {

			It("should return an error", func() {

				server.RouteToHandler("GET", "/v7/finance/quote",
					func(w http.ResponseWriter, r *http.Request) {
						query := r.URL.Query()
						fields := query.Get("fields")

						if fields == "regularMarketPrice,currency" {
							w.Write([]byte("invalid"))
						} else {
							json.NewEncoder(w).Encode(responseQuote1Fixture)
						}
					},
				)

				monitor := monitorPriceYahoo.NewMonitorPriceYahoo(monitorPriceYahoo.Config{
					UnaryAPI:                 unaryAPI,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				err := monitor.SetSymbols([]string{"NET"}, 0)
				Expect(err).To(MatchError(ContainSubstring("failed to get currency information")))

				monitor.Start()

			})
		})

	})

})
