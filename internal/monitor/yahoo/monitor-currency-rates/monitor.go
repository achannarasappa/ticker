package monitorCurrencyRate

import (
	"context"
	"errors"
	"maps"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
)

const (
	// cacheKeyCurrencyRate namespaces a single from -> to currency rate so
	// overlapping currencies are shared across instances regardless of the full
	// set requested.
	cacheKeyCurrencyRate = "yahoo:currency-rate:"
	// ttlCurrencyRates is kept short since exchange rates drift continuously.
	ttlCurrencyRates = time.Hour
)

// currencyRateKey builds the cache key for a single from -> to currency rate.
func currencyRateKey(fromCurrency, toCurrency string) string {
	return cacheKeyCurrencyRate + toCurrency + ":" + fromCurrency
}

// MonitorCurrencyRatesYahoo represents a Yahoo Finance monitor
type MonitorCurrencyRateYahoo struct {
	unaryAPI                 *unary.UnaryAPI
	cache                    c.Cache
	ctx                      context.Context
	cancel                   context.CancelFunc
	currencyRateCache        map[string]c.CurrencyRate
	targetCurrency           string
	mu                       sync.RWMutex
	isStarted                bool
	chanUpdateCurrencyRates  chan c.CurrencyRates
	chanRequestCurrencyRates chan []string
	chanError                chan error
}

// Config contains the required configuration for the Yahoo monitor
type Config struct {
	Ctx                      context.Context
	UnaryAPI                 *unary.UnaryAPI
	ChanUpdateCurrencyRates  chan c.CurrencyRates
	ChanRequestCurrencyRates chan []string
	ChanError                chan error
	Cache                    c.Cache
}

// NewMonitorCurrencyRateYahoo creates a new MonitorCurrencyRateYahoo
func NewMonitorCurrencyRateYahoo(config Config) *MonitorCurrencyRateYahoo {

	ctx, cancel := context.WithCancel(config.Ctx)

	monitor := &MonitorCurrencyRateYahoo{
		unaryAPI:                 config.UnaryAPI,
		cache:                    config.Cache,
		ctx:                      ctx,
		cancel:                   cancel,
		chanUpdateCurrencyRates:  config.ChanUpdateCurrencyRates,
		chanRequestCurrencyRates: config.ChanRequestCurrencyRates,
		chanError:                config.ChanError,
		currencyRateCache:        make(map[string]c.CurrencyRate),
	}

	return monitor
}

func (m *MonitorCurrencyRateYahoo) Start() error {

	if m.isStarted {
		return errors.New("monitor already started")
	}

	if m.targetCurrency == "" {
		m.targetCurrency = "USD"
	}

	go m.handleRequestCurrencyRates()

	m.isStarted = true

	return nil
}

func (m *MonitorCurrencyRateYahoo) Stop() error {

	if !m.isStarted {
		return errors.New("monitor not started")
	}

	m.cancel()
	m.isStarted = false

	return nil

}

func (m *MonitorCurrencyRateYahoo) SetTargetCurrency(targetCurrency string) {

	fromCurrencies := make([]string, 0, len(m.currencyRateCache))

	m.mu.RLock()
	for currency := range m.currencyRateCache {
		fromCurrencies = append(fromCurrencies, currency)
	}
	m.mu.RUnlock()

	rates, err := m.unaryAPI.GetCurrencyRates(fromCurrencies, targetCurrency)

	if err != nil {
		m.chanError <- err

		return
	}

	m.mu.Lock()
	m.targetCurrency = targetCurrency
	m.currencyRateCache = rates
	m.mu.Unlock()

}

func (m *MonitorCurrencyRateYahoo) handleRequestCurrencyRates() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case fromCurrencies := <-m.chanRequestCurrencyRates:

			// Skip if no fromCurrencies
			if len(fromCurrencies) == 0 {
				continue
			}

			// Filter out currencies that are already in the cache
			fromCurrenciesToRequest := make([]string, 0)
			m.mu.RLock()
			for _, currency := range fromCurrencies {
				if _, exists := m.currencyRateCache[currency]; !exists {
					fromCurrenciesToRequest = append(fromCurrenciesToRequest, currency)
				}
			}
			m.mu.RUnlock()

			// Skip if no new currencies to fetch
			if len(fromCurrenciesToRequest) == 0 {

				continue
			}

			// Resolve rates from the on-disk cache
			resolvedRates := make(map[string]c.CurrencyRate)
			currenciesToFetch := make([]string, 0, len(fromCurrenciesToRequest))

			for _, currency := range fromCurrenciesToRequest {
				var cached c.CurrencyRate
				if m.cache != nil && m.cache.Get(currencyRateKey(currency, m.targetCurrency), &cached) {
					resolvedRates[currency] = cached

					continue
				}

				currenciesToFetch = append(currenciesToFetch, currency)
			}

			if len(currenciesToFetch) > 0 {
				// Get currency rates from Yahoo unary API
				rates, err := m.unaryAPI.GetCurrencyRates(currenciesToFetch, m.targetCurrency)
				if err != nil {
					m.chanError <- err

					continue
				}

				for currency, rate := range rates {
					resolvedRates[currency] = rate

					if m.cache != nil {
						m.cache.Set(currencyRateKey(currency, m.targetCurrency), rate, ttlCurrencyRates)
					}
				}
			}

			// Update the cache
			m.mu.Lock()
			maps.Copy(m.currencyRateCache, resolvedRates)
			m.mu.Unlock()
			m.chanUpdateCurrencyRates <- m.currencyRateCache
		}
	}
}
