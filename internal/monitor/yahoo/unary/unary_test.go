package unary_test

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
	. "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega/gstruct"

	"net/http"

	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Unary", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewUnaryAPI", func() {
		It("should return a new UnaryAPI", func() {
			client := unary.NewUnaryAPI(unary.Config{
				BaseURL:           server.URL(),
				SessionRootURL:    server.URL(),
				SessionCrumbURL:   server.URL(),
				SessionConsentURL: server.URL(),
			})
			Expect(client).NotTo(BeNil())
		})
	})

	Describe("GetAssetQuotes", func() {

		It("should return a list of asset quotes", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
				),
			)

			client := unary.NewUnaryAPI(unary.Config{
				BaseURL:           server.URL(),
				SessionRootURL:    server.URL(),
				SessionCrumbURL:   server.URL(),
				SessionConsentURL: server.URL(),
			})
			outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{"NET"})
			Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(84.98),
						"PricePrevClose": Equal(84.00),
						"PriceOpen":      Equal(85.22),
						"PriceDayHigh":   Equal(90.00),
						"PriceDayLow":    Equal(80.00),
						"Change":         Equal(3.0800018),
						"ChangePercent":  Equal(3.7606857),
					}),
					"QuoteSource": Equal(c.QuoteSourceYahoo),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
					"Meta": g.MatchFields(g.IgnoreExtras, g.Fields{
						"SymbolInSourceAPI": Equal("NET"),
					}),
				}),
			}))
			Expect(outputMap).To(HaveKey("NET"))
			Expect(outputError).To(BeNil())
		})

		When("no symbols are provided", func() {
			It("should return an error", func() {
				client := unary.NewUnaryAPI(unary.Config{
					BaseURL:           server.URL(),
					SessionRootURL:    server.URL(),
					SessionCrumbURL:   server.URL(),
					SessionConsentURL: server.URL(),
				})
				outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{})
				Expect(outputSlice).To(BeEmpty())
				Expect(outputMap).To(BeEmpty())
				Expect(outputError).To(MatchError("no symbols provided"))
			})
		})

		Context("when the request fails", func() {
			When("the request fails", func() {
				It("should return an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, ""),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(BeEmpty())
					Expect(outputMap).To(BeEmpty())
					Expect(outputError).To(HaveOccurred())
				})

				When("the request is invalid", func() {

					It("should return an error", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, ""),
							),
						)

						client := unary.NewUnaryAPI(unary.Config{
							BaseURL:           string([]byte("0x0D")),
							SessionRootURL:    server.URL(),
							SessionCrumbURL:   server.URL(),
							SessionConsentURL: server.URL(),
						})
						outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{"NET"})
						Expect(outputSlice).To(BeEmpty())
						Expect(outputMap).To(BeEmpty())
						Expect(outputError).To(HaveOccurred())
					})

				})

			})

			When("the response is invalid", func() {
				It("should return an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusOK, "invalid"),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(BeEmpty())
					Expect(outputMap).To(BeEmpty())
					Expect(outputError).To(HaveOccurred())
				})
			})

			When("the request is invalid", func() {
				It("should return an error", func() {
					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           "invalid",
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, outputMap, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(BeEmpty())
					Expect(outputMap).To(BeEmpty())
					Expect(outputError).To(HaveOccurred())
				})
			})
		})

		Context("market states", func() {

			When("the market is in a regular trading session", func() {

				It("should return indicate the market is active and in a regular trading session", func() {

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeTrue(),
								"IsRegularTradingSession": BeTrue(),
							}),
						}),
					}))
					Expect(outputError).To(BeNil())
				})
			})

			When("the market is in a post market session", func() {
				It("should return indicate the market is active, not in a regular trading session, and set the price to the post market price", func() {

					var inputResponse unary.Response
					inputResponse = responseQuote1Fixture
					inputResponse.QuoteResponse.Quotes[0].MarketState = "POST"

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price":         Equal(86.56),     // Post Market Price
								"Change":        Equal(4.4562718), // Post Market Change
								"ChangePercent": Equal(5.1180357), // Post Market Change Percent
							}),
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeTrue(),
								"IsRegularTradingSession": BeFalse(),
							}),
						}),
					}))
					Expect(outputError).To(BeNil())
				})

				When("and the post market price is 0", func() {
					It("should return indicate the market not in a regular trading session but not set the price", func() {

						var inputResponse unary.Response
						inputResponse = responseQuote1Fixture
						inputResponse.QuoteResponse.Quotes[0].MarketState = "POST"
						inputResponse.QuoteResponse.Quotes[0].PostMarketPrice.Raw = 0.0

						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
							),
						)

						client := unary.NewUnaryAPI(unary.Config{
							BaseURL:           server.URL(),
							SessionRootURL:    server.URL(),
							SessionCrumbURL:   server.URL(),
							SessionConsentURL: server.URL(),
						})
						outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
						Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
									"IsActive":                BeTrue(),
									"IsRegularTradingSession": BeFalse(),
								}),
								"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Price": Equal(84.98),
								}),
							}),
						}))
						Expect(outputError).To(BeNil())
					})
				})

				When("and market state is not set to a post market state but the post market price is not zero / set", func() {
					It("should return treat the asset as if it were in a post market state", func() {

						var inputResponse unary.Response
						inputResponse = responseQuote1Fixture
						inputResponse.QuoteResponse.Quotes[0].MarketState = "UNKNOWN"
						inputResponse.QuoteResponse.Quotes[0].PostMarketPrice.Raw = 84.98

						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
							),
						)

						client := unary.NewUnaryAPI(unary.Config{
							BaseURL:           server.URL(),
							SessionRootURL:    server.URL(),
							SessionCrumbURL:   server.URL(),
							SessionConsentURL: server.URL(),
						})
						outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
						Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
									"IsActive":                BeFalse(),
									"IsRegularTradingSession": BeFalse(),
								}),
								"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Price": Equal(84.98),
								}),
							}),
						}))
						Expect(outputError).To(BeNil())
					})
				})
			})

			When("the market is in a pre market session", func() {
				It("should return indicate the market is active, not in a regular trading session, and set the price to the pre market price", func() {

					var inputResponse unary.Response
					inputResponse = responseQuote1Fixture
					inputResponse.QuoteResponse.Quotes[0].MarketState = "PRE"
					inputResponse.QuoteResponse.Quotes[0].PreMarketPrice.Raw = 84.98
					inputResponse.QuoteResponse.Quotes[0].PreMarketChange.Raw = 6
					inputResponse.QuoteResponse.Quotes[0].PreMarketChangePercent.Raw = 9

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeTrue(),
								"IsRegularTradingSession": BeFalse(),
							}),
							"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price":         Equal(84.98),
								"Change":        Equal(6.0),
								"ChangePercent": Equal(9.0),
							}),
						}),
					}))
					Expect(outputError).To(BeNil())
				})

				When("and the pre market price is 0", func() {
					It("should return indicate the market not in a regular trading session but not set the price", func() {

						var inputResponse unary.Response
						inputResponse = responseQuote1Fixture
						inputResponse.QuoteResponse.Quotes[0].MarketState = "PRE"
						inputResponse.QuoteResponse.Quotes[0].PreMarketPrice.Raw = 0.0

						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
							),
						)

						client := unary.NewUnaryAPI(unary.Config{
							BaseURL:           server.URL(),
							SessionRootURL:    server.URL(),
							SessionCrumbURL:   server.URL(),
							SessionConsentURL: server.URL(),
						})
						outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
						Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
									"IsActive":                BeFalse(),
									"IsRegularTradingSession": BeFalse(),
								}),
								"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Price": Equal(84.98),
								}),
							}),
						}))
						Expect(outputError).To(BeNil())
					})
				})
			})

			When("the market state can't be determined or is not active", func() {
				It("should return indicate the market is not active and not in a regular trading session", func() {

					var inputResponse unary.Response
					inputResponse = responseQuote1Fixture
					inputResponse.QuoteResponse.Quotes[0].MarketState = "CLOSED"
					inputResponse.QuoteResponse.Quotes[0].PostMarketPrice.Raw = 0.0

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
								"IsActive":                BeFalse(),
								"IsRegularTradingSession": BeFalse(),
							}),
						}),
					}))
					Expect(outputError).To(BeNil())
				})
			})

		})

		Context("asset class", func() {
			When("the asset class is a cryptocurrency", func() {
				It("should return the asset class as a cryptocurrency", func() {

					var inputResponse unary.Response
					inputResponse = responseQuote1Fixture
					inputResponse.QuoteResponse.Quotes[0].QuoteType = "CRYPTOCURRENCY"

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.RespondWithJSONEncoded(http.StatusOK, inputResponse),
						),
					)

					client := unary.NewUnaryAPI(unary.Config{
						BaseURL:           server.URL(),
						SessionRootURL:    server.URL(),
						SessionCrumbURL:   server.URL(),
						SessionConsentURL: server.URL(),
					})
					outputSlice, _, outputError := client.GetAssetQuotes([]string{"NET"})
					Expect(outputSlice).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
						"0": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Class": Equal(c.AssetClassCryptocurrency),
						}),
					}))
					Expect(outputError).To(BeNil())
				})
			})
		})
	})
})
