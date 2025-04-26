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
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/poller"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
)

var _ = Describe("Poller", func() {
	var (
		server                    *ghttp.Server
		ctx                       context.Context
		cancel                    context.CancelFunc
		inputUnaryAPI             *unary.UnaryAPI
		inputChanUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
		inputChanError            chan error
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		server.RouteToHandler("GET", "/v7/finance/quote",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v7/finance/quote", "symbols=NET&fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&formatted=true&lang=en-US&region=US&corsDomain=finance.yahoo.com"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, responseQuote1Fixture),
			),
		)

		inputChanUpdateAssetQuote = make(chan c.MessageUpdate[c.AssetQuote], 5)
		inputChanError = make(chan error, 5)
		ctx, cancel = context.WithCancel(context.Background())

		inputUnaryAPI = unary.NewUnaryAPI(unary.Config{
			BaseURL:           server.URL(),
			SessionRootURL:    server.URL(),
			SessionCrumbURL:   server.URL(),
			SessionConsentURL: server.URL(),
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewPoller", func() {
		It("should create a new poller instance", func() {
			p := poller.NewPoller(context.Background(), poller.PollerConfig{
				UnaryAPI:             inputUnaryAPI,
				ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
				ChanError:            inputChanError,
			})
			Expect(p).NotTo(BeNil())
		})
	})

	Describe("Start", func() {
		It("should start polling for price updates", func() {

			p := poller.NewPoller(ctx, poller.PollerConfig{
				UnaryAPI:             inputUnaryAPI,
				ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
				ChanError:            inputChanError,
			})

			p.SetSymbols([]string{"NET"}, 0)
			p.SetRefreshInterval(time.Millisecond * 100)

			err := p.Start()
			Expect(err).NotTo(HaveOccurred())

			Eventually(inputChanUpdateAssetQuote).Should(Receive(
				g.MatchFields(g.IgnoreExtras, g.Fields{
					"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
						"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price": Equal(84.98),
						}),
					}),
				}),
			))
			Eventually(inputChanError).ShouldNot(Receive())

		})

		When("the poller is already started", func() {
			When("and the poller is started again", func() {
				It("should return an error", func() {

					p := poller.NewPoller(ctx, poller.PollerConfig{
						UnaryAPI:             inputUnaryAPI,
						ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
						ChanError:            inputChanError,
					})

					p.SetSymbols([]string{"NET"}, 0)
					p.SetRefreshInterval(time.Millisecond * 100)

					err := p.Start()
					Expect(err).NotTo(HaveOccurred())

					err = p.Start()
					Expect(err).To(HaveOccurred())
				})
			})

			When("and the refresh interval is set again", func() {
				It("should return an error", func() {

					p := poller.NewPoller(ctx, poller.PollerConfig{
						UnaryAPI:             inputUnaryAPI,
						ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
						ChanError:            inputChanError,
					})

					p.SetSymbols([]string{"NET"}, 0)
					p.SetRefreshInterval(time.Millisecond * 100)

					err := p.Start()
					Expect(err).NotTo(HaveOccurred())

					err = p.SetRefreshInterval(time.Millisecond * 200)
					Expect(err).To(HaveOccurred())

				})
			})
		})

		When("the refresh interval is not set", func() {
			It("should return an error", func() {
				p := poller.NewPoller(ctx, poller.PollerConfig{
					UnaryAPI:             inputUnaryAPI,
					ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
					ChanError:            inputChanError,
				})

				p.SetSymbols([]string{"NET"}, 0)

				err := p.Start()
				Expect(err).To(HaveOccurred())
			})
		})

		When("the symbols are not set", func() {
			It("should not return any price updates", func() {

				p := poller.NewPoller(ctx, poller.PollerConfig{
					UnaryAPI:             inputUnaryAPI,
					ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
					ChanError:            inputChanError,
				})

				p.SetRefreshInterval(time.Millisecond * 100)
				p.SetSymbols([]string{}, 0)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				Consistently(inputChanUpdateAssetQuote).ShouldNot(Receive())
				Consistently(inputChanError).ShouldNot(Receive())

			})
		})

		When("the context is cancelled", func() {
			It("should stop the polling process", func() {
				p := poller.NewPoller(ctx, poller.PollerConfig{
					UnaryAPI:             inputUnaryAPI,
					ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
					ChanError:            inputChanError,
				})

				p.SetSymbols([]string{"NET"}, 0)
				p.SetRefreshInterval(time.Millisecond * 100)

				err := p.Start()
				Expect(err).NotTo(HaveOccurred())

				cancel()

				Consistently(inputChanUpdateAssetQuote).ShouldNot(Receive())
				Consistently(inputChanError).ShouldNot(Receive())

				Eventually(ctx.Done()).Should(BeClosed())

			})
		})
	})

	Describe("Stop", func() {
		It("should stop the polling process", func() {

			p := poller.NewPoller(ctx, poller.PollerConfig{
				UnaryAPI:             inputUnaryAPI,
				ChanUpdateAssetQuote: inputChanUpdateAssetQuote,
				ChanError:            inputChanError,
			})

			p.SetSymbols([]string{"NET"}, 0)
			p.SetRefreshInterval(time.Millisecond * 100)

			err := p.Start()
			Expect(err).NotTo(HaveOccurred())

			err = p.Stop()
			Expect(err).NotTo(HaveOccurred())

			Consistently(inputChanUpdateAssetQuote).ShouldNot(Receive())
			Consistently(inputChanError).ShouldNot(Receive())
		})
	})
})
