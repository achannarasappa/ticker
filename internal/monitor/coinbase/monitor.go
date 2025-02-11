package monitorCoinbase

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/poller"
	streamer "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/streamer"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
	resty "github.com/go-resty/resty/v2"
)

type MonitorCoinbase struct {
	unaryAPI                      *unary.UnaryAPI
	streamer                      *streamer.Streamer
	poller                        *poller.Poller
	unary                         *unary.UnaryAPI
	input                         input
	productIds                    []string // Coinbase APIs refer to trading pairs as Product IDs which symbols ticker accepts with a -USD suffix
	productIdsStreaming           []string
	productIdsPolling             []string
	assetQuotesResponse           []c.AssetQuote           // Asset quotes filtered to the productIds set in input.productIds
	assetQuotesCache              map[string]*c.AssetQuote // Asset quotes for all assets retrieved at least once
	chanStreamUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	chanStreamUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
	chanStreamUpdateExchange      chan c.MessageUpdate[c.Exchange]
	chanPollUpdateAssetQuote      chan c.MessageUpdate[c.AssetQuote]
	mu                            sync.RWMutex
	ctx                           context.Context
	cancel                        context.CancelFunc
	isStarted                     bool
}

type input struct {
	productIds        []string
	symbolsUnderlying []string
}

// Config contains the required configuration for the Coinbase monitor
type Config struct {
	Client   resty.Client
	OnUpdate func()
}

// Option defines an option for configuring the monitor
type Option func(*MonitorCoinbase)

func NewMonitorCoinbase(config Config, opts ...Option) *MonitorCoinbase {

	ctx, cancel := context.WithCancel(context.Background())

	unaryAPI := unary.NewUnaryAPI(config.Client)

	monitor := &MonitorCoinbase{
		assetQuotesCache:              make(map[string]*c.AssetQuote),
		assetQuotesResponse:           make([]c.AssetQuote, 0),
		chanStreamUpdateQuotePrice:    make(chan c.MessageUpdate[c.QuotePrice]),
		chanStreamUpdateQuoteExtended: make(chan c.MessageUpdate[c.QuoteExtended]),
		chanStreamUpdateExchange:      make(chan c.MessageUpdate[c.Exchange]),
		chanPollUpdateAssetQuote:      make(chan c.MessageUpdate[c.AssetQuote]),
		poller:                        poller.NewPoller(ctx, unaryAPI),
		unaryAPI:                      unaryAPI,
		ctx:                           ctx,
		cancel:                        cancel,
	}

	streamerConfig := streamer.StreamerConfig{
		ChanStreamUpdateQuotePrice:    monitor.chanStreamUpdateQuotePrice,
		ChanStreamUpdateQuoteExtended: monitor.chanStreamUpdateQuoteExtended,
	}

	monitor.streamer = streamer.NewStreamer(ctx, streamerConfig)

	for _, opt := range opts {
		opt(monitor)
	}

	return monitor
}

// WithSymbolsUnderlying sets the underlying symbols for the monitor
func WithSymbolsUnderlying(symbols []string) Option {
	return func(m *MonitorCoinbase) {
		m.input.symbolsUnderlying = symbols
	}
}

// WithStreamingURL sets the streaming URL for the monitor
func WithStreamingURL(url string) Option {
	return func(m *MonitorCoinbase) {
		m.streamer.SetURL(url)
	}
}

// WithRefreshInterval sets the refresh interval for the monitor
func WithRefreshInterval(interval time.Duration) Option {
	return func(m *MonitorCoinbase) {
		m.poller.SetRefreshInterval(interval)
	}
}

