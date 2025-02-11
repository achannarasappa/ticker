package streamer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/gorilla/websocket"
)

type messageSubscription struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type messageQuote struct {
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
	symbols          []string
	conn             *websocket.Conn
	isStarted        bool
	url              string
	assetQuoteChan   chan common.AssetQuote
	subscriptionChan chan messageSubscription
	onUpdate         func()
	wg               sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewStreamer(ctx context.Context, chanStreamUpdateQuotePrice chan messageUpdate[c.QuotePrice], chanStreamUpdateQuoteExtended chan messageUpdate[c.QuoteExtended]) *Streamer {
	ctx, cancel := context.WithCancel(ctx)

	s := &Streamer{
		ctx:              ctx,
		cancel:           cancel,
		wg:               sync.WaitGroup{},
		subscriptionChan: make(chan messageSubscription),
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

	err = s.unsubscribe()
	if err != nil {
		return err
	}

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
			var quote messageQuote
			err := s.conn.ReadJSON(&quote)
			if err != nil {
				return
			}

			// TODO: Send to correct channels
			s.assetQuoteChan <- transformQuoteStream(quote)
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
			fmt.Println("writing subscription", message)
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

func transformQuoteStream(quote messageQuote) common.AssetQuote {

	symbol := strings.Split(quote.ProductID, "-")[0]
	price, _ := strconv.ParseFloat(quote.Price, 64)

	return common.AssetQuote{
		// Name:          quote.ProductID,
		Symbol:   symbol,
		Class:    common.AssetClassCryptocurrency,
		Currency: common.Currency{FromCurrencyCode: "USD"},
		QuotePrice: common.QuotePrice{
			Price: price,
		},
		QuoteExtended: common.QuoteExtended{},
		QuoteFutures:  common.QuoteFutures{},
		QuoteSource:   common.QuoteSourceCoinbase,
		Exchange:      common.Exchange{Name: "Coinbase"},
		Meta:          common.Meta{},
	}
}
