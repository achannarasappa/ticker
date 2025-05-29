package monitor_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/achannarasappa/ticker/v5/internal/monitor"
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
	testWs "github.com/achannarasappa/ticker/v5/test/websocket"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Monitor", func() {

	var (
		serverCoinbase *ghttp.Server
		serverYahoo    *ghttp.Server
		wsServer       *httptest.Server
	)

	BeforeEach(func() {
		serverCoinbase = ghttp.NewServer()
		serverYahoo = ghttp.NewServer()
		wsServer = testWs.NewTestServer([]string{})
	})

	AfterEach(func() {
		serverCoinbase.Close()
		serverYahoo.Close()
		wsServer.Close()
	})

	It("should start all monitors", func() {
		// Set up mock responses for both servers
		setupCoinbaseMockHandler(serverCoinbase)
		setupYahooMockHandler(serverYahoo)

		// Create a mock websocket message for BTC price update
		wsServer = testWs.NewTestServer([]string{})

		// Create monitor with test server URLs
		m, err := monitor.NewMonitor(monitor.ConfigMonitor{
			RefreshInterval: 1,
			TargetCurrency:  "USD",
			ConfigMonitorPriceCoinbase: monitor.ConfigMonitorPriceCoinbase{
				BaseURL:      serverCoinbase.URL(),
				StreamingURL: "ws://" + wsServer.URL[7:],
			},
			ConfigMonitorsYahoo: monitor.ConfigMonitorsYahoo{
				BaseURL:           serverYahoo.URL(),
				SessionRootURL:    serverYahoo.URL(),
				SessionCrumbURL:   serverYahoo.URL(),
				SessionConsentURL: serverYahoo.URL(),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Set symbols to monitor
		m.SetAssetGroup(c.AssetGroup{
			SymbolsBySource: []c.AssetGroupSymbolsBySource{
				{
					Source:  c.QuoteSourceCoinbase,
					Symbols: []string{"BTC-USD"},
				},
				{
					Source:  c.QuoteSourceYahoo,
					Symbols: []string{"AAPL"},
				},
			},
		}, 0)

		// Start the monitor
		m.Start()

		// Verify that requests were made to both servers
		Eventually(func() int {
			return len(serverCoinbase.ReceivedRequests())
		}, 2*time.Second).Should(BeNumerically(">", 0))

		Eventually(func() int {
			return len(serverYahoo.ReceivedRequests())
		}, 2*time.Second).Should(BeNumerically(">", 0))

		// Clean up
		m.Stop()
	})

	Describe("handleUpdates", func() {

		When("an updated asset quote is sent", func() {

			It("should call the callback function", func() {

				outputCallCountSingleAsset := 0

				// Set up mock responses
				callCount := 0
				serverYahoo.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote"),
						func(w http.ResponseWriter, req *http.Request) {
							callCount++
							basePrice := 150.00
							// Increment price for each call in order to trigger an update message to be sent rather than ignoring updates do to price quote not changing
							incrementedPrice := basePrice + float64(callCount-1)*10.00

							response := unary.Response{
								QuoteResponse: unary.ResponseQuoteResponse{
									Quotes: []unary.ResponseQuote{
										{
											MarketState:                "REGULAR",
											ShortName:                  "Apple Inc.",
											PreMarketChange:            unary.ResponseFieldFloat{Raw: 1.0399933, Fmt: "1.0399933"},
											PreMarketChangePercent:     unary.ResponseFieldFloat{Raw: 1.2238094, Fmt: "1.2238094"},
											PreMarketPrice:             unary.ResponseFieldFloat{Raw: 86.03, Fmt: "86.03"},
											RegularMarketChange:        unary.ResponseFieldFloat{Raw: 3.0800018, Fmt: "3.0800018"},
											RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 3.7606857, Fmt: "3.7606857"},
											RegularMarketPrice:         unary.ResponseFieldFloat{Raw: incrementedPrice, Fmt: fmt.Sprintf("%.2f", incrementedPrice)},
											RegularMarketPreviousClose: unary.ResponseFieldFloat{Raw: 84.00, Fmt: "84.00"},
											RegularMarketOpen:          unary.ResponseFieldFloat{Raw: 85.22, Fmt: "85.22"},
											RegularMarketDayHigh:       unary.ResponseFieldFloat{Raw: 90.00, Fmt: "90.00"},
											RegularMarketDayLow:        unary.ResponseFieldFloat{Raw: 80.00, Fmt: "80.00"},
											PostMarketChange:           unary.ResponseFieldFloat{Raw: 1.37627, Fmt: "1.37627"},
											PostMarketChangePercent:    unary.ResponseFieldFloat{Raw: 1.35735, Fmt: "1.35735"},
											PostMarketPrice:            unary.ResponseFieldFloat{Raw: 86.56, Fmt: "86.56"},
											Symbol:                     "AAPL",
										},
									},
									Error: nil,
								},
							}

							json.NewEncoder(w).Encode(response)
						},
					),
				)

				// Create monitor with test server URLs
				m, err := monitor.NewMonitor(monitor.ConfigMonitor{
					RefreshInterval: 1,
					ConfigMonitorsYahoo: monitor.ConfigMonitorsYahoo{
						BaseURL:           serverYahoo.URL(),
						SessionRootURL:    serverYahoo.URL(),
						SessionCrumbURL:   serverYahoo.URL(),
						SessionConsentURL: serverYahoo.URL(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				// Set symbols to monitor
				err = m.SetAssetGroup(c.AssetGroup{
					SymbolsBySource: []c.AssetGroupSymbolsBySource{
						{
							Source:  c.QuoteSourceYahoo,
							Symbols: []string{"AAPL"},
						},
					},
				}, 0)
				Expect(err).NotTo(HaveOccurred())
				// Set the callback function
				err = m.SetOnUpdate(monitor.ConfigUpdateFns{
					OnUpdateAssetQuote: func(symbol string, assetQuote c.AssetQuote, versionVector int) {
						outputCallCountSingleAsset++
					},
					OnUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, versionVector int) {},
				})
				Expect(err).NotTo(HaveOccurred())

				// Start the monitor
				m.Start()

				// Verify callback functions were called
				Eventually(func() int {
					return outputCallCountSingleAsset
				}, 3*time.Second, 100*time.Millisecond).Should(Equal(1))

				// Clean up
				m.Stop()

			})

			When("the asset quote is outdated", func() {

				PIt("should skip calling the callback")

			})

			When("there is an error while listening for updates", func() {

				It("should send the error to logger", func() {

					logReader, logWriter, _ := os.Pipe()

					// Set up mock responses
					callCount := 0
					serverYahoo.RouteToHandler("GET", "/v7/finance/quote",
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v7/finance/quote"),
							func(w http.ResponseWriter, req *http.Request) {
								callCount++

								if callCount >= 5 {
									w.Header().Set("Location", "://bad-url")
									w.WriteHeader(http.StatusFound)
								} else {

									basePrice := 150.00
									// Increment price for each call in order to trigger an update message to be sent rather than ignoring updates do to price quote not changing
									incrementedPrice := basePrice + float64(callCount-1)*10.00

									response := unary.Response{
										QuoteResponse: unary.ResponseQuoteResponse{
											Quotes: []unary.ResponseQuote{
												{
													MarketState:                "REGULAR",
													ShortName:                  "Apple Inc.",
													PreMarketChange:            unary.ResponseFieldFloat{Raw: 1.0399933, Fmt: "1.0399933"},
													PreMarketChangePercent:     unary.ResponseFieldFloat{Raw: 1.2238094, Fmt: "1.2238094"},
													PreMarketPrice:             unary.ResponseFieldFloat{Raw: 86.03, Fmt: "86.03"},
													RegularMarketChange:        unary.ResponseFieldFloat{Raw: 3.0800018, Fmt: "3.0800018"},
													RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 3.7606857, Fmt: "3.7606857"},
													RegularMarketPrice:         unary.ResponseFieldFloat{Raw: incrementedPrice, Fmt: fmt.Sprintf("%.2f", incrementedPrice)},
													RegularMarketPreviousClose: unary.ResponseFieldFloat{Raw: 84.00, Fmt: "84.00"},
													RegularMarketOpen:          unary.ResponseFieldFloat{Raw: 85.22, Fmt: "85.22"},
													RegularMarketDayHigh:       unary.ResponseFieldFloat{Raw: 90.00, Fmt: "90.00"},
													RegularMarketDayLow:        unary.ResponseFieldFloat{Raw: 80.00, Fmt: "80.00"},
													PostMarketChange:           unary.ResponseFieldFloat{Raw: 1.37627, Fmt: "1.37627"},
													PostMarketChangePercent:    unary.ResponseFieldFloat{Raw: 1.35735, Fmt: "1.35735"},
													PostMarketPrice:            unary.ResponseFieldFloat{Raw: 86.56, Fmt: "86.56"},
													Symbol:                     "AAPL",
												},
											},
											Error: nil,
										},
									}
									json.NewEncoder(w).Encode(response)
								}
							},
						),
					)

					// Create monitor with test server URLs
					errorLogger := log.New(logWriter, "", 0)

					m, err := monitor.NewMonitor(monitor.ConfigMonitor{
						RefreshInterval: 1,
						Logger:          errorLogger,
						ConfigMonitorsYahoo: monitor.ConfigMonitorsYahoo{
							BaseURL:           serverYahoo.URL(),
							SessionRootURL:    serverYahoo.URL(),
							SessionCrumbURL:   serverYahoo.URL(),
							SessionConsentURL: serverYahoo.URL(),
						},
					})
					Expect(err).NotTo(HaveOccurred())

					// Set symbols to monitor
					m.SetAssetGroup(c.AssetGroup{
						SymbolsBySource: []c.AssetGroupSymbolsBySource{
							{
								Source:  c.QuoteSourceYahoo,
								Symbols: []string{"AAPL"},
							},
						},
					}, 0)

					// Start the monitor
					m.Start()

					// Wait for error response
					time.Sleep(2 * time.Second)

					// Close the log writer
					logWriter.Close()

					// Read the log output
					logOutput, _ := io.ReadAll(logReader)
					Expect(string(logOutput)).To(ContainSubstring("missing protocol scheme"))

					// Clean up
					m.Stop()
				})

			})

		})

	})

	Describe("SetOnUpdate", func() {

		It("should return nil when function functions are set", func() {
			m, err := monitor.NewMonitor(monitor.ConfigMonitor{
				RefreshInterval: 1,
				TargetCurrency:  "USD",
			})
			Expect(err).NotTo(HaveOccurred())

			err = m.SetOnUpdate(monitor.ConfigUpdateFns{
				OnUpdateAssetQuote:      func(symbol string, assetQuote c.AssetQuote, versionVector int) {},
				OnUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, versionVector int) {},
			})
			Expect(err).To(BeNil())
		})

		When("either asset or asset group quote update callback is not set", func() {
			It("should return an error indicating the callbacks should be set", func() {
				m, err := monitor.NewMonitor(monitor.ConfigMonitor{
					RefreshInterval: 1,
					TargetCurrency:  "USD",
				})
				Expect(err).NotTo(HaveOccurred())

				// Test with nil OnUpdateAssetQuote
				err = m.SetOnUpdate(monitor.ConfigUpdateFns{
					OnUpdateAssetQuote:      nil,
					OnUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, versionVector int) {},
				})
				Expect(err).To(MatchError("onUpdateAssetQuote and onUpdateAssetGroupQuote must be set"))

				// Test with nil OnUpdateAssetGroupQuote
				err = m.SetOnUpdate(monitor.ConfigUpdateFns{
					OnUpdateAssetQuote:      func(symbol string, assetQuote c.AssetQuote, versionVector int) {},
					OnUpdateAssetGroupQuote: nil,
				})
				Expect(err).To(MatchError("onUpdateAssetQuote and onUpdateAssetGroupQuote must be set"))
			})
		})

	})

	Describe("SetAssetGroup", func() {

		When("there is an error setting symbols for a monitor", func() {

			It("should return an error", func() {

				serverYahoo.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote"),
						func(w http.ResponseWriter, req *http.Request) {
							fields := req.URL.Query().Get("fields")
							if fields != "regularMarketPrice,currency" {
								w.Header().Set("Location", "://bad-url")
								w.WriteHeader(http.StatusFound)
								return
							}
							// Return normal response for valid fields
							json.NewEncoder(w).Encode(map[string]interface{}{
								"quoteResponse": map[string]interface{}{
									"quotes": []map[string]interface{}{
										{
											"symbol": "AAPL",
											"regularMarketPrice": map[string]interface{}{
												"raw": 150.00,
												"fmt": "150.00",
											},
											"currency": "USD",
										},
									},
									"error": nil,
								},
							})
						},
					),
				)

				setupCoinbaseMockHandler(serverCoinbase)

				m, err := monitor.NewMonitor(monitor.ConfigMonitor{
					RefreshInterval: 1,
					TargetCurrency:  "USD",
					ConfigMonitorPriceCoinbase: monitor.ConfigMonitorPriceCoinbase{
						BaseURL:      serverCoinbase.URL(),
						StreamingURL: "ws://" + wsServer.URL[7:],
					},
					ConfigMonitorsYahoo: monitor.ConfigMonitorsYahoo{
						BaseURL:           serverYahoo.URL(),
						SessionRootURL:    serverYahoo.URL(),
						SessionCrumbURL:   serverYahoo.URL(),
						SessionConsentURL: serverYahoo.URL(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				err = m.SetAssetGroup(c.AssetGroup{
					SymbolsBySource: []c.AssetGroupSymbolsBySource{
						{Source: c.QuoteSourceYahoo, Symbols: []string{"AAPL"}},
					},
				}, 0)

				Expect(err).To(MatchError(ContainSubstring("errors setting symbols")))
				Expect(err).To(MatchError(ContainSubstring("failed to make request")))
			})

		})

		When("it takes too long to set symbols on a monitor", func() {

			It("should return an error", func() {
				// Set up a slow response for Yahoo server
				serverYahoo.RouteToHandler("GET", "/v7/finance/quote",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v7/finance/quote"),
						func(w http.ResponseWriter, req *http.Request) {
							// Simulate a slow response by sleeping
							time.Sleep(3*time.Second + 100*time.Millisecond)
							json.NewEncoder(w).Encode(map[string]interface{}{
								"quoteResponse": map[string]interface{}{
									"quotes": []map[string]interface{}{
										{
											"symbol": "AAPL",
											"regularMarketPrice": map[string]interface{}{
												"raw": 150.00,
												"fmt": "150.00",
											},
											"currency": "USD",
										},
									},
									"error": nil,
								},
							})
						},
					),
				)

				setupCoinbaseMockHandler(serverCoinbase)

				m, err := monitor.NewMonitor(monitor.ConfigMonitor{
					RefreshInterval: 1,
					TargetCurrency:  "USD",
					ConfigMonitorPriceCoinbase: monitor.ConfigMonitorPriceCoinbase{
						BaseURL:      serverCoinbase.URL(),
						StreamingURL: "ws://" + wsServer.URL[7:],
					},
					ConfigMonitorsYahoo: monitor.ConfigMonitorsYahoo{
						BaseURL:           serverYahoo.URL(),
						SessionRootURL:    serverYahoo.URL(),
						SessionCrumbURL:   serverYahoo.URL(),
						SessionConsentURL: serverYahoo.URL(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				err = m.SetAssetGroup(c.AssetGroup{
					SymbolsBySource: []c.AssetGroupSymbolsBySource{
						{Source: c.QuoteSourceYahoo, Symbols: []string{"AAPL"}},
					},
				}, 0)

				Expect(err).To(MatchError(ContainSubstring("timeout waiting for monitor(s) to set symbols")))
			})

		})

	})

})

// setupCoinbaseMockHandler sets up a mock handler for Coinbase API responses
func setupCoinbaseMockHandler(server *ghttp.Server) {
	server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
				"products": []map[string]interface{}{
					{
						"symbol":           "BTC",
						"product_id":       "BTC-USD",
						"short_name":       "Bitcoin",
						"price":            "50000.00",
						"price_change_24h": "5.00",
						"volume_24h":       "1000000.00",
						"market_state":     "online",
						"currency":         "USD",
						"exchange_name":    "CBE",
						"product_type":     "SPOT",
					},
				},
			}),
		),
	)
}

// setupYahooMockHandler sets up a mock handler for Yahoo Finance API responses
func setupYahooMockHandler(server *ghttp.Server) {
	server.RouteToHandler("GET", "/v7/finance/quote",
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/v7/finance/quote"),
			ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
				"quoteResponse": map[string]interface{}{
					"quotes": []map[string]interface{}{
						{
							"symbol": "AAPL",
							"regularMarketPrice": map[string]interface{}{
								"raw": 150.00,
								"fmt": "150.00",
							},
							"regularMarketChange": map[string]interface{}{
								"raw": 2.5,
								"fmt": "+2.50",
							},
							"regularMarketChangePercent": map[string]interface{}{
								"raw": 1.67,
								"fmt": "+1.67%",
							},
							"regularMarketVolume": map[string]interface{}{
								"raw": 1000000,
								"fmt": "1M",
							},
							"marketState": "REGULAR",
							"currency":    "USD",
						},
					},
					"error": nil,
				},
			}),
		),
	)
}
