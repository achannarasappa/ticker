package monitorPriceCoinbase_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	monitorPriceCoinbase "github.com/achannarasappa/ticker/v5/internal/monitor/coinbase/monitor-price"
	unary "github.com/achannarasappa/ticker/v5/internal/monitor/coinbase/unary"
	testWs "github.com/achannarasappa/ticker/v5/test/websocket"
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

	Describe("NewMonitorPriceCoinbase", func() {
		It("should return a new MonitorCoinbase", func() {
			monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
				UnaryURL:                 server.URL(),
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			})
			Expect(monitor).NotTo(BeNil())
		})

		When("the streaming URL is set", func() {
			It("should set the streaming URL", func() {
				url := "wss://websocket-feed.exchange.coinbase.com"
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithStreamingURL(url))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the refresh interval is set", func() {
			It("should set the refresh interval", func() {
				interval := 10 * time.Second
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(interval))

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
				// Second call within SetSymbols gets quotes for all symbols
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

			monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
				UnaryURL:                 server.URL(),
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			})

			monitor.SetSymbols([]string{"BIT-31JAN25-CDE", "ETH-USD"}, 0)

			assetQuotes, err := monitor.GetAssetQuotes()
			Expect(err).NotTo(HaveOccurred())

			Expect(assetQuotes).To(HaveLen(2))
			Expect(assetQuotes[0].Symbol).To(Equal("BIT-31JAN25-CDE.CB"))
			Expect(assetQuotes[0].Name).To(Equal("Bitcoin January 2025 Future"))
			Expect(assetQuotes[0].Class).To(Equal(c.AssetClassFuturesContract))
			Expect(assetQuotes[1].Symbol).To(Equal("ETH.CB"))
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

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				monitor.SetSymbols([]string{"BTC-USD"}, 0)

				assetQuotes, err := monitor.GetAssetQuotes(true)
				Expect(err).To(HaveOccurred())
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

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				monitor.SetSymbols([]string{"BTC-USD"}, 0)

				// First call to populate cache
				firstQuotes, err := monitor.GetAssetQuotes(true)
				Expect(err).NotTo(HaveOccurred())
				// Second call with ignoreCache=false should return cached data
				secondQuotes, err := monitor.GetAssetQuotes(false)
				Expect(err).NotTo(HaveOccurred())

				Expect(secondQuotes).To(HaveLen(1))
				Expect(secondQuotes[0].QuotePrice.Price).To(Equal(firstQuotes[0].QuotePrice.Price))
				Expect(secondQuotes[0].QuotePrice.ChangePercent).To(Equal(firstQuotes[0].QuotePrice.ChangePercent))
			})
		})
	})

	Describe("SetSymbols", func() {
		When("there is a derivatives product (i.e. has an underlying asset)", func() {
			When("a symbol is already mapped to an underlying symbol", func() {
				It("should skip adding that symbol to the mapping between product ids and underlying product ids", func() {
					// Initial response for first SetSymbols call
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
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
									},
								},
							}),
						),
						// Response for getting all quotes after mapping is established
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE&product_ids=BTC-USD"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
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
									},
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

					// Verify that no additional API calls are made to get underlying symbols
					// when setting the same symbol again
					server.AppendHandlers(
						// Only expect the call to get all quotes, not the underlying mapping call
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE"),
							ghttp.RespondWithJSONEncoded(http.StatusOK, unary.Response{
								Products: []unary.ResponseQuote{
									{
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
									},
								},
							}),
						),
					)

					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					})

					// First call to SetSymbols establishes the mapping
					err := monitor.SetSymbols([]string{"BIT-31JAN25-CDE"}, 0)
					Expect(err).NotTo(HaveOccurred())

					// Second call to SetSymbols should skip getting underlying symbols
					err = monitor.SetSymbols([]string{"BIT-31JAN25-CDE"}, 1)
					Expect(err).NotTo(HaveOccurred())

					// Verify that all server handlers were called as expected
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})

			When("there is an error making the request to retrieve the underlying product ids", func() {
				It("should return an error", func() {
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BIT-31JAN25-CDE"),
							ghttp.RespondWith(http.StatusInternalServerError, "network error"),
						),
					)

					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					})

					err := monitor.SetSymbols([]string{"BIT-31JAN25-CDE"}, 0)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("request failed with status 500"))
				})
			})

		})
	})

	Describe("Start", func() {
		It("should start the monitor", func() {
			monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
				UnaryURL:                 server.URL(),
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second))

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

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second))

				monitor.SetSymbols([]string{"BTC-USD"}, 0)

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

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second),
					monitorPriceCoinbase.WithStreamingURL("wssss://invalid-url"))

				monitor.SetSymbols([]string{"ETH-USD"}, 0)

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

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				monitor.SetSymbols([]string{"ETH-USD"}, 0)

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("refresh interval is not set"))
			})
		})

		When("there is a streaming price update", func() {
			It("should send the updated price quote to the channel", func() {
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

				// Create a channel to receive updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					ChanUpdateAssetQuote:     updateChan,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second),
					monitorPriceCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

				monitor.SetSymbols([]string{"ETH-USD"}, 0)
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
				}, 5*time.Second).Should(Equal(1300.00))

				Expect(receivedQuote.Symbol).To(Equal("ETH.CB"))
			})

			When("the price has not changed", func() {
				It("should not send updates to the channel", func() {
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

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond),
						monitorPriceCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

					monitor.Start()
					monitor.SetSymbols([]string{"ETH-USD"}, 0)

					Consistently(func() bool {
						select {
						case <-updateChan:
							return true
						default:
							return false
						}
					}, 1*time.Second).Should(BeFalse())

					monitor.Stop()
				})
			})

			When("the product does not exist in the cache", func() {
				It("should not send updates to the channel", func() {
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

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond),
						monitorPriceCoinbase.WithStreamingURL("ws://"+inputServer.URL[7:]))

					monitor.Start()
					monitor.SetSymbols([]string{"ETH-USD"}, 0) // Only set ETH-USD in cache

					Consistently(func() bool {
						select {
						case <-updateChan:
							return true
						default:
							return false
						}
					}, 1*time.Second).Should(BeFalse())

					monitor.Stop()
				})
			})
		})

		When("there is a polling asset update", func() {
			It("should send the updated asset quote to the channel", func() {

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

				// Create a channel to receive updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				// Create a monitor with a short refresh interval for testing
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					ChanUpdateAssetQuote:     updateChan,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond))

				monitor.SetSymbols([]string{"BIT-31JAN25-CDE", "BTC-USD"}, 0)
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
				}, 1*time.Second).Should(Equal(75000.00))

				Expect(receivedQuote.Symbol).To(Equal("BIT-31JAN25-CDE.CB"))

				monitor.Stop()
			})

			When("the price has not changed", func() {
				It("should not send updates to the channel", func() {
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

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

					// Create a monitor with a short refresh interval for testing
					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond))

					monitor.SetSymbols([]string{"BTC-USD"}, 0)
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

					// Create a channel to receive updates
					updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)
					receivedSymbols := make(map[string]bool)

					// Create a monitor with a short refresh interval
					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						ChanUpdateAssetQuote:     updateChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond))

					monitor.SetSymbols([]string{"ETH-USD"}, 0) // Only set ETH-USD in cache
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

					// Check for updates and record which symbols we receive
					go func() {
						for update := range updateChan {
							receivedSymbols[update.Data.Symbol] = true
						}
					}()

					Consistently(func() bool {
						_, hasBTC := receivedSymbols["BTC.CB"]
						return hasBTC
					}, 500*time.Millisecond).Should(BeFalse())

					monitor.Stop()
				})
			})
		})

		When("there is a currency rate update", func() {
			It("should replace the currency rate cache", func() {
				var err error

				// Set up initial server response for asset quotes
				server.RouteToHandler("GET", "/api/v3/brokerage/market/products", func(w http.ResponseWriter, r *http.Request) {
					query := r.URL.Query()["product_ids"]

					if len(query) == 1 && query[0] == "BIT-31JAN25-CDE" {
						json.NewEncoder(w).Encode(unary.Response{
							Products: []unary.ResponseQuote{
								{
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
								},
							},
						})
						return
					}
					if len(query) == 2 && ((query[0] == "BTC-USD" && query[1] == "BIT-31JAN25-CDE") || (query[0] == "BIT-31JAN25-CDE" && query[1] == "BTC-USD")) {
						json.NewEncoder(w).Encode(unary.Response{
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
								{
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
								},
							},
						})
						return
					}
					w.WriteHeader(http.StatusNotFound)
				})

				// Create channels for currency rate updates and asset quote updates
				updateChan := make(chan c.MessageUpdate[c.AssetQuote], 10)

				// Create and start the monitor
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					ChanUpdateAssetQuote:     updateChan,
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond))

				monitor.SetSymbols([]string{"BIT-31JAN25-CDE"}, 0)
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

				Expect(quotes[0].Currency.Rate).To(Equal(0.85))
				Expect(quotes[0].Currency.FromCurrencyCode).To(Equal("USD"))
				Expect(quotes[0].Currency.ToCurrencyCode).To(Equal("EUR"))

				monitor.Stop()
			})

			When("there is an error getting asset quotes and replacing the cache after new currency rates are recieved", func() {
				It("should send a message to the error channel", func() {
					// Set up initial server response for asset quotes
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

					// Create channels for updates and errors
					errorChan := make(chan error, 1)

					// Create and start the monitor
					monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
						UnaryURL:                 server.URL(),
						ChanError:                errorChan,
						Ctx:                      context.Background(),
						ChanRequestCurrencyRates: make(chan []string, 1),
					}, monitorPriceCoinbase.WithRefreshInterval(100*time.Millisecond))

					monitor.SetSymbols([]string{"BTC-USD"}, 0)
					monitor.Start()

					// Set up error response for the subsequent asset quote request
					server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
							ghttp.RespondWith(http.StatusInternalServerError, "Internal Server Error"),
						),
					)

					// Send currency rates update to trigger asset quote refresh
					currencyRates := c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         0.85,
						},
					}

					err := monitor.SetCurrencyRates(currencyRates)
					Expect(err).To(HaveOccurred())

					Expect(err.Error()).To(ContainSubstring("request failed with status 500"))

					monitor.Stop()
				})
			})

		})

	})

	Describe("Stop", func() {
		It("should stop the monitor", func() {
			monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
				UnaryURL:                 server.URL(),
				Ctx:                      context.Background(),
				ChanRequestCurrencyRates: make(chan []string, 1),
			}, monitorPriceCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is not started", func() {
			It("should return an error", func() {
				monitor := monitorPriceCoinbase.NewMonitorPriceCoinbase(monitorPriceCoinbase.Config{
					UnaryURL:                 server.URL(),
					Ctx:                      context.Background(),
					ChanRequestCurrencyRates: make(chan []string, 1),
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})
})
