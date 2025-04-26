package monitorPriceYahoo

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/monitor-price/poller"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
)

// MonitorYahoo represents a Yahoo Finance monitor
type MonitorYahoo struct {
	unaryAPI                 *unary.UnaryAPI
	poller                   *poller.Poller
	unary                    *unary.UnaryAPI
	input                    input
	symbols                  []string
	symbolToCurrency         map[string]string        // Map of symbols to currency
	assetQuotesCache         []c.AssetQuote           // Asset quotes for all assets retrieved at start or on symbol change
	assetQuotesCacheLookup   map[string]*c.AssetQuote // Asset quotes for all assets retrieved at least once (symbol change does not remove symbols)
	chanPollUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
	chanError                chan error
	mu                       sync.RWMutex
	ctx                      context.Context
	cancel                   context.CancelFunc
	isStarted                bool
	chanUpdateAssetQuote     chan c.MessageUpdate[c.AssetQuote]
}

// input represents user input for the Yahoo monitor with any transformation
type input struct {
	symbols       []string
	symbolsLookup map[string]bool
}

// Config contains the required configuration for the Yahoo monitor
type Config struct {
	Ctx                  context.Context
	UnaryURL             string
	SessionRootURL       string
	SessionCrumbURL      string
	SessionConsentURL    string
	ChanError            chan error
	ChanUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
}

// Option defines an option for configuring the monitor
type Option func(*MonitorYahoo)

func NewMonitorYahoo(config Config, opts ...Option) *MonitorYahoo {

	ctx, cancel := context.WithCancel(config.Ctx)

	unaryAPI := unary.NewUnaryAPI(unary.Config{
		BaseURL:           config.UnaryURL,
		SessionRootURL:    config.SessionRootURL,
		SessionCrumbURL:   config.SessionCrumbURL,
		SessionConsentURL: config.SessionConsentURL,
	})

	monitor := &MonitorYahoo{
		assetQuotesCacheLookup:   make(map[string]*c.AssetQuote),
		symbolToCurrency:         make(map[string]string),
		assetQuotesCache:         make([]c.AssetQuote, 0),
		chanPollUpdateAssetQuote: make(chan c.MessageUpdate[c.AssetQuote]),
		chanError:                config.ChanError,
		unaryAPI:                 unaryAPI,
		ctx:                      ctx,
		cancel:                   cancel,
		chanUpdateAssetQuote:     config.ChanUpdateAssetQuote,
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
	return func(m *MonitorYahoo) {
		m.poller.SetRefreshInterval(interval)
	}
}

// GetAssetQuotes returns the asset quotes either from the cache or from the unary API if ignoreCache is set
func (m *MonitorYahoo) GetAssetQuotes(ignoreCache ...bool) ([]c.AssetQuote, error) {

	// If ignoreCache is set, get the asset quotes from the unary API and update the cache
	if len(ignoreCache) > 0 && ignoreCache[0] {
		assetQuotes, err := m.getAssetQuotesAndReplaceCache()
		if err != nil {
			return []c.AssetQuote{}, err
		}
		return assetQuotes, nil
	}

	// If ignoreCache is not set, return the asset quotes from the cache without making a HTTP request
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.assetQuotesCache, nil
}

// SetSymbols sets the symbols to monitor
func (m *MonitorYahoo) SetSymbols(symbols []string, nonce int) error {

	var err error

	m.mu.Lock()

	// Deduplicate symbols since input may have duplicates
	slices.Sort(symbols)
	m.symbols = slices.Compact(symbols)
	m.input.symbols = symbols
	m.input.symbolsLookup = make(map[string]bool)
	for _, symbol := range symbols {
		m.input.symbolsLookup[symbol] = true
	}

	m.mu.Unlock()

	// Check for symbols which don't have a known currency and retrieve the currency for those symbols
	// TODO: conditionally bypass if currency conversion is not enabled
	// err = m.getCurrencyForEachSymbolAndUpdateCurrencyMap()
	// if err != nil {
	// 	return err
	// }

	// Since the symbols have changed, make a synchronous call to get price quotes for the new symbols
	_, err = m.getAssetQuotesAndReplaceCache()
	if err != nil {
		return err
	}

	// Set the symbols to monitor on the poller
	m.poller.SetSymbols(m.symbols, nonce)

	return nil

}

// Start the monitor
func (m *MonitorYahoo) Start() error {

	var err error

	if m.isStarted {
		return fmt.Errorf("monitor already started")
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
func (m *MonitorYahoo) Stop() error {

	if !m.isStarted {
		return fmt.Errorf("monitor not started")
	}

	m.cancel()
	return nil
}

// handleUpdates listens for asset quote change messages and updates the cache
func (m *MonitorYahoo) handleUpdates() {
	for {
		select {
		case <-m.ctx.Done():
			return

		case updateMessage := <-m.chanPollUpdateAssetQuote:
			// Check if cache exists and values have changed before acquiring write lock
			m.mu.RLock()

			assetQuote, exists := m.assetQuotesCacheLookup[updateMessage.ID]

			// If product id does not exist in cache, skip update (this would happen if the API returns a price for a symbol that was not requested)
			if !exists {
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

			// Update properties on the asset quote which may have changed
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

		default:
		}
	}
}

// Get asset quotes from unary API, add futures quotes, filter out assets not explicitly requested, and replace the asset quotes cache
func (m *MonitorYahoo) getAssetQuotesAndReplaceCache() ([]c.AssetQuote, error) {

	// Make a synchronous call to get price quotes
	assetQuotes, assetQuotesByProductId, err := m.unaryAPI.GetAssetQuotes(m.symbols)
	if err != nil {
		return []c.AssetQuote{}, err
	}

	// Replace the cache with new sets of asset quotes
	m.mu.Lock()
	defer m.mu.Unlock()

	m.assetQuotesCache = assetQuotes
	m.assetQuotesCacheLookup = assetQuotesByProductId

	return m.assetQuotesCache, nil
}

func (m *MonitorYahoo) getCurrencyForEachSymbolAndUpdateCurrencyMap() error {
	// No need to process if no symbols are provided
	if len(m.symbols) == 0 {
		return nil
	}

	symbolsWithoutCurrency := make([]string, 0)

	// Check if symbols already have a currency mapping
	for _, symbol := range m.symbols {
		if _, exists := m.symbolToCurrency[symbol]; !exists {
			symbolsWithoutCurrency = append(symbolsWithoutCurrency, symbol)
		}
	}

	// Get currency information for each symbol
	symbolToCurrency, err := m.unaryAPI.GetCurrencyMap(symbolsWithoutCurrency)
	if err != nil {
		return fmt.Errorf("failed to get currency information: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update currency information in the existing asset quotes cache
	for symbol, currency := range symbolToCurrency {
		m.symbolToCurrency[symbol] = currency.FromCurrency
	}

	return nil
}

// GetCurrencyRates accepts an array of ISO 4217 currency codes and a target ISO 4217 currency code and returns a conversion rate for each of the input currencies to the target currency
func (m *MonitorYahoo) GetCurrencyRates(inputCurrencies []string, targetCurrency string) (c.CurrencyRates, error) {

	// currencyRates, err := m.unaryAPI.GetCurrencyRates(inputCurrencies, targetCurrency)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get currency rates: %w", err)
	// }

	// return currencyRates, nil
	return nil, nil
}
