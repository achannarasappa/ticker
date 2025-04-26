package monitorYahoo_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorYahoo "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo"
	"github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Monitor Yahoo", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewMonitorYahoo", func() {
		It("should return a new MonitorYahoo", func() {
			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
				Ctx:      context.Background(),
			})
			Expect(monitor).NotTo(BeNil())
		})
	})

	Describe("GetAssetQuotes", func() {
		It("should return the asset quotes", func() {

			server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
				Ctx:      context.Background(),
			}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

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
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				assetQuotes, err := monitor.GetAssetQuotes(true)
				Expect(err).To(HaveOccurred())
				Expect(assetQuotes).To(BeEmpty())

			})
		})

		When("the ignoreCache flag is set to true", func() {
			It("should return the asset quotes from the cache", func() {
				responseQuote1 := responseQuote1Fixture

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

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
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
					ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
				Ctx:      context.Background(),
			}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

			monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

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
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		When("the poller fails to start", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
				})

				monitor.SetSymbols([]string{"NET", "GOOG"}, 0)

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("refresh interval is not set"))
			})
		})

		When("there is a polling asset update", func() {
			It("should send the updated asset quote to the channel", func() {

				responseQuote1 := responseQuote1Fixture

				calledCount := 0

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {
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
						},
					),
				)

				// Create a channel to receive updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				// Create a monitor with a short refresh interval for testing
				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL:             server.URL(),
					ChanUpdateAssetQuote: updateChan,
					Ctx:                  context.Background(),
				}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

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
					monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
						UnaryURL:             server.URL(),
						ChanUpdateAssetQuote: updateChan,
						Ctx:                  context.Background(),
					}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

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

					monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
						UnaryURL:             server.URL(),
						ChanUpdateAssetQuote: updateChan,
						Ctx:                  context.Background(),
					}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

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
	})

	Describe("Stop", func() {
		It("should stop the monitor", func() {

			server.RouteToHandler("GET", "/v7/finance/quote",
				ghttp.CombineHandlers(
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
				Ctx:      context.Background(),
			}, monitorYahoo.WithRefreshInterval(10*time.Second))

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

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
					Ctx:      context.Background(),
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})
})
