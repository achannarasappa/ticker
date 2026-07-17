package monitorPriceCoinPaprika

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	poller "github.com/achannarasappa/ticker/v5/internal/monitor/coinpaprika/monitor-price/poller"
	unary "github.com/achannarasappa/ticker/v5/internal/monitor/coinpaprika/unary"
)

const (
	fromCurrencyCode = "USD"
	// cacheKeyCoinID namespaces cached symbol to CoinPaprika coin id mappings
	cacheKeyCoinID = "coinpaprika:coin-id:"
	// ttlCoinID caches the mapping from a ticker symbol to a CoinPaprika coin
	// id, which is effectively static, so it can be reused for a long time
	ttlCoinID = 7 * 24 * time.Hour
)

// MonitorPriceCoinPaprika represents a CoinPaprika price monitor
type MonitorPriceCoinPaprika struct {
	unaryAPI                  *unary.UnaryAPI
	poller                    *poller.Poller
	cache                     c.Cache
	symbols                   []string          // Ticker symbols requested by the user (e.g. BTC)
	coinIds                   []string          // CoinPaprika coin ids for the requested symbols (e.g. btc-bitcoin)
	symbolToCoinID            map[string]string // Map of ticker symbols to CoinPaprika coin ids
	assetQuotesCache          []*c.AssetQuote
	assetQuotesCacheLookup    map[string]*c.AssetQuote // Asset quotes by CoinPaprika coin id
	currencyRatesCache        c.CurrencyRates
	currencyHasRequestedRates bool // Whether currency rates have been requested; all CoinPaprika quotes are denominated in USD
	chanPollUpdateAssetQuote  chan c.MessageUpdate[c.AssetQuote]
	chanError                 chan error
	mu                        sync.RWMutex
	muCurrencyRates           sync.RWMutex
	ctx                       context.Context
	cancel                    context.CancelFunc
	isStarted                 bool
	chanUpdateAssetQuote      chan c.MessageUpdate[c.AssetQuote]
	chanRequestCurrencyRates  chan []string
}

// Config contains the required configuration for the CoinPaprika monitor
type Config struct {
	Ctx                      context.Context
	UnaryURL                 string
	ChanError                chan error
	ChanUpdateAssetQuote     chan c.MessageUpdate[c.AssetQuote]
	ChanRequestCurrencyRates chan []string
	Cache                    c.Cache
}

// Option defines an option for configuring the monitor
type Option func(*MonitorPriceCoinPaprika)

func NewMonitorPriceCoinPaprika(config Config, opts ...Option) *MonitorPriceCoinPaprika {
	ctx, cancel := context.WithCancel(config.Ctx)

	unaryAPI := unary.NewUnaryAPI(config.UnaryURL)

	monitor := &MonitorPriceCoinPaprika{
		symbolToCoinID:           make(map[string]string),
		assetQuotesCacheLookup:   make(map[string]*c.AssetQuote),
		assetQuotesCache:         make([]*c.AssetQuote, 0),
		chanPollUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote]),
		chanError:                config.ChanError,
		unaryAPI:                 unaryAPI,
		cache:                    config.Cache,
		ctx:                      ctx,
		cancel:                   cancel,
		chanUpdateAssetQuote:     config.ChanUpdateAssetQuote,
		chanRequestCurrencyRates: config.ChanRequestCurrencyRates,
	}

	pollerConfig := poller.PollerConfig{
		ChanUpdateAssetQuote: monitor.chanPollUpdateAssetQuote,
		ChanError:            monitor.chanError,
		UnaryAPI:             unaryAPI,
	}
	monitor.poller = poller.NewPoller(ctx, pollerConfig)

	for _, opt := range opts {
		opt(monitor)
	}

	return monitor
}

// WithRefreshInterval sets the refresh interval for the monitor
func WithRefreshInterval(interval time.Duration) Option {
	return func(m *MonitorPriceCoinPaprika) {
		// TODO: handle error
		m.poller.SetRefreshInterval(interval) //nolint:errcheck
	}
}

