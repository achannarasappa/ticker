package monitorCoinbase_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
)

var _ = Describe("Monitor Coinbase", func() {

	Describe("NewMonitorCoinbase", func() {
		It("should return a new MonitorCoinbase", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				Client: *resty.New(),
			})
			Expect(monitor).NotTo(BeNil())
		})

		When("the underlying symbols are set", func() {
			It("should set the underlying symbols", func() {
				underlyingSymbols := []string{"BTC-USD"}
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *resty.New(),
				}, monitorCoinbase.WithSymbolsUnderlying(underlyingSymbols))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the streaming URL is set", func() {
			It("should set the streaming URL", func() {
				url := "wss://websocket-feed.exchange.coinbase.com"
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *resty.New(),
				}, monitorCoinbase.WithStreamingURL(url))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the refresh interval is set", func() {
			It("should set the refresh interval", func() {
				interval := 10 * time.Second
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *resty.New(),
				}, monitorCoinbase.WithRefreshInterval(interval))

				Expect(monitor).NotTo(BeNil())
			})
		})

		When("the onUpdate function is set", func() {
			It("should set the onUpdate function", func() {
				onUpdate := func(symbol string, pq c.QuotePrice) {}
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *resty.New(),
				})
				monitor.SetOnUpdate(onUpdate)

				Expect(monitor).NotTo(BeNil())
			})
		})
	})

	Describe("GetAssetQuotes", func() {
		BeforeEach(func() {
			responseFixture := `{
				"products": [
					{
						"base_display_symbol": "BTC",
						"product_type": "SPOT",
						"product_id": "BTC-USD",
						"base_name": "Bitcoin",
						"price": "50000.00",
						"price_percentage_change_24h": "2.5",
						"volume_24h": "1000.50",
						"display_name": "Bitcoin",
						"status": "online",
						"quote_currency_id": "USD",
						"product_venue": "CBE"
					}
				]
			}`
			responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
			httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewStringResponse(200, responseFixture)
				resp.Header.Set("Content-Type", "application/json")
				return resp, nil
			})
		})

		It("should return the asset quotes", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				Client: *client,
			})

			monitor.SetSymbols([]string{"BTC-USD"})

			assetQuotes := monitor.GetAssetQuotes(true)
			Expect(assetQuotes).To(HaveLen(1))
			Expect(assetQuotes[0].Symbol).To(Equal("BTC"))
			Expect(assetQuotes[0].Name).To(Equal("Bitcoin"))
			Expect(assetQuotes[0].QuotePrice.Price).To(Equal(50000.00))
			Expect(assetQuotes[0].QuotePrice.ChangePercent).To(Equal(2.5))
		})

		When("the http request fails", func() {
			It("should return an error", func() {
				// Override the previous responder with an error response
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl,
					httpmock.NewErrorResponder(fmt.Errorf("network error")))

				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *client,
				})

				monitor.SetSymbols([]string{"BTC-USD"})

				assetQuotes := monitor.GetAssetQuotes(true)
				Expect(assetQuotes).To(BeEmpty())
			})
		})

		When("the ignoreCache flag is set to true", func() {
			It("should return the asset quotes from the cache", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *client,
				})

				monitor.SetSymbols([]string{"BTC-USD"})

				// First call to populate cache
				firstQuotes := monitor.GetAssetQuotes(true)

				// Modify the HTTP mock to return different data
				responseFixture := `{
					"products": [
						{
							"base_display_symbol": "BTC",
							"product_type": "SPOT",
							"product_id": "BTC-USD",
							"base_name": "Bitcoin",
							"price": "55000.00",
							"price_percentage_change_24h": "5.0",
							"volume_24h": "2000.50",
							"display_name": "Bitcoin",
							"status": "online",
							"quote_currency_id": "USD",
							"product_venue": "CBE"
						}
					]
				}`
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseFixture)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})

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
				Client: *client,
			}, monitorCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is already started", func() {
			It("should return an error", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *client,
				}, monitorCoinbase.WithRefreshInterval(10*time.Second))

				err := monitor.Start()
				Expect(err).NotTo(HaveOccurred())

				err = monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor already started"))
			})
		})

		When("the initial unary request for quotes fails", func() {
			BeforeEach(func() {
				responseUrl := `=~\/api\/v3\/brokerage\/market\/products.*product_ids\=BTC\-USD.*`
				httpmock.RegisterResponder("GET", responseUrl,
					httpmock.NewErrorResponder(fmt.Errorf("network error")))
			})

			It("should return an error", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *client,
				}, monitorCoinbase.WithRefreshInterval(10*time.Second))

				monitor.SetSymbols([]string{"BTC-USD"})

				err := monitor.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network error"))
			})
		})

		When("the streamer fails to start", func() {
			PIt("should return an error", func() {})
		})

		When("the poller fails to start", func() {
			PIt("should return an error", func() {})
		})

		When("there is a price update", func() {

			PIt("should call the onUpdate function", func() {})
			PIt("should update the asset quote cache", func() {})

		})

		When("there is a extended quote update", func() {

			PIt("should call the onUpdate function", func() {})
			PIt("should update the asset quote cache", func() {})

		})

	})

	Describe("Stop", func() {

		It("should stop the monitor", func() {
			monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
				Client: *client,
			}, monitorCoinbase.WithRefreshInterval(10*time.Second))

			err := monitor.Start()
			Expect(err).NotTo(HaveOccurred())

			err = monitor.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the monitor is not started", func() {
			It("should return an error", func() {
				monitor := monitorCoinbase.NewMonitorCoinbase(monitorCoinbase.Config{
					Client: *client,
				})

				err := monitor.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("monitor not started"))
			})
		})
	})
})
