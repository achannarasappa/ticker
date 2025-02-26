package poller_test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	g "github.com/onsi/gomega/gstruct"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/poller"
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
				UnaryAPI:                unary.NewUnaryAPI(server.URL()),
				ChanUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
				ChanUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
			})
			Expect(p).NotTo(BeNil())
		})
	})

	Describe("Start", func() {
		It("should start polling for price updates", func() {

			inputChanUpdateQuotePrice := make(chan c.MessageUpdate[c.QuotePrice], 5)
			inputChanUpdateQuoteExtended := make(chan c.MessageUpdate[c.QuoteExtended], 5)

			p := poller.NewPoller(context.Background(), poller.PollerConfig{
				UnaryAPI:                unary.NewUnaryAPI(server.URL()),
				ChanUpdateQuotePrice:    inputChanUpdateQuotePrice,
				ChanUpdateQuoteExtended: inputChanUpdateQuoteExtended,
			})
			p.SetSymbols([]string{"BTC-USD"})
			p.SetRefreshInterval(time.Millisecond * 250)

			err := p.Start()
			Expect(err).NotTo(HaveOccurred())

			Eventually(inputChanUpdateQuotePrice).Should(Receive(
				g.MatchFields(g.IgnoreExtras, g.Fields{
					"ID": Equal("BTC-USD"),
					"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price": Equal(50000.00),
					}),
				}),
			))
			Eventually(inputChanUpdateQuoteExtended).Should(Receive(
				g.MatchFields(g.IgnoreExtras, g.Fields{
					"ID": Equal("BTC-USD"),
					"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Volume": Equal(1000000.00),
					}),
				}),
			))
		})

		When("the poller is already started", func() {
			When("and the poller is started again", func() {
				It("should return an error", func() {
					p := poller.NewPoller(context.Background(), poller.PollerConfig{
						UnaryAPI:                unary.NewUnaryAPI(server.URL()),
						ChanUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
						ChanUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
					})
					p.SetSymbols([]string{"BTC-USD"})
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
						UnaryAPI:                unary.NewUnaryAPI(server.URL()),
						ChanUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
						ChanUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
					})
					p.SetSymbols([]string{"BTC-USD"})
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
					UnaryAPI:                unary.NewUnaryAPI(server.URL()),
					ChanUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
					ChanUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
				})
				p.SetSymbols([]string{"BTC-USD"})

				err := p.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("refresh interval is not set"))
			})
		})

		When("the symbols are not set", func() {
			It("should not return any price updates", func() {

				inputChanUpdateQuotePrice := make(chan c.MessageUpdate[c.QuotePrice], 5)
				inputChanUpdateQuoteExtended := make(chan c.MessageUpdate[c.QuoteExtended], 5)

				p := poller.NewPoller(context.Background(), poller.PollerConfig{
					UnaryAPI:                unary.NewUnaryAPI(server.URL()),
					ChanUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
					ChanUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
				})
				p.SetRefreshInterval(time.Millisecond * 100)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				Consistently(inputChanUpdateQuotePrice).ShouldNot(Receive())
				Consistently(inputChanUpdateQuoteExtended).ShouldNot(Receive())
			})
		})

	})

	Describe("Stop", func() {
		It("should stop the polling process", func() {
			// Test implementation will go here
		})
	})
})
