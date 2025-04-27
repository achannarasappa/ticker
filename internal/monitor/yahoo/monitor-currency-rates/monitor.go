package monitorCurrencyRate

import (
	"context"
	"fmt"
	"sync"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
)

// MonitorCurrencyRatesYahoo represents a Yahoo Finance monitor
type MonitorCurrencyRateYahoo struct {
	unaryAPI                 *unary.UnaryAPI
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
}

// NewMonitorCurrencyRateYahoo creates a new MonitorCurrencyRateYahoo
func NewMonitorCurrencyRateYahoo(config Config) *MonitorCurrencyRateYahoo {

	ctx, cancel := context.WithCancel(config.Ctx)

	monitor := &MonitorCurrencyRateYahoo{
		unaryAPI:                 config.UnaryAPI,
		ctx:                      ctx,
		cancel:                   cancel,
		chanUpdateCurrencyRates:  config.ChanUpdateCurrencyRates,
		chanRequestCurrencyRates: config.ChanRequestCurrencyRates,
		chanError:                config.ChanError,
	}

	return monitor
}

func (m *MonitorCurrencyRateYahoo) Start() error {

	if m.isStarted {
		return fmt.Errorf("monitor already started")
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
		return fmt.Errorf("monitor not started")
	}

	m.cancel()
	m.isStarted = false

	return nil

}

func (m *MonitorCurrencyRateYahoo) SetTargetCurrency(targetCurrency string) {

	fromCurrencies := make([]string, 0)

	m.mu.RLock()
	for currency := range m.currencyRateCache {
		fromCurrencies = append(fromCurrencies, currency)
	}
	m.mu.RUnlock()

	rates, err := m.unaryAPI.GetCurrencyRates(fromCurrencies, m.targetCurrency)
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

			// Get currency rates from Yahoo unary API
			rates, err := m.unaryAPI.GetCurrencyRates(fromCurrenciesToRequest, m.targetCurrency)
			if err != nil {
				m.chanError <- err
				continue
			}

			// Update the cache
			m.mu.Lock()
			m.currencyRateCache = rates
			m.mu.Unlock()

			m.chanUpdateCurrencyRates <- m.currencyRateCache
		}
	}
}
