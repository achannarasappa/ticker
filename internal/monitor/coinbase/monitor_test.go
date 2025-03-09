package monitorCoinbase_test

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
	testWs "github.com/achannarasappa/ticker/v4/test/websocket"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Monitor Coinbase", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewMonitorCoinbase", func() {
		It("should return a new MonitorCoinbase", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				UnaryURL: server.URL(),
			})
			Expect(monitor).NotTo(BeNil())
		})

		When("the streaming URL is set", func() {
			It("should set the streaming URL", func() {
				url := "wss://websocket-feed.exchange.coinbase.com"
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithStreamingURL(url))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the refresh interval is set", func() {
			It("should set the refresh interval", func() {
				interval := 10 * time.Second
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(interval))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the onUpdate function is set", func() {
			It("should set the onUpdate function", func() {
				onUpdate := func(symbol string, assetQuote c.AssetQuote) {}
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				})
				monitor.SetOnUpdateAssetQuote(onUpdate)

				Expect(monitor).NotTo(BeNil())
			})
		})
	})

	Describe("GetAssetQuotes", func() {
		It("should return the asset quotes", func() {

			quoteFutures := unary.ResponseQuote{
				Symbol:         "BIT-31JAN25-CDE",
				ProductID:      "BIT-31JAN25-CDE",
				ShortName:      "Bitcoin Futures",
				Price:          "60000.00",
				PriceChange24H: "5.00",
				Volume24H:      "1000000.00",
				MarketState:    "online",
				Currency:       "USD",
				ExchangeName:   "CDE",
				ProductType:    "FUTURE",
				FutureProductDetails: unary.ResponseQuoteFutureProductDetails{
					ContractRootUnit:   "BTC",
					GroupDescription:   "Bitcoin January 2025 Future",
					ExpirationDate:     "2025-01-31",
					ExpirationTimezone: "America/New_York",
				},
			}
			quoteSpotBTC := unary.ResponseQuote{
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
			}
			quoteSpotETH := unary.ResponseQuote{
				Symbol:         "ETH",
				ProductID:      "ETH-USD",
				ShortName:      "Ethereum",
				Price:          "1000.00",
				PriceChange24H: "10.00",
				Volume24H:      "10000.00",
				MarketState:    "online",
				Currency:       "USD",
				ExchangeName:   "CBE",
				ProductType:    "SPOT",
			}

			server.AppendHandlers(
				// First call within SetSymbols checks for symbols with underlying assets
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
						Products: []unary.ResponseQuote{
							quoteFutures,
						},
					}),
				),
				// Second call within SetSymbolsgets quotes for all symbols
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE&product_ids=BTC-USD&product_ids=ETH-USD"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
						Products: []unary.ResponseQuote{
							quoteFutures,
							quoteSpotBTC,
							quoteSpotETH,
						},
					}),
				),
			)

			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				UnaryURL: server.URL(),
			})

			monitor.SetSymbols([]string{"BIT-31JAN25-CDE", "ETH-USD"})
			assetQuotes := monitor.GetAssetQuotes()

			Expect(assetQuotes).To(HaveLen(2))
			Expect(assetQuotes[0].Symbol).To(Equal("BIT-31JAN25-CDE"))
			Expect(assetQuotes[0].Name).To(Equal("Bitcoin January 2025 Future"))
			Expect(assetQuotes[0].Class).To(Equal(c.AssetClassFuturesContract))
			Expect(assetQuotes[1].Symbol).To(Equal("ETH"))
			Expect(assetQuotes[1].Name).To(Equal("Ethereum"))
			Expect(assetQuotes[1].Class).To(Equal(c.AssetClassCryptocurrency))
		})

		When("the http request fails", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						ghttp.RespondWith(http.StatusInternalServerError, ""),
					),
				)

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				})

				monitor.SetSymbols([]string{"BTC-USD"})

				assetQuotes := monitor.GetAssetQuotes(true)
				Expect(assetQuotes).To(BeEmpty())
			})
		})

		When("the ignoreCache flag is set to true", func() {
			It("should return the asset quotes from the cache", func() {
				baseResponse := func() http.HandlerFunc {
					return ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "BTC",
									ProductID:      "BTC-USD",
									ShortName:      "Bitcoin",
									Price:          "50000.00",
									PriceChange24H: "2.5",
									Volume24H:      "1000.50",
									DisplayName:    "Bitcoin",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					)
				}

				// First response
				server.AppendHandlers(
					baseResponse(),
					baseResponse(),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "BTC",
									ProductID:      "BTC-USD",
									ShortName:      "Bitcoin",
									Price:          "55000.00",
									PriceChange24H: "5.0",
									Volume24H:      "1000.50",
									DisplayName:    "Bitcoin",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					),
				)

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				})

				monitor.SetSymbols([]string{"BTC-USD"})

				// First call to populate cache
				firstQuotes := monitor.GetAssetQuotes(true)
				// Second call with ignoreCache=false should return cached data
				secondQuotes := monitor.GetAssetQuotes(false)

				Expect(secondQuotes).To(HaveLen(1))
				Expect(secondQuotes[0].QuotePrice.Price).To(Equal(firstQuotes[0].QuotePrice.Price))
				Expect(secondQuotes[0].QuotePrice.ChangePercent).To(Equal(firstQuotes[0].QuotePrice.ChangePercent))
			})
		})
	})

	Describe("Start", func() {
		It("should start the monitor", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				UnaryURL: server.URL(),
			}, monitorCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(10*time.Second))

				err := monitor.Start()
				Expect(err).NotTo(HaveOccurred())

				err = monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor already started"))
			})
		})

		When("the initial unary request for quotes fails", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						ghttp.RespondWith(http.StatusInternalServerError, "network error"),
					),
				)

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(10*time.Second))

				monitor.SetSymbols([]string{"BTC-USD"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("request failed with status 500"))
			})
		})

		When("the streamer fails to start", func() {
			It("should return an error", func() {
				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "ETH",
									ProductID:      "ETH-USD",
									ShortName:      "Ethereum",
									Price:          "1285.22",
									PriceChange24H: "2.5",
									Volume24H:      "245532.79",
									DisplayName:    "Ethereum",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					),
				)

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(10*time.Second),
					monitorCoinbase.WithStreamingURL("wssss://invalid-url"))

				monitor.SetSymbols([]string{"ETH-USD"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("malformed ws or wss URL"))
			})
		})

		When("the poller fails to start", func() {
			It("should return an error", func() {

				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "ETH",
									ProductID:      "ETH-USD",
									ShortName:      "Ethereum",
									Price:          "1285.22",
									PriceChange24H: "2.5",
									Volume24H:      "245532.79",
									DisplayName:    "Ethereum",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					),
				)

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				})

				monitor.SetSymbols([]string{"ETH-USD"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("refresh interval is not set"))
			})
		})

		When("there is a streaming price update", func() {
			It("should call the onUpdate function with the updated price quote", func() {
				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
							Products: []unary.ResponseQuote{
								{
									Symbol:         "ETH",
									ProductID:      "ETH-USD",
									ShortName:      "Ethereum",
									Price:          "1285.22",
									PriceChange24H: "2.5",
									Volume24H:      "245532.79",
									DisplayName:    "Ethereum",
									MarketState:    "online",
									Currency:       "USD",
									ExchangeName:   "CBE",
									ProductType:    "SPOT",
								},
							},
						}),
					),
				)

				inputTick := `{
					"type": "ticker",
					"sequence": 37475248783,
					"product_id": "ETH-USD",
					"price": "1300.00",
					"open_24h": "1310.79",
					"volume_24h": "245532.79269678",
					"low_24h": "1280.52",
					"high_24h": "1313.8",
					"volume_30d": "9788783.60117027",
					"best_bid": "1285.04",
					"best_bid_size": "0.46688654",
					"best_ask": "1285.27",
					"best_ask_size": "1.56637040",
					"side": "buy",
					"time": "2022-10-19T23:28:22.061769Z",
					"trade_id": 370843401,
					"last_size": "11.4396987"
				}`
				inputServer := testWs.NewTestServer([]string{inputTick})

				outputCalled := false
				outputAssetQuote := c.AssetQuote{}
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(10*time.Second),
					monitorCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

				monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
					outputCalled = true
					outputAssetQuote = assetQuote
				})

				monitor.SetSymbols([]string{"ETH-USD"})
				monitor.Start()

				Eventually(func() bool {
					return outputCalled
				}, 5*time.Second).Should(BeTrue())

				Expect(outputAssetQuote.QuotePrice.Price).To(Equal(1300.00))

			})

			When("the price has not changed", func() {
				It("should not call the onUpdate function", func() {
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "ETH",
										ProductID:      "ETH-USD",
										ShortName:      "Ethereum",
										Price:          "1300.00",
										PriceChange24H: "2.5",
										Volume24H:      "245532.79",
										DisplayName:    "Ethereum",
										MarketState:    "online",
										Currency:       "USD",
										ExchangeName:   "CBE",
										ProductType:    "SPOT",
									},
								},
							}),
						),
					)

					inputTick := `{
						"type": "ticker",
						"product_id": "ETH-USD",
						"price": "1300.00",
						"time": "2022-10-19T23:28:22.061769Z"
					}`
					inputServer := testWs.NewTestServer([]string{inputTick, inputTick})

					outputCalled := false
					monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
						UnaryURL: server.URL(),
					}, monitorCoinbase.WithRefreshInterval(10*time.Second),
						monitorCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						outputCalled = true
					})

					monitor.Start()
					monitor.SetSymbols([]string{"ETH-USD"})

					Consistently(func() bool {
						return outputCalled
					}, 2*time.Second).Should(BeFalse())
				})
			})
			When("the product does not exist in the cache", func() {
				It("should not call the onUpdate function", func() {
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "ETH",
										ProductID:      "ETH-USD",
										ShortName:      "Ethereum",
										Price:          "1300.00",
										PriceChange24H: "2.5",
										Volume24H:      "245532.79",
										DisplayName:    "Ethereum",
										MarketState:    "online",
										Currency:       "USD",
										ExchangeName:   "CBE",
										ProductType:    "SPOT",
									},
								},
							}),
						),
					)

					// Send a ticker update for BTC-USD which is not in the cache
					inputTick := `{
						"type": "ticker",
						"product_id": "BTC-USD",
						"price": "50000.00",
						"time": "2022-10-19T23:28:22.061769Z"
					}`
					inputServer := testWs.NewTestServer([]string{inputTick})

					outputCalled := false
					monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
						UnaryURL: server.URL(),
					}, monitorCoinbase.WithRefreshInterval(10*time.Second),
						monitorCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						outputCalled = true
					})

					monitor.Start()
					monitor.SetSymbols([]string{"ETH-USD"}) // Only set ETH-USD in cache

					Consistently(func() bool {
						return outputCalled
					}, 2*time.Second).Should(BeFalse())
				})
			})
		})

		When("there is a polling asset update", func() {
			It("should call the onUpdate function with the updated asset quote", func() {

				quoteFutures := unary.ResponseQuote{
					Symbol:         "BIT-31JAN25-CDE",
					ProductID:      "BIT-31JAN25-CDE",
					ShortName:      "Bitcoin Futures",
					Price:          "60000.00",
					PriceChange24H: "5.00",
					Volume24H:      "1000000.00",
					MarketState:    "online",
					Currency:       "USD",
					ExchangeName:   "CDE",
					ProductType:    "FUTURE",
					FutureProductDetails: unary.ResponseQuoteFutureProductDetails{
						ContractRootUnit:   "BTC",
						GroupDescription:   "Bitcoin January 2025 Future",
						ExpirationDate:     "2025-01-31",
						ExpirationTimezone: "America/New_York",
					},
				}
				quoteSpotBTC := unary.ResponseQuote{
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
				}

				calledCount := 0

				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						func(w http.ResponseWriter, r *http.Request) {

							if calledCount > 3 {
								quoteFutures.Price = "75000.00"
							}

							calledCount++
							query := r.URL.Query()

							if query["product_ids"][0] == "BTC-USD,BIT-31JAN25-CDE" || query["product_ids"][0] == "BIT-31JAN25-CDE,BTC-USD" {
								json.NewEncoder(w).Encode(unary.Response{
									Products: []unary.ResponseQuote{
										quoteFutures,
										quoteSpotBTC,
									},
								})
							}

							if r.URL.Query().Get("product_ids") == "BIT-31JAN25-CDE" {
								json.NewEncoder(w).Encode(unary.Response{
									Products: []unary.ResponseQuote{
										quoteFutures,
									},
								})
							}

							if r.URL.Query().Get("product_ids") == "BTC-USD" {
								json.NewEncoder(w).Encode(unary.Response{
									Products: []unary.ResponseQuote{
										quoteSpotBTC,
									},
								})
							}

						},
					),
				)

				var outputAssetQuote c.AssetQuote
				var outputCalled bool

				// Create a monitor with a short refresh interval for testing
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				}, monitorCoinbase.WithRefreshInterval(500*time.Millisecond))

				monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
					outputAssetQuote = assetQuote
					outputCalled = true
				})

				monitor.SetSymbols([]string{"BIT-31JAN25-CDE"})
				monitor.Start()

				// Wait for the update to be called
				Eventually(func() bool {
					return outputCalled
				}, 2*time.Second).Should(BeTrue())
				Expect(outputAssetQuote.QuotePrice.Price).To(Equal(75000.00))
				Expect(outputAssetQuote.Symbol).To(Equal("BIT-31JAN25-CDE"))

				monitor.Stop()
			})

			When("the price has not changed", func() {
				It("should not call the onUpdate function", func() {
					// Initial response to set up the cache
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "BTC",
										ProductID:      "BTC-USD",
										ShortName:      "Bitcoin",
										Price:          "50000.00",
										PriceChange24H: "2.5",
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

					// Set up a channel to detect if onUpdate is called
					outputCalled := false

					// Create a monitor with a short refresh interval for testing
					monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
						UnaryURL: server.URL(),
					}, monitorCoinbase.WithRefreshInterval(100*time.Millisecond))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						outputCalled = true
					})

					monitor.SetSymbols([]string{"BTC-USD"})
					monitor.Start()

					// Set up the second response with the same price
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "BTC",
										ProductID:      "BTC-USD",
										ShortName:      "Bitcoin",
										Price:          "50000.00", // Same price
										PriceChange24H: "2.5",      // Same change
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

					// Wait a bit to ensure the polling happens
					Consistently(func() bool {
						return outputCalled
					}, 500*time.Millisecond).Should(BeFalse())

					monitor.Stop()
				})
			})

			When("the product does not exist in the cache", func() {
				It("should not call the onUpdate function", func() {
					// Initial response to set up the cache with ETH-USD
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "ETH",
										ProductID:      "ETH-USD",
										ShortName:      "Ethereum",
										Price:          "3000.00",
										PriceChange24H: "1.5",
										Volume24H:      "500000.00",
										MarketState:    "online",
										Currency:       "USD",
										ExchangeName:   "CBE",
										ProductType:    "SPOT",
									},
								},
							}),
						),
					)

					// Set up a channel to detect if onUpdate is called
					outputCalled := false

					// Create a monitor with a short refresh interval for testing
					monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
						UnaryURL: server.URL(),
					}, monitorCoinbase.WithRefreshInterval(100*time.Millisecond))

					monitor.SetOnUpdateAssetQuote(func(symbol string, assetQuote c.AssetQuote) {
						// Only mark as called if we get an update for BTC-USD
						if symbol == "BTC" {
							outputCalled = true
						}
					})

					monitor.SetSymbols([]string{"ETH-USD"}) // Only set ETH-USD in cache
					monitor.Start()

					// Set up the second response with BTC-USD which is not in the cache
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=ETH-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
										Symbol:         "ETH",
										ProductID:      "ETH-USD",
										ShortName:      "Ethereum",
										Price:          "3000.00",
										PriceChange24H: "1.5",
										Volume24H:      "500000.00",
										MarketState:    "online",
										Currency:       "USD",
										ExchangeName:   "CBE",
										ProductType:    "SPOT",
									},
									{
										Symbol:         "BTC", // This is not in the cache
										ProductID:      "BTC-USD",
										ShortName:      "Bitcoin",
										Price:          "50000.00",
										PriceChange24H: "2.5",
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

					// Wait a bit to ensure the polling happens
					Consistently(func() bool {
						return outputCalled
					}, 500*time.Millisecond).Should(BeFalse())

					monitor.Stop()
				})
			})
		})

		When("there is a extended quote update", func() {
			PIt("should call the onUpdate function", func() {})
			PIt("should update the asset quote cache", func() {})
		})
	})

	Describe("Stop", func() {
		It("should stop the monitor", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				UnaryURL: server.URL(),
			}, monitorCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is not started", func() {
			It("should return an error", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					UnaryURL: server.URL(),
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})
})