// GetAssetQuotes returns the asset quotes either from the cache or from the unary API if ignoreCache is set
func (m *MonitorPriceCoinPaprika) GetAssetQuotes(ignoreCache ...bool) ([]c.AssetQuote, error) {

	// If ignoreCache is set, get the asset quotes from the unary API and update the cache
	if len(ignoreCache) > 0 && ignoreCache[0] {
		assetQuotes, err := m.getAssetQuotesAndReplaceCache()
		if err != nil {
			return []c.AssetQuote{}, err
		}
		result := make([]c.AssetQuote, len(assetQuotes))
		for i, quote := range assetQuotes {
			result[i] = *quote
		}

		return result, nil
	}

	// If ignoreCache is not set, return the asset quotes from the cache without making a HTTP request
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]c.AssetQuote, len(m.assetQuotesCache))
	for i, quote := range m.assetQuotesCache {
		result[i] = *quote
	}

	return result, nil
}

// SetSymbols sets the ticker symbols to monitor
func (m *MonitorPriceCoinPaprika) SetSymbols(symbols []string, versionVector int) error {

	var err error

	m.mu.Lock()

	// Deduplicate symbols since input may have duplicates
	slices.Sort(symbols)
	m.symbols = slices.Compact(symbols)

	m.mu.Unlock()

	m.muCurrencyRates.Lock()

	// Request the default currency (USD) once since all CoinPaprika quotes are denominated in USD
	if !m.currencyHasRequestedRates {
		m.chanRequestCurrencyRates <- []string{fromCurrencyCode}
		m.currencyHasRequestedRates = true
	}

	m.muCurrencyRates.Unlock()

	// Resolve ticker symbols to CoinPaprika coin ids for any new symbols
	err = m.getCoinIDForEachSymbolAndUpdateCoinIds()
	if err != nil {
		return err
	}

	// Since the symbols have changed, make a synchronous call to get price quotes for the new symbols
	_, err = m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

	// Set the coin ids to monitor on the poller
	m.poller.SetSymbols(m.coinIds, versionVector)

	return nil

}

// Start the monitor
func (m *MonitorPriceCoinPaprika) Start() error {

	var err error

	if m.isStarted {
		return errors.New("monitor already started")
	}

	// On start, get initial quotes from unary API
	_, err = m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

	// Start polling for price quotes
	err = m.poller.Start()
	if err != nil {
		return err
	}

	// Start listening for price quote updates
	go m.handleUpdates()

	m.isStarted = true

	return nil
}

// Stop the monitor
func (m *MonitorPriceCoinPaprika) Stop() error {

	if !m.isStarted {
		return errors.New("monitor not started")
	}

	m.cancel()

	return nil
}

// SetCurrencyRates sets the currency rates for the monitor and updates the asset quotes
func (m *MonitorPriceCoinPaprika) SetCurrencyRates(currencyRates c.CurrencyRates) error {
	m.muCurrencyRates.Lock()
	m.currencyRatesCache = currencyRates
	m.muCurrencyRates.Unlock()

	// Map over each asset quote and update the currency rate
	_, err := m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

	return nil
}

// handleUpdates listens for asset quote change messages and updates the cache
func (m *MonitorPriceCoinPaprika) handleUpdates() {
	for {
		select {
		case <-m.ctx.Done():
			return

		case updateMessage := <-m.chanPollUpdateAssetQuote:
			// Check if cache exists and values have changed before acquiring write lock
			m.mu.RLock()

			assetQuote, exists := m.assetQuotesCacheLookup[updateMessage.ID]

			// If coin id does not exist in cache, skip update (this would happen if the API returns a price for a coin that was not requested)
			if !exists {
				m.mu.RUnlock()

				continue
			}

			// Skip update if nothing has changed
			if assetQuote.QuotePrice.Price == updateMessage.Data.QuotePrice.Price {
				m.mu.RUnlock()

				continue
			}
			m.mu.RUnlock()

			// Price is different so update cache
			m.mu.Lock()

			// Update properties on the asset quote which may have changed
			assetQuote.QuotePrice.Price = updateMessage.Data.QuotePrice.Price
			assetQuote.QuotePrice.Change = updateMessage.Data.QuotePrice.Change
			assetQuote.QuotePrice.ChangePercent = updateMessage.Data.QuotePrice.ChangePercent
			assetQuote.QuoteExtended.MarketCap = updateMessage.Data.QuoteExtended.MarketCap
			assetQuote.QuoteExtended.Volume = updateMessage.Data.QuoteExtended.Volume

			m.mu.Unlock()

			// Send a message with an updated quote
			m.chanUpdateAssetQuote <- c.MessageUpdate[c.AssetQuote]{
				ID:            assetQuote.Symbol,
				Data:          *assetQuote,
				VersionVector: updateMessage.VersionVector,
			}

			continue
		}
	}
}

