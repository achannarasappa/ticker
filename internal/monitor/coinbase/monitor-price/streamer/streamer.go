package streamer

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
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
	connCancel                    context.CancelFunc
	isStarted                     bool
	isReconnecting                atomic.Bool
	url                           string
	subscriptionChan              chan messageSubscription
	wg                            sync.WaitGroup
	ctx                           context.Context
	cancel                        context.CancelFunc
	chanStreamUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	chanStreamUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
	chanError                     chan error
	chanReconnect                 chan struct{}
	backoffMin                    time.Duration
	backoffMax                    time.Duration
	versionVector                 int
	mu                            sync.RWMutex
}

type StreamerConfig struct {
	ChanStreamUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	ChanStreamUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
	ChanError                     chan error
	ChanReconnect                 chan struct{}
	BackoffMin                    time.Duration
	BackoffMax                    time.Duration
}

func NewStreamer(ctx context.Context, config StreamerConfig) *Streamer {
	ctx, cancel := context.WithCancel(ctx)

	backoffMin := config.BackoffMin
	if backoffMin <= 0 {
		backoffMin = 100 * time.Millisecond
	}

	backoffMax := config.BackoffMax
	if backoffMax <= 0 {
		backoffMax = 30 * time.Second
	}

	s := &Streamer{
		chanStreamUpdateQuotePrice:    config.ChanStreamUpdateQuotePrice,
		chanStreamUpdateQuoteExtended: config.ChanStreamUpdateQuoteExtended,
		chanError:                     config.ChanError,
		chanReconnect:                 config.ChanReconnect,
		backoffMin:                    backoffMin,
		backoffMax:                    backoffMax,
		ctx:                           ctx,
		cancel:                        cancel,
		wg:                            sync.WaitGroup{},
		subscriptionChan:              make(chan messageSubscription, 10),
		versionVector:                 0,
	}

	return s
}

func (s *Streamer) Start() error {
	s.mu.Lock()
	if s.isStarted {
		s.mu.Unlock()

		return errors.New("streamer already started")
	}

	if s.url == "" {
		s.mu.Unlock()

		return nil
	}

	url := s.url
	s.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.DialContext(s.ctx, url, nil)
	if err != nil {
		if s.ctx.Err() != nil {
			return fmt.Errorf("connection aborted: %w", s.ctx.Err())
		}

		return err
	}

	s.mu.Lock()
	s.conn = conn
	s.isStarted = true
	connCtx, connCancel := context.WithCancel(s.ctx)
	s.connCancel = connCancel
	s.mu.Unlock()

	go func() {
		<-s.ctx.Done()
		s.mu.Lock()
		if s.connCancel != nil {
			s.connCancel()
		}
		if s.conn != nil {
			_ = s.conn.Close()
		}
		s.isStarted = false
		s.symbols = []string{}
		s.mu.Unlock()
		s.wg.Wait()
	}()

	s.wg.Add(2)
	go s.readStreamQuote(connCtx, connCancel, conn)
	go s.writeStreamSubscription(connCtx, connCancel, conn)

	return nil
}

func (s *Streamer) Stop() {
	s.cancel()
}

func (s *Streamer) SetReconnectChannel(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.chanReconnect = ch
}

func (s *Streamer) SetSymbolsAndUpdateSubscriptions(symbols []string, versionVector int) error {
	s.mu.Lock()
	s.symbols = symbols
	s.versionVector = versionVector
	started := s.isStarted
	s.mu.Unlock()

	if !started {
		return nil
	}

	return s.subscribe(symbols)
}

func (s *Streamer) SetURL(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStarted {
		return errors.New("cannot set URL while streamer is connected")
	}

	s.url = url

	return nil
}

func (s *Streamer) sendError(err error) {
	if s.chanError == nil || err == nil {
		return
	}

	select {
	case s.chanError <- err:
	default:
		go func() {
			select {
			case s.chanError <- err:
			case <-s.ctx.Done():
			}
		}()
	}
}

func (s *Streamer) signalReconnect() {
	if s.chanReconnect == nil {
		return
	}

	select {
	case s.chanReconnect <- struct{}{}:
	default:
		go func() {
			select {
			case s.chanReconnect <- struct{}{}:
			case <-s.ctx.Done():
			}
		}()
	}
}

func (s *Streamer) triggerReconnect() {
	if s.ctx.Err() != nil {
		return
	}

	if s.isReconnecting.CompareAndSwap(false, true) {
		go s.reconnect()
	}
}

