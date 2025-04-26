package monitorPriceCoinbase

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/monitor-price/poller"
	streamer "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/monitor-price/streamer"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
)

type MonitorCoinbase struct {
	unaryAPI                         *unary.UnaryAPI
	streamer                         *streamer.Streamer
	poller                           *poller.Poller
	unary                            *unary.UnaryAPI
	input                            input
	productIds                       []string // Coinbase APIs refer to trading pairs as Product IDs which symbols ticker accepts with a -USD suffix
	productIdsStreaming              []string
	productIdsPolling                []string
	productIdsToUnderlyingProductIds map[string]string        // Map of productIds to underlying productIds
	productIdsUnderlyingOnly         map[string]bool          // Product IDs which are only underlying assets of explicit requested assets (e.g. BTC-USD is an underlying asset of BIT-31JAN25-CDE)
	assetQuotesCache                 []c.AssetQuote           // Asset quotes for all assets retrieved at start or on symbol change
	assetQuotesCacheLookup           map[string]*c.AssetQuote // Asset quotes for all assets retrieved at least once (symbol change does not remove symbols)
	chanStreamUpdateQuotePrice       chan c.MessageUpdate[c.QuotePrice]
	chanStreamUpdateQuoteExtended    chan c.MessageUpdate[c.QuoteExtended]
	chanStreamUpdateExchange         chan c.MessageUpdate[c.Exchange]
	chanPollUpdateAssetQuote         chan c.MessageUpdate[c.AssetQuote]
	chanError                        chan error
	mu                               sync.RWMutex
	ctx                              context.Context
	cancel                           context.CancelFunc
	isStarted                        bool
	chanUpdateAssetQuote             chan c.MessageUpdate[c.AssetQuote]
}

type input struct {
	productIds       []string
	productIdsLookup map[string]bool
}

// Config contains the required configuration for the Coinbase monitor
type Config struct {
	Ctx                  context.Context
	UnaryURL             string
	ChanError            chan error
	ChanUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
}

// Option defines an option for configuring the monitor
type Option func(*MonitorCoinbase)

