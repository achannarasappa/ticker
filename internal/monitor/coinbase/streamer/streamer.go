package streamer

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/gorilla/websocket"
)

type messageSubscription struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type messagePriceTick struct {
	Type        string `json:"type"`
	Sequence    int64  `json:"sequence"`
	ProductID   string `json:"product_id"`
	Price       string `json:"price"`
	Open24h     string `json:"open_24h"`
	Volume24h   string `json:"volume_24h"`
	Low24h      string `json:"low_24h"`
	High24h     string `json:"high_24h"`
	Volume30d   string `json:"volume_30d"`
	BestBid     string `json:"best_bid"`
	BestBidSize string `json:"best_bid_size"`
	BestAsk     string `json:"best_ask"`
	BestAskSize string `json:"best_ask_size"`
	Side        string `json:"side"`
	Time        string `json:"time"`
	TradeID     int64  `json:"trade_id"`
	LastSize    string `json:"last_size"`
}

type Streamer struct {
	symbols                       []string
	conn                          *websocket.Conn
	isStarted                     bool
	url                           string
	assetQuoteChan                chan c.AssetQuote
	subscriptionChan              chan messageSubscription
	onUpdate                      func()
	wg                            sync.WaitGroup
	ctx                           context.Context
	cancel                        context.CancelFunc
	chanStreamUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	chanStreamUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
}

type StreamerConfig struct {
	ChanStreamUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	ChanStreamUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
}

func NewStreamer(ctx context.Context, config StreamerConfig) *Streamer {
	ctx, cancel := context.WithCancel(ctx)

	s := &Streamer{
		chanStreamUpdateQuotePrice:    config.ChanStreamUpdateQuotePrice,
		chanStreamUpdateQuoteExtended: config.ChanStreamUpdateQuoteExtended,
		ctx:                           ctx,
		cancel:                        cancel,
		wg:                            sync.WaitGroup{},
		subscriptionChan:              make(chan messageSubscription),
	}

	return s
}

func (s *Streamer) Start() error {
	if s.isStarted {
		return fmt.Errorf("streamer already started")
	}

	if s.url == "" {
		// TODO: log streaming not started
		return nil
	}

	// Create connection channel for result
	connChan := make(chan *websocket.Conn)
	errChan := make(chan error)

	// Connect the websocket address in a goroutine
	go func() {
		url := s.url
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	// Wait for either connection, error, or context cancellation
	select {
	case conn := <-connChan:
		s.conn = conn
	case err := <-errChan:
		return err
	case <-s.ctx.Done():
		return fmt.Errorf("connection aborted: %w", s.ctx.Err())
	}

	// Disconnect on stop signal
	go func() {
		<-s.ctx.Done()
		s.wg.Wait()
		s.conn.Close()
		s.isStarted = false
		s.symbols = []string{}
	}()

	s.isStarted = true

	s.wg.Add(2)
	go s.readStreamQuote()
	go s.writeStreamSubscription()

	return nil
}

func (s *Streamer) SetSymbolsAndUpdateSubscriptions(symbols []string) error {

	var err error

	if !s.isStarted {
		return nil
	}

	s.symbols = symbols

	// err = s.unsubscribe()
	// if err != nil {
	// 	return err
	// }

	err = s.subscribe(s.symbols)
	if err != nil {
		return err
	}

	return nil
}

func (s *Streamer) SetURL(url string) error {

	if s.isStarted {
		return fmt.Errorf("cannot set URL while streamer is connected")
	}

	s.url = "wss://" + url
	return nil
}

func (s *Streamer) readStreamQuote() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			var message messagePriceTick
			err := s.conn.ReadJSON(&message)
			if err != nil {
				return
			}

			// Only handle ticker messages; first message is a subscription confirmation
			if message.Type != "ticker" {
				continue
			}

			qp, qe := transformPriceTick(message)
			s.chanStreamUpdateQuotePrice <- qp
			s.chanStreamUpdateQuoteExtended <- qe
		}
	}
}

func (s *Streamer) writeStreamSubscription() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case message := <-s.subscriptionChan:

			err := s.conn.WriteJSON(message)
			if err != nil {
				return
			}
		default:
		}
	}
}

func (s *Streamer) subscribe(productIDs []string) error {

	message := messageSubscription{
		Type:       "subscribe",
		ProductIDs: productIDs,
		Channels:   []string{"ticker"},
	}

	s.subscriptionChan <- message
	return nil
}

func (s *Streamer) unsubscribe() error {

	message := messageSubscription{
		Type:     "unsubscribe",
		Channels: []string{"ticker"},
	}

	s.subscriptionChan <- message
	return nil
}

func transformPriceTick(message messagePriceTick) (qp c.MessageUpdate[c.QuotePrice], qe c.MessageUpdate[c.QuoteExtended]) {

	price, _ := strconv.ParseFloat(message.Price, 64)
	priceOpen, _ := strconv.ParseFloat(message.Open24h, 64)
	priceDayHigh, _ := strconv.ParseFloat(message.High24h, 64)
	priceDayLow, _ := strconv.ParseFloat(message.Low24h, 64)
	change := price - priceOpen
	changePercent := change / priceOpen

	qp = c.MessageUpdate[c.QuotePrice]{
		ID:       message.ProductID,
		Sequence: message.Sequence,
		Data: c.QuotePrice{
			Price:          price,
			PricePrevClose: priceOpen,
			PriceOpen:      priceOpen,
			PriceDayHigh:   priceDayHigh,
			PriceDayLow:    priceDayLow,
			Change:         change,
			ChangePercent:  changePercent,
		},
	}

	volume, _ := strconv.ParseFloat(message.Volume24h, 64)

	qe = c.MessageUpdate[c.QuoteExtended]{
		ID:       message.ProductID,
		Sequence: message.Sequence,
		Data: c.QuoteExtended{
			Volume: volume,
		},
	}

	return qp, qe
}
