package unary_test

import (
	"net/http"

	"github.com/achannarasappa/ticker/v5/internal/monitor/coinpaprika/unary"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	c "github.com/achannarasappa/ticker/v5/internal/common"
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
		It("should return the asset quotes", func() {
			server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
				ghttp.RespondWithJSONEncoded(http.StatusOK, responseTickerBitcoinFixture),
			)
			server.RouteToHandler("GET", "/v1/tickers/eth-ethereum",
				ghttp.RespondWithJSONEncoded(http.StatusOK, responseTickerEthereumFixture),
			)

			api := unary.NewUnaryAPI(server.URL())
			quotes, quotesByID, err := api.GetAssetQuotes([]string{"btc-bitcoin", "eth-ethereum"})

			Expect(err).NotTo(HaveOccurred())
			Expect(quotes).To(HaveLen(2))

			Expect(quotes[0].Name).To(Equal("Bitcoin"))
			Expect(quotes[0].Symbol).To(Equal("BTC.CP"))
			Expect(quotes[0].Class).To(Equal(c.AssetClassCryptocurrency))
			Expect(quotes[0].Currency.FromCurrencyCode).To(Equal("USD"))
			Expect(quotes[0].QuotePrice.Price).To(Equal(50000.0))
			Expect(quotes[0].QuotePrice.ChangePercent).To(Equal(2.5))
			Expect(quotes[0].QuotePrice.Change).To(Equal(50000.0 * 0.025))
			Expect(quotes[0].QuoteExtended.MarketCap).To(Equal(1264515966773.0))
			Expect(quotes[0].QuoteExtended.Volume).To(Equal(21359052876.635258))
			Expect(quotes[0].QuoteSource).To(Equal(c.QuoteSourceCoinPaprika))
			Expect(quotes[0].Exchange.Name).To(Equal("CoinPaprika"))
			Expect(quotes[0].Exchange.IsActive).To(BeTrue())
			Expect(quotes[0].Exchange.IsRegularTradingSession).To(BeTrue())
			Expect(quotes[0].Meta.SymbolInSourceAPI).To(Equal("btc-bitcoin"))
			Expect(quotes[0].Meta.IsVariablePrecision).To(BeTrue())

			Expect(quotes[1].Symbol).To(Equal("ETH.CP"))
			Expect(quotes[1].QuotePrice.Price).To(Equal(4000.0))

			Expect(quotesByID).To(HaveKey("btc-bitcoin"))
			Expect(quotesByID).To(HaveKey("eth-ethereum"))
			Expect(quotesByID["btc-bitcoin"].Symbol).To(Equal("BTC.CP"))
		})

		It("should lowercase coin ids before making a request", func() {
			server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
				ghttp.RespondWithJSONEncoded(http.StatusOK, responseTickerBitcoinFixture),
			)

			api := unary.NewUnaryAPI(server.URL())
			quotes, _, err := api.GetAssetQuotes([]string{"BTC-BITCOIN"})

			Expect(err).NotTo(HaveOccurred())
			Expect(quotes).To(HaveLen(1))
			Expect(quotes[0].Symbol).To(Equal("BTC.CP"))
		})

		When("there are no coin ids", func() {
			It("should not make a request and return no quotes", func() {
				api := unary.NewUnaryAPI(server.URL())
				quotes, quotesByID, err := api.GetAssetQuotes([]string{})

				Expect(err).NotTo(HaveOccurred())
				Expect(quotes).To(BeEmpty())
				Expect(quotesByID).To(BeEmpty())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		When("a coin id is not found", func() {
			It("should skip the unknown coin id and return the other quotes", func() {
				server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseTickerBitcoinFixture),
				)
				server.RouteToHandler("GET", "/v1/tickers/not-a-real-coin",
					ghttp.RespondWithJSONEncoded(http.StatusNotFound, map[string]string{"error": "id not found"}),
				)

				api := unary.NewUnaryAPI(server.URL())
				quotes, _, err := api.GetAssetQuotes([]string{"btc-bitcoin", "not-a-real-coin"})

				Expect(err).NotTo(HaveOccurred())
				Expect(quotes).To(HaveLen(1))
				Expect(quotes[0].Symbol).To(Equal("BTC.CP"))
			})
		})

		When("the request fails", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
					ghttp.RespondWith(http.StatusInternalServerError, ""),
				)

				api := unary.NewUnaryAPI(server.URL())
				quotes, _, err := api.GetAssetQuotes([]string{"btc-bitcoin"})

				Expect(err).To(HaveOccurred())
				Expect(quotes).To(BeEmpty())
			})
		})

		When("the response is invalid JSON", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/v1/tickers/btc-bitcoin",
					ghttp.RespondWith(http.StatusOK, "not json"),
				)

				api := unary.NewUnaryAPI(server.URL())
				_, _, err := api.GetAssetQuotes([]string{"btc-bitcoin"})

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetCoinIDs", func() {
		It("should resolve the highest ranked active coin with an exactly matching symbol", func() {
			server.RouteToHandler("GET", "/v1/search/",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v1/search/", "q=BTC&c=currencies&modifier=symbol_search&limit=250"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseSearchBitcoinFixture),
				),
			)

			api := unary.NewUnaryAPI(server.URL())
			coinIds, err := api.GetCoinIDs([]string{"BTC"})

			Expect(err).NotTo(HaveOccurred())
			Expect(coinIds).To(HaveLen(1))
			Expect(coinIds["BTC"]).To(Equal("btc-bitcoin"))
		})

		When("no coin matches the symbol exactly", func() {
			It("should omit the symbol from the result", func() {
				server.RouteToHandler("GET", "/v1/search/",
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{"currencies": []map[string]interface{}{}}),
				)

				api := unary.NewUnaryAPI(server.URL())
				coinIds, err := api.GetCoinIDs([]string{"NOTACOIN"})

				Expect(err).NotTo(HaveOccurred())
				Expect(coinIds).To(BeEmpty())
			})
		})

		When("the request fails", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/v1/search/",
					ghttp.RespondWith(http.StatusInternalServerError, ""),
				)

				api := unary.NewUnaryAPI(server.URL())
				_, err := api.GetCoinIDs([]string{"BTC"})

				Expect(err).To(HaveOccurred())
			})
		})

		When("the response is invalid JSON", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/v1/search/",
					ghttp.RespondWith(http.StatusOK, "not json"),
				)

				api := unary.NewUnaryAPI(server.URL())
				_, err := api.GetCoinIDs([]string{"BTC"})

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