func NewMonitorCoinbase(config Config, opts ...Option) *MonitorCoinbase {

	ctx, cancel := context.WithCancel(config.Ctx)

	unaryAPI := unary.NewUnaryAPI(config.UnaryURL)

	monitor := &MonitorCoinbase{
		assetQuotesCacheLookup:           make(map[string]*c.AssetQuote),
		assetQuotesCache:                 make([]c.AssetQuote, 0),
		productIdsToUnderlyingProductIds: make(map[string]string),
		chanStreamUpdateQuotePrice:       make(chan c.MessageUpdate[c.QuotePrice]),
		chanStreamUpdateQuoteExtended:    make(chan c.MessageUpdate[c.QuoteExtended]),
		chanStreamUpdateExchange:         make(chan c.MessageUpdate[c.Exchange]),
		chanPollUpdateAssetQuote:         make(chan c.MessageUpdate[c.AssetQuote]),
		chanError:                        config.ChanError,
		unaryAPI:                         unaryAPI,
		ctx:                              ctx,
		cancel:                           cancel,
		chanUpdateAssetQuote:             config.ChanUpdateAssetQuote,
	}

	pollerConfig := poller.PollerConfig{
		ChanUpdateAssetQuote: monitor.chanPollUpdateAssetQuote,
		ChanError:            monitor.chanError,
		UnaryAPI:             unaryAPI,
	}
	monitor.poller = poller.NewPoller(ctx, pollerConfig)

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

func (m *MonitorCoinbase) GetAssetQuotes(ignoreCache ...bool) ([]c.AssetQuote, error) {
	if len(ignoreCache) > 0 && ignoreCache[0] {
		assetQuotes, err := m.getAssetQuotesAndReplaceCache()
		if err != nil {
			return []c.AssetQuote{}, err
		}
		return assetQuotes, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.assetQuotesCache, nil
}

func (m *MonitorCoinbase) SetSymbols(productIds []string, nonce int) error {

	var err error

	m.mu.Lock()

	// Deduplicate productIds since input may have duplicates
	slices.Sort(productIds)
	m.productIds = slices.Compact(productIds)
	m.input.productIds = productIds
	m.input.productIdsLookup = make(map[string]bool)
	for _, productId := range productIds {
		m.input.productIdsLookup[productId] = true
	}

	m.mu.Unlock()

	err = m.getUnderlyingAssetsAndUpdateProductIds()
	if err != nil {
		return err
	}

	m.mu.Lock()

	m.productIdsStreaming, m.productIdsPolling = partitionProductIds(m.productIds)

	m.mu.Unlock()

	// Since the symbols have changed, make a synchronous call to get price quotes for the new symbols
	_, err = m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

	// Coinbase steaming API for CBE (spot) only and not CDE (futures)
	m.streamer.SetSymbolsAndUpdateSubscriptions(m.productIdsStreaming, nonce)
	m.poller.SetSymbols(m.productIdsPolling, nonce)

	return nil

}

// Start the monitor
func (m *MonitorCoinbase) Start() error {

	var err error

	if m.isStarted {
		return fmt.Errorf("monitor already started")
	}

	// On start, get initial quotes from unary API
	_, err = m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

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

func isStreamingProductId(productId string) bool {
	return !hasUnderlyingProductId(productId)
}

func hasUnderlyingProductId(productId string) bool {
	return strings.HasSuffix(productId, "-CDE") || strings.HasPrefix(productId, "CDE")
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

func mergeProductIds(symbolsA, symbolsB []string) []string {
	merged := make([]string, 0, len(symbolsA)+len(symbolsB))
	merged = append(merged, symbolsA...)
	merged = append(merged, symbolsB...)
	slices.Sort(merged)
	return merged
}

func (m *MonitorCoinbase) handleUpdates() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.chanStreamUpdateQuoteExtended:
			// TODO: handle extended quote
			continue

		case updateMessage := <-m.chanPollUpdateAssetQuote:

			// Check if cache exists and values have changed before acquiring write lock
			m.mu.RLock()

			assetQuote, exists := m.assetQuotesCacheLookup[updateMessage.ID]

			if !exists {
				// If product id does not exist in cache, skip update
				// TODO: log product not found in cache - should not happen
				m.mu.RUnlock()
				continue
			}

			// Skip update if nothing has changed
			if assetQuote.QuotePrice.Price == updateMessage.Data.QuotePrice.Price &&
				assetQuote.Exchange.IsActive == updateMessage.Data.Exchange.IsActive &&
				assetQuote.QuotePrice.PriceDayHigh == updateMessage.Data.QuotePrice.PriceDayHigh {

				m.mu.RUnlock()
				continue
			}
			m.mu.RUnlock()

			// Price is different so update cache
			m.mu.Lock()

			assetQuote.QuotePrice.Price = updateMessage.Data.QuotePrice.Price
			assetQuote.QuotePrice.Change = updateMessage.Data.QuotePrice.Change
			assetQuote.QuotePrice.ChangePercent = updateMessage.Data.QuotePrice.ChangePercent
			assetQuote.QuotePrice.PriceDayHigh = updateMessage.Data.QuotePrice.PriceDayHigh
			assetQuote.QuotePrice.PriceDayLow = updateMessage.Data.QuotePrice.PriceDayLow
			assetQuote.QuotePrice.PriceOpen = updateMessage.Data.QuotePrice.PriceOpen
			assetQuote.QuotePrice.PricePrevClose = updateMessage.Data.QuotePrice.PricePrevClose
			assetQuote.QuoteExtended.FiftyTwoWeekHigh = updateMessage.Data.QuoteExtended.FiftyTwoWeekHigh
			assetQuote.QuoteExtended.FiftyTwoWeekLow = updateMessage.Data.QuoteExtended.FiftyTwoWeekLow
			assetQuote.QuoteExtended.MarketCap = updateMessage.Data.QuoteExtended.MarketCap
			assetQuote.QuoteExtended.Volume = updateMessage.Data.QuoteExtended.Volume
			assetQuote.Exchange.IsActive = updateMessage.Data.Exchange.IsActive
			assetQuote.Exchange.IsRegularTradingSession = updateMessage.Data.Exchange.IsRegularTradingSession

			m.mu.Unlock()

			// Send a message with an updated quote
			m.chanUpdateAssetQuote <- c.MessageUpdate[c.AssetQuote]{
				ID:    assetQuote.Symbol,
				Data:  *assetQuote,
				Nonce: updateMessage.Nonce,
			}

			continue

		case updateMessage := <-m.chanStreamUpdateQuotePrice:

			var assetQuote *c.AssetQuote
			var exists bool

			// Check if cache exists and values have changed before acquiring write lock
			m.mu.RLock()

			assetQuote, exists = m.assetQuotesCacheLookup[updateMessage.ID]

			if !exists {
				// If product id does not exist in cache, skip update
				// TODO: log product not found in cache - should not happen
				m.mu.RUnlock()
				continue
			}

			// Skip update if price has not changed
			if assetQuote.QuotePrice.Price == updateMessage.Data.Price {
				m.mu.RUnlock()
				continue
			}
			m.mu.RUnlock()

			// Price is different so update cache
			m.mu.Lock()

			assetQuote.QuotePrice.Price = updateMessage.Data.Price
			assetQuote.QuotePrice.Change = updateMessage.Data.Change
			assetQuote.QuotePrice.ChangePercent = updateMessage.Data.ChangePercent
			assetQuote.QuotePrice.PriceDayHigh = updateMessage.Data.PriceDayHigh
			assetQuote.QuotePrice.PriceDayLow = updateMessage.Data.PriceDayLow
			assetQuote.QuotePrice.PriceOpen = updateMessage.Data.PriceOpen
			assetQuote.QuotePrice.PricePrevClose = updateMessage.Data.PricePrevClose

			m.mu.Unlock()

			// TODO: when underlying asset price changes, callback to update basis else skip

			// Send a message with an updated quote
			m.chanUpdateAssetQuote <- c.MessageUpdate[c.AssetQuote]{
				ID:    assetQuote.Symbol,
				Data:  *assetQuote,
				Nonce: updateMessage.Nonce,
			}

			continue
		default:
		}
	}
}

// Get asset quotes from unary API, add futures quotes, filter out assets not explicitly requested, and replace the asset quotes cache
func (m *MonitorCoinbase) getAssetQuotesAndReplaceCache() ([]c.AssetQuote, error) {

	assetQuotes, assetQuotesByProductId, err := m.unaryAPI.GetAssetQuotes(m.productIds)
	if err != nil {
		return []c.AssetQuote{}, err
	}

	// Filter asset quotes to only include explicitly requested ones
	assetQuotesEnriched := make([]c.AssetQuote, 0, len(m.input.productIds))

	for _, quote := range assetQuotes {
		// Check if this quote is explicitly requested and if not, skip
		if !m.input.productIdsLookup[quote.Meta.SymbolInSourceAPI] {
			continue
		}

		// Check if quote is a futures contract and if yes add properties based on the underlying asset quote
		if quote.Class == c.AssetClassFuturesContract {

			// Check if there is a quote for the underlying asset
			if quoteUnderlying, exists := assetQuotesByProductId[quote.QuoteFutures.SymbolUnderlying]; exists {
				quote.QuoteFutures.IndexPrice = quoteUnderlying.QuotePrice.Price
				quote.QuoteFutures.Basis = (quoteUnderlying.QuotePrice.Price - quote.QuotePrice.Price) / quote.QuotePrice.Price
			}
		}

		assetQuotesEnriched = append(assetQuotesEnriched, quote)

	}

	// Lock updates to asset quotes while symbols are changed and subscriptions updates. ensure data from unary call supercedes potentially oudated streaming data
	m.mu.Lock()
	defer m.mu.Unlock()

	m.assetQuotesCache = assetQuotesEnriched
	m.assetQuotesCacheLookup = assetQuotesByProductId

	return m.assetQuotesCache, nil
}

func (m *MonitorCoinbase) getUnderlyingAssetsAndUpdateProductIds() error {

	underlyingSymbolsResponse := make([]string, 0)
	symbolsWithUnderlying := make([]string, 0)

	m.mu.RLock()

	// Filter productIds to only include those that have underlying assets
	for _, productId := range m.productIds {

		// Skip if productId has already been mapped to an underlying symbol
		if _, exists := m.productIdsToUnderlyingProductIds[productId]; exists {
			continue
		}

		// Append productId for productIds that have underlying assets and have not been mapped yet
		if hasUnderlyingProductId(productId) {
			symbolsWithUnderlying = append(symbolsWithUnderlying, productId)
		}
	}

	m.mu.RUnlock()

	// No new symbols with underlying assets so return early
	if len(symbolsWithUnderlying) == 0 {
		return nil
	}

	// Get new quotes for symbols with underlying assets in order to get their underlying symbols
	underlyingAssetQuotes, _, err := m.unaryAPI.GetAssetQuotes(symbolsWithUnderlying)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, quote := range underlyingAssetQuotes {
		// Add mapping between symbol and underlying symbol to map for lookup
		m.productIdsToUnderlyingProductIds[quote.Meta.SymbolInSourceAPI] = quote.QuoteFutures.SymbolUnderlying

		// Append underlying symbol to list of all symbols
		underlyingSymbolsResponse = append(underlyingSymbolsResponse, quote.QuoteFutures.SymbolUnderlying)
	}

	// Merge and deduplicate productIds since and underlying symbol could also have been explicitly requested
	m.productIds = mergeProductIds(m.productIds, underlyingSymbolsResponse)
	slices.Sort(m.productIds)
	m.productIds = slices.Compact(m.productIds)

	return nil
}
