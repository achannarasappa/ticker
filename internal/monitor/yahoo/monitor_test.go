package monitorYahoo_test

import (
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
			})
			Expect(monitor).NotTo(BeNil())
		})
	})

	Describe("GetAssetQuotes", func() {
		It("should return the asset quotes", func() {

			server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
			}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

			outputAssetQuotes := []c.AssetQuote{}
			monitor.SetOnUpdateAssetQuotes(func(assetQuotes []c.AssetQuote) {
				outputAssetQuotes = assetQuotes
			})

			monitor.SetSymbols([]string{"NET", "GOOG"})

			assetQuotes, err := monitor.GetAssetQuotes()

			Expect(err).NotTo(HaveOccurred())
			Expect(len(assetQuotes)).To(Equal(2))
			Expect(assetQuotes[0].Symbol).To(Equal("NET"))
			Expect(assetQuotes[1].Symbol).To(Equal("GOOG"))
			Expect(assetQuotes[0].QuotePrice.Price).To(Equal(84.98))
			Expect(assetQuotes[1].QuotePrice.Price).To(Equal(166.25))
			Expect(outputAssetQuotes).To(HaveLen(2))
			Expect(outputAssetQuotes[0].Symbol).To(Equal("NET"))
			Expect(outputAssetQuotes[0].Name).To(Equal("Cloudflare, Inc."))
			Expect(outputAssetQuotes[0].Class).To(Equal(c.AssetClassStock))
			Expect(outputAssetQuotes[1].Symbol).To(Equal("GOOG"))
			Expect(outputAssetQuotes[1].Name).To(Equal("Google Inc."))
			Expect(outputAssetQuotes[1].Class).To(Equal(c.AssetClassStock))
		})

		When("the http request fails", func() {
			It("should return an error", func() {

				server.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"})

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
				})

				monitor.SetSymbols([]string{"GOOG", "NET"})

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
					ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
				UnaryURL: server.URL(),
			}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

			monitor.SetSymbols([]string{"NET", "GOOG"})

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"})

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
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
				}, monitorYahoo.WithRefreshInterval(time.Millisecond*100))

				monitor.SetSymbols([]string{"NET", "GOOG"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		When("the poller fails to start", func() {
			It("should return an error", func() {

				server.RouteToHandler(http.MethodGet, "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=GOOG,NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
					),
				)

				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
				})

				monitor.SetSymbols([]string{"NET", "GOOG"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("refresh interval is not set"))
			})
		})

		When("there is a polling asset update", func() {
			It("should call the onUpdate function with the updated asset quote", func() {

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

				var outputAssetQuote c.AssetQuote
				var outputCalled bool

				// Create a monitor with a short refresh interval for testing
				monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
					UnaryURL: server.URL(),
				}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

				monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
					outputAssetQuote = assetQuote
					outputCalled = true
				})

				monitor.SetSymbols([]string{"NET", "GOOG"})
				monitor.Start()

				// Wait for the update to be called
				Eventually(func() bool {
					return outputCalled
				}, 2*time.Second).Should(BeTrue())
				Expect(outputAssetQuote.QuotePrice.Price).To(Equal(310.00))
				Expect(outputAssetQuote.Symbol).To(Equal("NET"))

				monitor.Stop()

			})

			When("the price has not changed", func() {
				It("should not call the onUpdate function", func() {

					// Set up initial request handler
					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
						),
					)

					// Detect if the onUpdate function is called
					outputCalled := false

					// Create a monitor with a short refresh interval
					monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
						UnaryURL: server.URL(),
					}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						outputCalled = true
					})

					monitor.SetSymbols([]string{"NET", "GOOG"})
					monitor.Start()

					// Wait a bit to ensure the polling happens
					Consistently(func() bool {
						return outputCalled
					}, 500*time.Millisecond).Should(BeFalse())

					monitor.Stop()
				})
			})

			When("the product does not exist in the cache", func() {
				It("should not call the onUpdate function", func() {
					responseQuote1 := unary.Response{
						QuoteResponse: unary.ResponseQuoteResponse{
							Quotes: []unary.ResponseQuote{
								quoteCloudflareFixture,
								quoteGoogleFixture,
							},
							Error: nil,
						},
					}

					outputSymbols := make(map[string]bool)
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
					monitor := monitorYahoo.NewMonitorYahoo(monitorYahoo.Config{
						UnaryURL: server.URL(),
					}, monitorYahoo.WithRefreshInterval(100*time.Millisecond))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						outputSymbols[symbol] = true
					})

					monitor.SetSymbols([]string{"GOOG", "NET"})

					monitor.Start()

					Consistently(func() bool {
						_, hasFB := outputSymbols["FB"]
						return hasFB
					}, 2*time.Second).Should(BeFalse())
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
			}, monitorYahoo.WithRefreshInterval(10*time.Second))

			monitor.SetSymbols([]string{"NET", "GOOG"})

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
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})
})
