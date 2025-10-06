package monitorCurrencyRate_test

import (
	"math"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"context"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	. "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/monitor-currency-rates"
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
)

var _ = Describe("MonitorCurrencyRates", func() {

	var (
		server                   *ghttp.Server
		client                   *unary.UnaryAPI
		float64EqualityTolerance = 1e-9
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		client = unary.NewUnaryAPI(unary.Config{
			BaseURL:           server.URL(),
			SessionRootURL:    server.URL(),
			SessionCrumbURL:   server.URL(),
			SessionConsentURL: server.URL(),
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Start", func() {

		It("should start the monitor", func() {

			monitor := NewMonitorCurrencyRateYahoo(Config{
				Ctx:                      context.Background(),
				UnaryAPI:                 client,
				ChanUpdateCurrencyRates:  make(chan c.CurrencyRates),
				ChanRequestCurrencyRates: make(chan []string),
				ChanError:                make(chan error),
			})

			err := monitor.Start()
			Expect(err).To(BeNil())
		})

		When("the monitor is already started", func() {

			It("should return an error", func() {
				ctx := context.Background()
				updateCh := make(chan c.CurrencyRates)
				requestCh := make(chan []string)
				errCh := make(chan error)

				monitor := NewMonitorCurrencyRateYahoo(Config{
					Ctx:                      ctx,
					UnaryAPI:                 client,
					ChanUpdateCurrencyRates:  updateCh,
					ChanRequestCurrencyRates: requestCh,
					ChanError:                errCh,
				})

				err := monitor.Start()
				Expect(err).To(BeNil())

				err = monitor.Start()
				Expect(err).To(MatchError("monitor already started"))
			})

		})

		Describe("handleRequestCurrencyRates", func() {

			When("a request for currency rates is received", func() {

				It("should return currency rates requested, and minor units, and all previous currency rates", func() {
					// Setup server to respond to EURUSD=X
					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "EURUSD=X"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuoteForCurrencyRates1Fixture),
						),
					)

					updateCh := make(chan c.CurrencyRates, 2)
					requestCh := make(chan []string, 2)
					errCh := make(chan error, 1)

					monitor := NewMonitorCurrencyRateYahoo(Config{
						Ctx:                      context.Background(),
						UnaryAPI:                 client,
						ChanUpdateCurrencyRates:  updateCh,
						ChanRequestCurrencyRates: requestCh,
						ChanError:                errCh,
					})

					err := monitor.Start()
					Expect(err).To(BeNil())

					// Request EUR
					requestCh <- []string{"EUR"}

					var rates c.CurrencyRates
					Eventually(updateCh, 500*time.Millisecond).Should(Receive(&rates))
					Expect(rates).To(HaveKey("EUR"))
					Expect(rates["EUR"].FromCurrency).To(Equal("EUR"))
					Expect(rates["EUR"].ToCurrency).To(Equal("USD"))
					Expect(rates["EUR"].Rate).To(Equal(1.1))

					Expect(rates).To(HaveKey("EUr"))
					Expect(rates["EUr"].FromCurrency).To(Equal("EUr"))
					Expect(rates["EUr"].ToCurrency).To(Equal("USD"))
					Expect(float64ValuesAreEqual(rates["EUr"].Rate, 0.011, float64EqualityTolerance)).To(BeTrue())

					// Replace API to only return quotes for GBPUSD=X
					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "GBPUSD=X"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								QuoteResponse: unary.ResponseQuoteResponse{
									Quotes: []unary.ResponseQuote{{
										Symbol:             "GBPUSD=X",
										RegularMarketPrice: unary.ResponseFieldFloat{Raw: 1.3, Fmt: "1.3"},
										Currency:           "USD",
									}},
									Error: nil,
								},
							}),
						),
					)

					// Request both EUR and GBP
					requestCh <- []string{"EUR", "GBP"}

					Eventually(updateCh, 500*time.Millisecond).Should(Receive(&rates))

					Expect(rates).To(HaveKey("EUR"))
					Expect(rates).To(HaveKey("GBP"))
					Expect(rates["EUR"].Rate).To(Equal(1.1))
					Expect(rates["GBP"].Rate).To(Equal(1.3))

					Expect(rates).To(HaveKey("EUr"))
					Expect(rates).To(HaveKey("GBp"))
					Expect(float64ValuesAreEqual(rates["EUr"].Rate, 0.011, float64EqualityTolerance)).To(BeTrue())
					Expect(float64ValuesAreEqual(rates["GBp"].Rate, 0.013, float64EqualityTolerance)).To(BeTrue())

					// Replace API to only return quotes for JPNUSD=X (no minor currency)
					server.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "JPYUSD=X"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								QuoteResponse: unary.ResponseQuoteResponse{
									Quotes: []unary.ResponseQuote{{
										Symbol:             "JPYUSD=X",
										RegularMarketPrice: unary.ResponseFieldFloat{Raw: 0.0068, Fmt: "0.0068"},
										Currency:           "USD",
									}},
									Error: nil,
								},
							}),
						),
					)

					// Request EUR, GBP & JPY
					requestCh <- []string{"EUR", "GBP", "JPY"}

					Eventually(updateCh, 500*time.Millisecond).Should(Receive(&rates))

					Expect(rates).To(HaveKey("EUR"))
					Expect(rates).To(HaveKey("GBP"))
					Expect(rates).To(HaveKey("JPY"))
					Expect(rates["EUR"].Rate).To(Equal(1.1))
					Expect(rates["GBP"].Rate).To(Equal(1.3))
					Expect(rates["JPY"].Rate).To(Equal(0.0068))

					Expect(rates).To(HaveKey("EUr"))
					Expect(rates).To(HaveKey("GBp"))
					Expect(rates).NotTo(HaveKey("JPy"))
					Expect(float64ValuesAreEqual(rates["EUr"].Rate, 0.011, float64EqualityTolerance)).To(BeTrue())
					Expect(float64ValuesAreEqual(rates["GBp"].Rate, 0.013, float64EqualityTolerance)).To(BeTrue())

				})

				When("the currency request is empty", func() {

					It("should skip this request", func() {
						updateCh := make(chan c.CurrencyRates, 1)
						requestCh := make(chan []string, 1)
						errCh := make(chan error, 1)

						monitor := NewMonitorCurrencyRateYahoo(Config{
							Ctx:                      context.Background(),
							UnaryAPI:                 client,
							ChanUpdateCurrencyRates:  updateCh,
							ChanRequestCurrencyRates: requestCh,
							ChanError:                errCh,
						})

						err := monitor.Start()
						Expect(err).To(BeNil())

						// Send empty request
						requestCh <- []string{}

						// Ensure no update or error is sent
						Consistently(updateCh, 300*time.Millisecond).ShouldNot(Receive())
						Consistently(errCh, 300*time.Millisecond).ShouldNot(Receive())
					})

				})

				When("a currency is already in the cache", func() {
					It("should not make a request for that currency", func() {
						updateCh := make(chan c.CurrencyRates, 2)
						requestCh := make(chan []string, 2)
						errCh := make(chan error, 1)

						// Setup server to respond to EURUSD=X
						server.RouteToHandler("GET", "/v7/finance/quote",
							ghttp.CombineHandlers(
								verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "EURUSD=X"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuoteForCurrencyRates1Fixture),
							),
						)

						monitor := NewMonitorCurrencyRateYahoo(Config{
							Ctx:                      context.Background(),
							UnaryAPI:                 client,
							ChanUpdateCurrencyRates:  updateCh,
							ChanRequestCurrencyRates: requestCh,
							ChanError:                errCh,
						})

						err := monitor.Start()
						Expect(err).To(BeNil())

						// First request for EUR (should hit the server)
						requestCh <- []string{"EUR"}
						var rates c.CurrencyRates
						Eventually(updateCh, 500*time.Millisecond).Should(Receive(&rates))
						Expect(rates).To(HaveKey("EUR"))
						Expect(rates["EUR"].FromCurrency).To(Equal("EUR"))
						Expect(rates["EUR"].ToCurrency).To(Equal("USD"))
						Expect(rates["EUR"].Rate).To(Equal(1.1))

						// Second request for EUR (should NOT hit the server)
						requestCh <- []string{"EUR"}
						Consistently(updateCh, 300*time.Millisecond).ShouldNot(Receive())

						// Only one HTTP request should have been made
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				When("there is an error making the request for current rates", func() {

					It("should send the error to the error channel", func() {
						// Use an invalid UnaryAPI to force an error
						invalidClient := unary.NewUnaryAPI(unary.Config{
							BaseURL:           "http://example.com/\x00",
							SessionRootURL:    server.URL(),
							SessionCrumbURL:   server.URL(),
							SessionConsentURL: server.URL(),
						})

						updateCh := make(chan c.CurrencyRates, 1)
						requestCh := make(chan []string, 1)
						errCh := make(chan error, 1)

						monitor := NewMonitorCurrencyRateYahoo(Config{
							Ctx:                      context.Background(),
							UnaryAPI:                 invalidClient,
							ChanUpdateCurrencyRates:  updateCh,
							ChanRequestCurrencyRates: requestCh,
							ChanError:                errCh,
						})

						err := monitor.Start()
						Expect(err).To(BeNil())

						// Send a request that will trigger an error
						requestCh <- []string{"EUR"}

						// Should receive an error on the error channel
						Eventually(errCh, 500*time.Millisecond).Should(Receive(MatchError(ContainSubstring("failed to create request"))))
					})

				})

			})

		})

	})

	Describe("Stop", func() {

		It("should stop the monitor", func() {
			ctx := context.Background()
			updateCh := make(chan c.CurrencyRates)
			requestCh := make(chan []string)
			errCh := make(chan error)

			monitor := NewMonitorCurrencyRateYahoo(Config{
				Ctx:                      ctx,
				UnaryAPI:                 client,
				ChanUpdateCurrencyRates:  updateCh,
				ChanRequestCurrencyRates: requestCh,
				ChanError:                errCh,
			})

			err := monitor.Start()
			Expect(err).To(BeNil())

			err = monitor.Stop()
			Expect(err).To(BeNil())
		})

		When("the monitor is not started", func() {

			It("should return an error", func() {
				ctx := context.Background()
				updateCh := make(chan c.CurrencyRates)
				requestCh := make(chan []string)
				errCh := make(chan error)

				monitor := NewMonitorCurrencyRateYahoo(Config{
					Ctx:                      ctx,
					UnaryAPI:                 client,
					ChanUpdateCurrencyRates:  updateCh,
					ChanRequestCurrencyRates: requestCh,
					ChanError:                errCh,
				})

				err := monitor.Stop()
				Expect(err).To(MatchError("monitor not started"))
			})
		})
	})

	Describe("SetTargetCurrency", func() {

		It("should set the target currency", func() {
			server.RouteToHandler("GET", "/v7/finance/quote",
				ghttp.CombineHandlers(
					verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "EURUSD=X"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuoteForCurrencyRates1Fixture),
				),
			)

			outputErrorChan := make(chan error)

			monitor := NewMonitorCurrencyRateYahoo(Config{
				Ctx:                      context.Background(),
				UnaryAPI:                 client,
				ChanUpdateCurrencyRates:  make(chan c.CurrencyRates),
				ChanRequestCurrencyRates: make(chan []string),
				ChanError:                outputErrorChan,
			})

			err := monitor.Start()
			Expect(err).To(BeNil())

			monitor.SetTargetCurrency("EURUSD=X")

			Consistently(outputErrorChan, 500*time.Millisecond).ShouldNot(Receive())

		})

		When("there is an error making the request for current rates", func() {

			It("should send the error to the error channel", func() {

				server.AppendHandlers(
					ghttp.CombineHandlers(
						verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "EURUSD=X"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuoteForCurrencyRates1Fixture),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						verifyRequest(server, "GET", "/v7/finance/quote", "symbols", "EURGBP=X"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, "invalid"),
					),
				)

				errCh := make(chan error, 2)
				requestCh := make(chan []string, 1)
				updateCh := make(chan c.CurrencyRates, 1)
				monitor := NewMonitorCurrencyRateYahoo(Config{
					Ctx:                      context.Background(),
					UnaryAPI:                 client,
					ChanUpdateCurrencyRates:  updateCh,
					ChanRequestCurrencyRates: requestCh,
					ChanError:                errCh,
				})

				err := monitor.Start()
				Expect(err).To(BeNil())

				// Send a request to set the cache
				requestCh <- []string{"EUR"}

				// Should receive an update for the first currency update
				Eventually(updateCh, 500*time.Millisecond).Should(Receive(Not(BeNil())))

				// Send a request that will trigger an error
				monitor.SetTargetCurrency("GBP")

				// Should receive an error when requesting the new currency
				Eventually(errCh, 500*time.Millisecond).Should(Receive(MatchError(ContainSubstring("failed to decode response"))))
			})

		})

	})

})

func float64ValuesAreEqual(f1 float64, f2 float64, tolerance float64) bool {
	return math.Abs(f1-f2) < tolerance
}

func verifyRequest(server *ghttp.Server, method, path string, queryKey, queryValue string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		Expect(r.Method).To(Equal(method))
		Expect(r.URL.Path).To(Equal(path))
		Expect(r.URL.Query().Get(queryKey)).To(Equal(queryValue))
	}
}

var responseQuoteForCurrencyRates1Fixture = unary.Response{
	QuoteResponse: unary.ResponseQuoteResponse{
		Quotes: []unary.ResponseQuote{
			{
				Symbol:             "EURUSD=X",
				RegularMarketPrice: unary.ResponseFieldFloat{Raw: 1.1, Fmt: "1.1"},
				Currency:           "USD",
			},
		},
		Error: nil,
	},
}