// Get asset quotes from the unary API and replace the asset quotes cache
func (m *MonitorPriceCoinPaprika) getAssetQuotesAndReplaceCache() ([]*c.AssetQuote, error) {

	lookup := make(map[string]*c.AssetQuote)
	cache := make([]*c.AssetQuote, 0)

	// Make a synchronous call to get price quotes
	assetQuotes, _, err := m.unaryAPI.GetAssetQuotes(m.coinIds)
	if err != nil {
		return []*c.AssetQuote{}, err
	}

	// Add currency rate to each price quote
	m.muCurrencyRates.RLock()
	for _, quote := range assetQuotes {

		if currencyRate, exists := m.currencyRatesCache[fromCurrencyCode]; exists {
			quote.Currency.Rate = currencyRate.Rate
			quote.Currency.ToCurrencyCode = currencyRate.ToCurrency
		}

		lookup[quote.Meta.SymbolInSourceAPI] = &quote
		cache = append(cache, &quote)
	}
	m.muCurrencyRates.RUnlock()

	// Replace the cache with new sets of asset quotes
	m.mu.Lock()
	defer m.mu.Unlock()

	m.assetQuotesCache = cache
	m.assetQuotesCacheLookup = lookup

	return m.assetQuotesCache, nil
}

// Resolve the CoinPaprika coin id for each symbol without a known coin id and update the list of coin ids to poll
func (m *MonitorPriceCoinPaprika) getCoinIDForEachSymbolAndUpdateCoinIds() error {

	symbolsWithoutCoinID := make([]string, 0)

	m.mu.RLock()

	// Check if symbols already have a coin id mapping
	for _, symbol := range m.symbols {
		if _, exists := m.symbolToCoinID[symbol]; !exists {
			symbolsWithoutCoinID = append(symbolsWithoutCoinID, symbol)
		}
	}

	m.mu.RUnlock()

	// Resolve a symbol's coin id from cache since the coin id for a symbol is effectively static
	symbolToCoinID := make(map[string]string)
	symbolsToFetch := make([]string, 0, len(symbolsWithoutCoinID))

	for _, symbol := range symbolsWithoutCoinID {
		var cached string
		if m.cache != nil && m.cache.Get(cacheKeyCoinID+symbol, &cached) {
			symbolToCoinID[symbol] = cached

			continue
		}

		symbolsToFetch = append(symbolsToFetch, symbol)
	}

	if len(symbolsToFetch) > 0 {
		// Resolve coin ids for symbols not found in the cache; symbols unknown
		// to CoinPaprika are omitted from the result and skipped
		fetched, err := m.unaryAPI.GetCoinIDs(symbolsToFetch)
		if err != nil {
			return err
		}

		for symbol, coinID := range fetched {
			symbolToCoinID[symbol] = coinID

			if m.cache != nil {
				m.cache.Set(cacheKeyCoinID+symbol, coinID, ttlCoinID)
			}
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for symbol, coinID := range symbolToCoinID {
		m.symbolToCoinID[symbol] = coinID
	}

	// Rebuild the list of coin ids to poll from the requested symbols
	coinIds := make([]string, 0, len(m.symbols))

	for _, symbol := range m.symbols {
		if coinID, exists := m.symbolToCoinID[symbol]; exists {
			coinIds = append(coinIds, coinID)
		}
	}

	slices.Sort(coinIds)
	m.coinIds = slices.Compact(coinIds)

	return nil
}