func (s *Streamer) drainSubscriptions() {
	for {
		select {
		case <-s.subscriptionChan:
		default:
			return
		}
	}
}

func (s *Streamer) reconnect() {
	defer s.isReconnecting.Store(false)

	s.mu.Lock()
	if s.conn != nil {
		_ = s.conn.Close()
	}
	s.mu.Unlock()

	backoff := s.backoffMin
	timer := time.NewTimer(backoff)
	defer timer.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
		}

		if s.ctx.Err() != nil {
			return
		}

		s.mu.RLock()
		url := s.url
		s.mu.RUnlock()

		conn, _, err := websocket.DefaultDialer.DialContext(s.ctx, url, nil)
		if err != nil {
			if s.ctx.Err() != nil {
				return
			}

			s.sendError(err)
			backoff *= 2
			if backoff > s.backoffMax {
				backoff = s.backoffMax
			}
			timer.Reset(backoff)

			continue
		}

		s.mu.Lock()
		s.conn = conn
		symbols := slices.Clone(s.symbols)
		connCtx, connCancel := context.WithCancel(s.ctx)
		s.connCancel = connCancel
		s.mu.Unlock()

		s.drainSubscriptions()

		s.wg.Add(2)
		go s.readStreamQuote(connCtx, connCancel, conn)
		go s.writeStreamSubscription(connCtx, connCancel, conn)

		if len(symbols) > 0 {
			_ = s.subscribe(symbols)
		}

		s.signalReconnect()

		return
	}
}

func (s *Streamer) readStreamQuote(connCtx context.Context, connCancel context.CancelFunc, conn *websocket.Conn) {
	defer s.wg.Done()

	for {
		select {
		case <-connCtx.Done():
			return
		default:
			var message messagePriceTick
			err := conn.ReadJSON(&message)
			if err != nil {
				if s.ctx.Err() != nil {
					return
				}

				s.sendError(err)
				connCancel()
				s.triggerReconnect()

				return
			}

			// Only handle ticker messages; first message is a subscription confirmation
			if message.Type != "ticker" {
				continue
			}

			s.mu.RLock()
			versionVector := s.versionVector
			s.mu.RUnlock()

			qp, qe := transformPriceTick(message, versionVector)
			if s.chanStreamUpdateQuotePrice != nil {
				select {
				case s.chanStreamUpdateQuotePrice <- qp:
				case <-connCtx.Done():
					return
				}
			}

			if s.chanStreamUpdateQuoteExtended != nil {
				select {
				case s.chanStreamUpdateQuoteExtended <- qe:
				case <-connCtx.Done():
					return
				}
			}
		}
	}
}

func (s *Streamer) writeStreamSubscription(connCtx context.Context, connCancel context.CancelFunc, conn *websocket.Conn) {
	defer s.wg.Done()

	for {
		select {
		case <-connCtx.Done():
			return
		case message := <-s.subscriptionChan:
			err := conn.WriteJSON(message)
			if err != nil {
				if s.ctx.Err() != nil {
					return
				}

				s.sendError(err)
				connCancel()
				s.triggerReconnect()

				return
			}
		}
	}
}

func (s *Streamer) subscribe(productIDs []string) error {
	message := messageSubscription{
		Type:       "subscribe",
		ProductIDs: productIDs,
		Channels:   []string{"ticker"},
	}

	select {
	case s.subscriptionChan <- message:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *Streamer) unsubscribe() error { //nolint:unused
	message := messageSubscription{
		Type:     "unsubscribe",
		Channels: []string{"ticker"},
	}

	select {
	case s.subscriptionChan <- message:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func transformPriceTick(message messagePriceTick, versionVector int) (qp c.MessageUpdate[c.QuotePrice], qe c.MessageUpdate[c.QuoteExtended]) {
	price, _ := strconv.ParseFloat(message.Price, 64)
	priceOpen, _ := strconv.ParseFloat(message.Open24h, 64)
	priceDayHigh, _ := strconv.ParseFloat(message.High24h, 64)
	priceDayLow, _ := strconv.ParseFloat(message.Low24h, 64)
	change := price - priceOpen
	changePercent := change / priceOpen

	qp = c.MessageUpdate[c.QuotePrice]{
		ID:            message.ProductID,
		Sequence:      message.Sequence,
		VersionVector: versionVector,
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
		ID:            message.ProductID,
		Sequence:      message.Sequence,
		VersionVector: versionVector,
		Data: c.QuoteExtended{
			Volume: volume,
		},
	}

	return qp, qe
}
