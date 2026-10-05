package streamer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/onsi/gomega/gstruct"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	streamer "github.com/achannarasappa/ticker/v5/internal/monitor/coinbase/monitor-price/streamer"
	testWs "github.com/achannarasappa/ticker/v5/test/websocket"
	"github.com/gorilla/websocket"
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
			s.Stop()
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
			s.Stop()
			inputServer.Close()
		})

		It("should set the symbols and not return an error", func() {
			err := s.Start()
			Expect(err).NotTo(HaveOccurred())

			err = s.SetSymbolsAndUpdateSubscriptions([]string{"BTC-USD"}, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		When("the streamer is not started", func() {
			It("should return early without error", func() {
				err := s.SetSymbolsAndUpdateSubscriptions([]string{"BTC-USD"}, 0)
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

				s.Stop()
				inputServer.Close()
			})
		})
	})

	Describe("Stop", func() {
		It("should stop the streamer without an error", func() {
			inputServer = testWs.NewTestServer([]string{})
			s = streamer.NewStreamer(context.Background(), streamer.StreamerConfig{})
			s.SetURL("ws://" + inputServer.URL[7:])
			err := s.Start()
			Expect(err).NotTo(HaveOccurred())

			s.Stop()
			inputServer.Close()
		})
	})

	Describe("Error recovery and reconnect signaling", func() {
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

		When("the websocket connection drops", func() {
			It("should report the error, reconnect with backoff, and signal on ChanReconnect", func() {
				upgrader := websocket.Upgrader{
					CheckOrigin: func(r *http.Request) bool { return true },
				}
				var connectionCount atomic.Int32

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					conn, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						return
					}
					count := connectionCount.Add(1)
					if count == 1 {
						conn.Close()

						return
					}

					defer conn.Close()
					_ = conn.WriteMessage(websocket.TextMessage, []byte(inputTick))
					time.Sleep(500 * time.Millisecond)
				}))
				defer server.Close()

				chanErr := make(chan error, 5)
				chanReconnect := make(chan struct{}, 5)
				chanPrice := make(chan c.MessageUpdate[c.QuotePrice], 5)
				chanExt := make(chan c.MessageUpdate[c.QuoteExtended], 5)

				testStreamer := streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
					ChanStreamUpdateQuotePrice:    chanPrice,
					ChanStreamUpdateQuoteExtended: chanExt,
					ChanError:                     chanErr,
					ChanReconnect:                 chanReconnect,
					BackoffMin:                    50 * time.Millisecond,
					BackoffMax:                    500 * time.Millisecond,
				})
				defer testStreamer.Stop()

				err := testStreamer.SetURL("ws://" + server.URL[7:])
				Expect(err).NotTo(HaveOccurred())

				err = testStreamer.Start()
				Expect(err).NotTo(HaveOccurred())

				Eventually(chanErr, 2*time.Second).Should(Receive())
				Eventually(chanReconnect, 2*time.Second).Should(Receive())
				Eventually(chanPrice, 2*time.Second).Should(Receive(
					g.MatchFields(g.IgnoreExtras, g.Fields{
						"ID": Equal("ETH-USD"),
						"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price": Equal(1285.22),
						}),
					}),
				))
			})
		})

		When("ChanError is nil and a websocket error occurs", func() {
			It("should not block indefinitely and should reconnect successfully", func() {
				upgrader := websocket.Upgrader{
					CheckOrigin: func(r *http.Request) bool { return true },
				}
				var connectionCount atomic.Int32

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					conn, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						return
					}
					count := connectionCount.Add(1)
					if count == 1 {
						conn.Close()

						return
					}

					defer conn.Close()
					_ = conn.WriteMessage(websocket.TextMessage, []byte(inputTick))
					time.Sleep(500 * time.Millisecond)
				}))
				defer server.Close()

				chanReconnect := make(chan struct{}, 5)
				chanPrice := make(chan c.MessageUpdate[c.QuotePrice], 5)

				testStreamer := streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
					ChanStreamUpdateQuotePrice: chanPrice,
					ChanError:                  nil,
					ChanReconnect:              chanReconnect,
					BackoffMin:                 50 * time.Millisecond,
					BackoffMax:                 500 * time.Millisecond,
				})
				defer testStreamer.Stop()

				err := testStreamer.SetURL("ws://" + server.URL[7:])
				Expect(err).NotTo(HaveOccurred())

				err = testStreamer.Start()
				Expect(err).NotTo(HaveOccurred())

				Eventually(chanReconnect, 2*time.Second).Should(Receive())
				Eventually(chanPrice, 2*time.Second).Should(Receive(
					g.MatchFields(g.IgnoreExtras, g.Fields{
						"ID": Equal("ETH-USD"),
						"Data": g.MatchFields(g.IgnoreExtras, g.Fields{
							"Price": Equal(1285.22),
						}),
					}),
				))
			})
		})

		When("the streamer is stopped during reconnection", func() {
			It("should abort reconnect loop cleanly", func() {
				upgrader := websocket.Upgrader{
					CheckOrigin: func(r *http.Request) bool { return true },
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					conn, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						return
					}
					conn.Close()
				}))
				defer server.Close()

				chanErr := make(chan error, 10)
				testStreamer := streamer.NewStreamer(context.Background(), streamer.StreamerConfig{
					ChanError:  chanErr,
					BackoffMin: 50 * time.Millisecond,
					BackoffMax: 100 * time.Millisecond,
				})

				err := testStreamer.SetURL("ws://" + server.URL[7:])
				Expect(err).NotTo(HaveOccurred())

				err = testStreamer.Start()
				Expect(err).NotTo(HaveOccurred())

				Eventually(chanErr, 2*time.Second).Should(Receive())

				testStreamer.Stop()

				chanReconnect := make(chan struct{}, 5)
				testStreamer.SetReconnectChannel(chanReconnect)
				Consistently(chanReconnect, 200*time.Millisecond).ShouldNot(Receive())
			})
		})
	})
})
