package unary_test

import (
	"net/http"

	"github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
	. "github.com/onsi/ginkgo/v2"
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
			api := unary.NewUnaryAPI(server.URL())
			Expect(api).NotTo(BeNil())
		})
	})

	Describe("GetAssetQuotes", func() {

		It("should return a list of asset quotes", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
						Products: []unary.ResponseQuote{
							{
								Symbol:         "BTC",
								ProductID:      "BTC-USD",
								ShortName:      "Bitcoin",
								Price:          "50000.00",
								PriceChange24H: "5.00",
								Volume24H:      "1000000.00",
								MarketState:    "online",
								Currency:       "USD",
								ExchangeName:   "CBE",
							},
						},
					}),
				),
			)

			api := unary.NewUnaryAPI(server.URL())
			quotes, err := api.GetAssetQuotes([]string{"BTC-USD"})

			Expect(err).NotTo(HaveOccurred())
			Expect(quotes).To(HaveLen(1))
			Expect(quotes[0].Symbol).To(Equal("BTC"))
			Expect(quotes[0].QuotePrice.Price).To(Equal(50000.00))
		})

		When("there is a quote for a futures contract", func() {
			PIt("should return the futures product with calculated properties for the underlying asset", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "BIT-31JAN25-CDE",
									ProductID:      "BIT-31JAN25-CDE",
									ShortName:      "Bitcoin January 2025 Future",
									Price:          "60000.00",
									PriceChange24H: "5.00",
									Volume24H:      "1000000.00",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CDE",
									ProductType:    "FUTURE",
									FutureProductDetails: unary.ResponseQuoteFutureProductDetails{
										ContractRootUnit: "BTC",
									},
								},
							},
						}),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "BTC",
									ProductID:      "BTC-USD",
									ShortName:      "Bitcoin",
									Price:          "50000.00",
									PriceChange24H: "5.00",
									Volume24H:      "1000000.00",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					),
				)

				api := unary.NewUnaryAPI(server.URL())
				quotes, err := api.GetAssetQuotes([]string{"BIT-31JAN25-CDE"})

				Expect(err).NotTo(HaveOccurred())
				Expect(quotes).To(HaveLen(1))
				Expect(quotes[0].Symbol).To(Equal("BIT-31JAN25-CDE"))
				Expect(quotes[0].QuotePrice.Price).To(Equal(60000.00))
				Expect(quotes[0].QuoteFutures.SymbolUnderlying).To(Equal("BTC"))
				Expect(quotes[0].QuoteFutures.IndexPrice).To(Equal(50000.00))
				Expect(quotes[0].QuoteFutures.Basis).To(Equal(0.2))
				Expect(quotes[0].QuoteFutures.Expiry).To(Equal("31JAN25"))
			})
		})

		Context("when the request fails", func() {
			When("the request fails", func() {
				It("should return an error", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products"),
							ghttp.RespondWith(http.StatusInternalServerError, ""),
						),
					)

					api := unary.NewUnaryAPI(server.URL())
					quotes, err := api.GetAssetQuotes([]string{"BTC-USD"})

					Expect(err).To(HaveOccurred())
					Expect(quotes).To(BeEmpty())
				})
			})

			When("the response is invalid", func() {
				It("should return an error", func() {
					server.Reset()

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products"),
							ghttp.RespondWith(http.StatusOK, "invalid"),
						),
					)

					api := unary.NewUnaryAPI(server.URL())
					quotes, err := api.GetAssetQuotes([]string{"BTC-USD"})

					Expect(err).To(HaveOccurred())
					Expect(quotes).To(BeEmpty())
				})
			})

			When("the request is invalid", func() {
				It("should return an error", func() {
					api := unary.NewUnaryAPI("invalid")
					quotes, err := api.GetAssetQuotes([]string{"BTC-USD"})

					Expect(err).To(HaveOccurred())
					Expect(quotes).To(BeEmpty())
				})
			})
		})

		When("there are no symbols set", func() {
			It("should return an empty list", func() {
				api := unary.NewUnaryAPI(server.URL())
				quotes, err := api.GetAssetQuotes([]string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(quotes).To(BeEmpty())
			})
		})
	})
})