func (m *MonitorCoinbase) GetAssetQuotes(ignoreCache ...bool) []c.AssetQuote {
	if len(ignoreCache) > 0 && ignoreCache[0] {
		return m.unaryAPI.GetAssetQuotes(m.productIds)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	assetQuotes := make([]c.AssetQuote, 0, len(m.productIds))
	for _, productId := range m.productIds {
		if quote, exists := m.assetQuotesCache[productId]; exists {
			assetQuotes = append(assetQuotes, *quote)
		}
	}

	return assetQuotes
}

func (m *MonitorCoinbase) SetSymbols(productIds []string) {

	// Underlying symbols may also be explicitly set so merge and deduplicate
	productIdsUnique := mergeAndDeduplicateProductIds(m.input.symbolsUnderlying, productIds)

	m.productIdsStreaming, m.productIdsPolling = partitionProductIds(productIdsUnique)

	m.input.productIds = productIds
	m.productIds = productIdsUnique

	// Lock updates to asset quotes while symbols are changed and subscriptions updates. ensure data from unary call supercedes potentially oudated streaming data
	m.mu.Lock()
	defer m.mu.Unlock()

	// Execute one unary API call to get data not sent by streaming API and set initial prices
	m.assetQuotesResponse = m.unaryAPI.GetAssetQuotes(m.productIds) // TODO: update to return and handle error
	m.updateAssetQuotesCache(m.assetQuotesResponse)

	// Coinbase steaming API for CBE (spot) only and not CDE (futures)
	m.streamer.SetSymbolsAndUpdateSubscriptions(m.productIdsStreaming) // TODO: update to return and handle error

}

// Start the monitor
func (m *MonitorCoinbase) Start() error {

	var err error

	if m.isStarted {
		return fmt.Errorf("monitor already started")
	}

	// On start, get initial quotes from unary API
	m.assetQuotesResponse = m.unaryAPI.GetAssetQuotes(m.productIds)
	m.updateAssetQuotesCache(m.assetQuotesResponse)

	err = m.streamer.Start()
	if err != nil {
		return err
	}

	err = m.poller.Start()
	if err != nil {
		return err
	}

	go m.handleUpdates()

	m.isStarted = true

	return nil
}

func (m *MonitorCoinbase) Stop() error {

	if !m.isStarted {
		return fmt.Errorf("monitor not started")
	}

	m.cancel()
	return nil
}

func (m *MonitorCoinbase) updateAssetQuotesCache(assetQuotes []c.AssetQuote) {
	for _, quote := range assetQuotes {
		// quote.Symbol is the base symbol so index by trading pair / product ID
		m.assetQuotesCache[quote.Meta.SymbolInSourceAPI] = &quote
	}
}

func isStreamingProductId(productId string) bool {
	return !strings.HasSuffix(productId, "-CDE") && !strings.HasPrefix(productId, "CDE")
}

func partitionProductIds(productIds []string) (productIdsStreaming []string, productIdsPolling []string) {
	productIdsStreaming = make([]string, 0)
	productIdsPolling = make([]string, 0)

	for _, productId := range productIds {
		if isStreamingProductId(productId) {
			productIdsStreaming = append(productIdsStreaming, productId)
		} else {
			productIdsPolling = append(productIdsPolling, productId)
		}
	}

	return productIdsStreaming, productIdsPolling
}

func mergeAndDeduplicateProductIds(symbolsA, symbolsB []string) []string {
	merged := make([]string, 0, len(symbolsA)+len(symbolsB))
	merged = append(merged, symbolsA...)
	merged = append(merged, symbolsB...)
	slices.Sort(merged)
	return slices.Compact(merged)
}

func (m *MonitorCoinbase) handleUpdates() {
	for {
		select {
		case updateMessage := <-m.chanStreamUpdateQuotePrice:

			var assetQuote *c.AssetQuote
			var exists bool

			// Check if cache exists and values have changed before acquiring write lock
			m.mu.RLock()
			defer m.mu.RUnlock()

			assetQuote, exists = m.assetQuotesCache[updateMessage.ID]

			if !exists {
				// If product id does not exist in cache, skip update
				// TODO: log product not found in cache - should not happen
				continue
			}

			// Skip update if price has not changed
			if assetQuote.QuotePrice.Price == updateMessage.Data.Price {
				continue
			}
			m.mu.RUnlock()

			// Price is different so update cache
			m.mu.Lock()
			defer m.mu.Unlock()

			assetQuote.QuotePrice.Price = updateMessage.Data.Price
			assetQuote.QuotePrice.Change = updateMessage.Data.Change
			assetQuote.QuotePrice.ChangePercent = updateMessage.Data.ChangePercent
			assetQuote.QuotePrice.PriceDayHigh = updateMessage.Data.PriceDayHigh
			assetQuote.QuotePrice.PriceDayLow = updateMessage.Data.PriceDayLow
			assetQuote.QuotePrice.PriceOpen = updateMessage.Data.PriceOpen
			assetQuote.QuotePrice.PricePrevClose = updateMessage.Data.PricePrevClose

		case <-m.ctx.Done():
			return
		default:
		}
	}
}
