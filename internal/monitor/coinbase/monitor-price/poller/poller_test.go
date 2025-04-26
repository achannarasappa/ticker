package poller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	g "github.com/onsi/gomega/gstruct"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/monitor-price/poller"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
)

var _ = Describe("Poller", func() {
	var (
		server *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

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
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewPoller", func() {
		It("should create a new poller instance", func() {
			p := poller.NewPoller(context.Background(), poller.PollerConfig{
				UnaryAPI:             unary.NewUnaryAPI(server.URL()),
				ChanUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote], 5),
			})
			Expect(p).NotTo(BeNil())
		})
	})

	Describe("Start", func() {
		It("should start polling for price updates", func() {

			inputChanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 5)

			p := poller.NewPoller(context.Background(), poller.PollerConfig{
				UnaryAPI:             unary.NewUnaryAPI(server.URL()),
				ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
			})
			p.SetSymbols([]string{"BTC-USD"}, 0)
			p.SetRefreshInterval(time.Millisecond * 250)

			err := p.Start()
			Expect(err).NotTo(HaveOccurred())

			Eventually(inputChanUpdateAssetQuote).Should(Receive(
				g.MatchFields(g.IgnoreExtras, g.Fields{
					"ID": Equal("BTC-USD"),
					"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price": Equal(50000.00),
						}),
					}),
				}),
			))
		})

		When("the poller is already started", func() {
			When("and the poller is started again", func() {
				It("should return an error", func() {
					p := poller.NewPoller(context.Background(), poller.PollerConfig{
						UnaryAPI:             unary.NewUnaryAPI(server.URL()),
						ChanUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote], 5),
					})
					p.SetSymbols([]string{"BTC-USD"}, 0)
					p.SetRefreshInterval(time.Second * 1)

					err := p.Start()
					Expect(err).NotTo(HaveOccurred())
					err = p.Start()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("poller already started"))
				})

			})

			When("and the refresh interval is set again", func() {
				It("should return an error", func() {
					p := poller.NewPoller(context.Background(), poller.PollerConfig{
						UnaryAPI:             unary.NewUnaryAPI(server.URL()),
						ChanUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote], 5),
					})
					p.SetSymbols([]string{"BTC-USD"}, 0)
					p.SetRefreshInterval(time.Second * 1)

					err := p.Start()
					Expect(err).NotTo(HaveOccurred())
					err = p.SetRefreshInterval(time.Second * 1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("cannot set refresh interval while poller is started"))
				})

			})

		})

		When("the refresh interval is not set", func() {
			It("should return an error", func() {
				p := poller.NewPoller(context.Background(), poller.PollerConfig{
					UnaryAPI:             unary.NewUnaryAPI(server.URL()),
					ChanUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote], 5),
				})
				p.SetSymbols([]string{"BTC-USD"}, 0)

				err := p.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("refresh interval is not set"))
			})
		})

		When("the symbols are not set", func() {
			It("should not return any price updates", func() {

				inputChanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 5)

				p := poller.NewPoller(context.Background(), poller.PollerConfig{
					UnaryAPI:             unary.NewUnaryAPI(server.URL()),
					ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
				})
				p.SetRefreshInterval(time.Millisecond * 100)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				Consistently(inputChanUpdateAssetQuote).ShouldNot(Receive())
			})
		})

		When("the unary API returns an error", func() {
			It("should return an error", func() {

				requestCount := 0

				server.RouteToHandler("GET", "/api/v3/brokerage/market/products",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v3/brokerage/market/products", "product_ids=BTC-USD"),
						func(w http.ResponseWriter, r *http.Request) {
							requestCount++
							if requestCount > 3 {
								w.WriteHeader(http.StatusInternalServerError)
							} else {
								w.WriteHeader(http.StatusOK)
								json.NewEncoder(w).Encode(unary.Response{
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
								})
							}
						},
					),
				)

				outputChanError := make(chan error, 1)

				p := poller.NewPoller(context.Background(), poller.PollerConfig{
					UnaryAPI:             unary.NewUnaryAPI(server.URL()),
					ChanUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote], 5),
					ChanError:            outputChanError,
				})
				p.SetSymbols([]string{"BTC-USD"}, 0)
				p.SetRefreshInterval(time.Millisecond * 20)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				Eventually(outputChanError).Should(Receive(
					MatchError("request failed with status 500"),
				))

			})
		})

		When("the context is cancelled", func() {
			It("should stop the polling process", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				outputChanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 5)
				outputChanError := make(chan error, 5)

				p := poller.NewPoller(ctx, poller.PollerConfig{
					UnaryAPI:             unary.NewUnaryAPI(server.URL()),
					ChanUpdateAssetQuote: outputChanUpdateAssetQuote,
					ChanError:            outputChanError,
				})
				p.SetSymbols([]string{"BTC-USD"}, 0)
				p.SetRefreshInterval(time.Millisecond * 20)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				cancel()

				Consistently(outputChanUpdateAssetQuote).ShouldNot(Receive())
				Consistently(outputChanError).ShouldNot(Receive())
			})
		})

	})

	Describe("Stop", func() {
		It("should stop the polling process", func() {
			// Test implementation will go here
		})
	})
})
