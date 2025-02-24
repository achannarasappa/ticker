package streamer_test

import (
	"context"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/onsi/gomega/gstruct"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/streamer"
	testWs "github.com/achannarasappa/ticker/v4/test/websocket"
)

var _ = Describe("Streamer", func() {
	var (
		inputServer *httptest.Server
		s           *streamer.Streamer
	)

	Describe("NewStreamer", func() {
		It("should return a new Streamer", func() {
			s := streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
			Expect(s).NotTo(BeNil())
		})
	})

	Describe("Start", func() {
		BeforeEach(func() {
			inputServer = testWs.NewTestServer([]string{})
			s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
				ChanStreamUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice], 5),
				ChanStreamUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended], 5),
			})
			s.SetURL("ws://" + inputServer.URL[7:])
		})

		AfterEach(func() {
			inputServer.Close()
		})

		It("should start the streamer without an error", func() {
			err := s.Start()
			Expect(err).NotTo(HaveOccurred())
		})

		When("the streamer is already started", func() {
			It("should return the error 'streamer already started'", func() {
				err := s.Start()
				Expect(err).NotTo(HaveOccurred())

				err = s.Start()
				Expect(err).To(MatchError("streamer already started"))
			})
		})

		When("the url is not set", func() {
			It("should not start the streamer and not return an error", func() {
				s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
				err := s.Start()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the websocket connection is not successful", func() {
			It("should return an error containing the text 'connection aborted'", func() {
				inputServer = testWs.NewTestServer([]string{})
				s.SetURL("http://" + inputServer.URL[7:])
				err := s.Start()

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("malformed ws or wss URL")))
			})
		})

		When("the context is cancelled while trying to connect to the websocket", func() {
			It("should return an error containing the text 'connection aborted'", func() {
				ctx, cancel := context.WithCancel(context.Background())
				s = streamer.NewStreamer(ctx, streamer.StreamerConfig{})
				s.SetURL("ws://" + inputServer.URL[7:])
				started := make(chan struct{})
				var err error

				go func() {
					err = s.Start()
					close(started)
				}()
				// Relies on the assumption that it will take longer to open websocket connection than cancel the context
				cancel()
				Eventually(started).Should(BeClosed())
				Expect(err).To(MatchError(ContainSubstring("connection aborted")))
			})
		})

		Describe("readStreamQuote", func() {
			When("a tick message is received", func() {
				It("should send send a quote and extended quote to the channels", func() {
					inputTick := `{
						"type": "ticker",
						"sequence": 37475248783,
						"product_id": "ETH-USD",
						"price": "1285.22",
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
					inputServer = testWs.NewTestServer([]string{inputTick})
					outputChanStreamUpdateQuotePrice := make(chan c.MessageUpdate[c.QuotePrice], 5)
					outputChanStreamUpdateQuoteExtended := make(chan c.MessageUpdate[c.QuoteExtended], 5)

					s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
						ChanStreamUpdateQuotePrice:    outputChanStreamUpdateQuotePrice,
						ChanStreamUpdateQuoteExtended: outputChanStreamUpdateQuoteExtended,
					})
					s.SetURL("ws://" + inputServer.URL[7:])

					err := s.Start()
					Expect(err).NotTo(HaveOccurred())

					Eventually(outputChanStreamUpdateQuotePrice).Should(Receive(
						g.MatchFields(g.IgnoreExtras, g.Fields{
							"ID":       Equal("ETH-USD"),
							"Sequence": Equal(int64(37475248783)),
							"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Price": Equal(1285.22),
							}),
						}),
					))
					Eventually(outputChanStreamUpdateQuoteExtended).Should(Receive(
						g.MatchFields(g.IgnoreExtras, g.Fields{
							"ID":       Equal("ETH-USD"),
							"Sequence": Equal(int64(37475248783)),
							"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
								"Volume": Equal(245532.79269678),
							}),
						}),
					))
				})
			})

			When("a message is not a price quote or extended quote", func() {
				It("should not send anything to the channel", func() {
					invalidMessage := `{"type": "unknown"}`
					inputServer = testWs.NewTestServer([]string{invalidMessage})
					outputChanStreamUpdateQuotePrice := make(chan c.MessageUpdate[c.QuotePrice], 5)
					outputChanStreamUpdateQuoteExtended := make(chan c.MessageUpdate[c.QuoteExtended], 5)

					s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
						ChanStreamUpdateQuotePrice:    outputChanStreamUpdateQuotePrice,
						ChanStreamUpdateQuoteExtended: outputChanStreamUpdateQuoteExtended,
					})
					s.SetURL("ws://" + inputServer.URL[7:])

					err := s.Start()
					Expect(err).NotTo(HaveOccurred())

					Consistently(outputChanStreamUpdateQuotePrice, 100*time.Millisecond).ShouldNot(Receive())
					Consistently(outputChanStreamUpdateQuoteExtended, 100*time.Millisecond).ShouldNot(Receive())
				})
			})
		})
	})

	Describe("SetSymbolsAndUpdateSubscriptions", func() {
		BeforeEach(func() {
			inputServer = testWs.NewTestServer([]string{})
			s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
			s.SetURL("ws://" + inputServer.URL[7:])
		})

		AfterEach(func() {
			inputServer.Close()
		})

		It("should set the symbols and not return an error", func() {
			err := s.Start()
			Expect(err).NotTo(HaveOccurred())

			err = s.SetSymbolsAndUpdateSubscriptions([]string{"BTC-USD"})
			Expect(err).NotTo(HaveOccurred())
		})

		When("the streamer is not started", func() {
			It("should return early without error", func() {
				err := s.SetSymbolsAndUpdateSubscriptions([]string{"BTC-USD"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("SetURL", func() {
		It("should set the url and not return an error", func() {
			s := streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
			err := s.SetURL("wss://example.com")
			Expect(err).NotTo(HaveOccurred())
		})

		When("the streamer is started", func() {
			It("should return the error 'cannot set URL while streamer is connected'", func() {
				inputServer = testWs.NewTestServer([]string{})
				s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
				s.SetURL("ws://" + inputServer.URL[7:])

				err := s.Start()
				Expect(err).NotTo(HaveOccurred())

				err = s.SetURL("wss://example.com")
				Expect(err).To(MatchError("cannot set URL while streamer is connected"))

				inputServer.Close()
			})
		})
	})
})
