package monitor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	monitorPriceCoinbase "github.com/achannarasappa/ticker/v5/internal/monitor/coinbase/monitor-price"
	monitorCurrencyRate "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/monitor-currency-rates"
	monitorPriceYahoo "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/monitor-price"
	unaryClientYahoo "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
)

// Monitor represents an overall monitor which manages API specific monitors
type Monitor struct {
	monitors                map[c.QuoteSource]c.Monitor
	monitorCurrencyRate     c.MonitorCurrencyRate
	chanError               chan error
	chanUpdateAssetQuote    chan c.MessageUpdate[c.AssetQuote]
	chanUpdateCurrencyRates chan c.CurrencyRates
	onUpdateAssetQuote      func(symbol string, assetQuote c.AssetQuote, versionVector int)
	onUpdateAssetGroupQuote func(assetGroupQuote c.AssetGroupQuote, versionVector int)
	assetGroupVersionVector int
	assetGroup              c.AssetGroup
	requestedVersionVector  int          // versionVector of the asset group most recently requested with SetAssetGroup
	attemptsInFlight        atomic.Int32 // attempts to set an asset group with a call to a monitor still running
	refreshInterval         time.Duration
	mu                      sync.RWMutex
	logger                  *log.Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
}

// ConfigMonitor represents the configuration for the main monitor
type ConfigMonitor struct {
	RefreshInterval int
	TargetCurrency  string
	Logger          *log.Logger
	Cache           c.Cache
	ConfigMonitorPriceCoinbase
	ConfigMonitorsYahoo
}

// ConfigMonitorPriceCoinbase represents the configuration for the Coinbase monitor
type ConfigMonitorPriceCoinbase struct {
	BaseURL      string
	StreamingURL string
}

// ConfigMonitorsYahoo represents the configuration for the Yahoo monitors (price and currency rate)
type ConfigMonitorsYahoo struct {
	BaseURL           string
	SessionRootURL    string
	SessionCrumbURL   string
	SessionConsentURL string
}

// ConfigUpdateFns represents the callback functions for when asset quotes are updated
type ConfigUpdateFns struct {
	OnUpdateAssetQuote      func(symbol string, assetQuote c.AssetQuote, versionVector int)
	OnUpdateAssetGroupQuote func(assetGroupQuote c.AssetGroupQuote, versionVector int)
}

// New creates a new instance of the Coinbase monitor
func NewMonitor(configMonitor ConfigMonitor) (*Monitor, error) {

	chanError := make(chan error, 5)
	chanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 10)
	chanUpdateCurrencyRate := make(chan c.CurrencyRates, 10)
	chanRequestCurrencyRate := make(chan []string, 10)

	ctx, cancel := context.WithCancel(context.Background())

	coinbase := monitorPriceCoinbase.NewMonitorPriceCoinbase(
		monitorPriceCoinbase.Config{
			Ctx:                      ctx,
			UnaryURL:                 configMonitor.ConfigMonitorPriceCoinbase.BaseURL,
			ChanError:                chanError,
			ChanUpdateAssetQuote:     chanUpdateAssetQuote,
			ChanRequestCurrencyRates: chanRequestCurrencyRate,
			Cache:                    configMonitor.Cache,
		},
		monitorPriceCoinbase.WithStreamingURL(configMonitor.ConfigMonitorPriceCoinbase.StreamingURL),
		monitorPriceCoinbase.WithRefreshInterval(time.Duration(configMonitor.RefreshInterval)*time.Second),
	)

	// Create and configure the API client for the Yahoo API shared between monitors
	unaryAPI := unaryClientYahoo.NewUnaryAPI(unaryClientYahoo.Config{
		BaseURL:           configMonitor.ConfigMonitorsYahoo.BaseURL,
		SessionRootURL:    configMonitor.ConfigMonitorsYahoo.SessionRootURL,
		SessionCrumbURL:   configMonitor.ConfigMonitorsYahoo.SessionCrumbURL,
		SessionConsentURL: configMonitor.ConfigMonitorsYahoo.SessionConsentURL,
		Cache:             configMonitor.Cache,
	})

	yahoo := monitorPriceYahoo.NewMonitorPriceYahoo(
		monitorPriceYahoo.Config{
			Ctx:                      ctx,
			UnaryAPI:                 unaryAPI,
			ChanError:                chanError,
			ChanUpdateAssetQuote:     chanUpdateAssetQuote,
			ChanRequestCurrencyRates: chanRequestCurrencyRate,
			Cache:                    configMonitor.Cache,
		},
		monitorPriceYahoo.WithRefreshInterval(time.Duration(configMonitor.RefreshInterval)*time.Second),
	)

	yahooCurrencyRate := monitorCurrencyRate.NewMonitorCurrencyRateYahoo(
		monitorCurrencyRate.Config{
			Ctx:                      ctx,
			UnaryAPI:                 unaryAPI,
			ChanUpdateCurrencyRates:  chanUpdateCurrencyRate,
			ChanRequestCurrencyRates: chanRequestCurrencyRate,
			ChanError:                chanError,
			Cache:                    configMonitor.Cache,
		},
	)

	yahooCurrencyRate.SetTargetCurrency(configMonitor.TargetCurrency)

	m := &Monitor{
		monitors: map[c.QuoteSource]c.Monitor{
			c.QuoteSourceCoinbase: coinbase,
			c.QuoteSourceYahoo:    yahoo,
		},
		monitorCurrencyRate:     yahooCurrencyRate,
		chanUpdateAssetQuote:    chanUpdateAssetQuote,
		chanUpdateCurrencyRates: chanUpdateCurrencyRate,
		chanError:               chanError,
		onUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, versionVector int) {},
		onUpdateAssetQuote:      func(symbol string, assetQuote c.AssetQuote, versionVector int) {},
		refreshInterval:         getRetryInterval(configMonitor.RefreshInterval),
		logger:                  configMonitor.Logger,
		ctx:                     ctx,
		cancel:                  cancel,
	}

	return m, nil
}

// SetAssetGroup sets the asset group for the monitor
//
// If the symbols cannot be set on every monitor, for example because an API is unreachable or
// rate limited when ticker starts, the error is returned and setting the asset group is retried
// at the refresh interval until it succeeds or a newer asset group is set. Otherwise nothing
// would request quotes for the asset group again and the watchlist would stay empty.
func (m *Monitor) SetAssetGroup(assetGroup c.AssetGroup, versionVector int) error {
	m.mu.Lock()
	m.requestedVersionVector = versionVector
	m.mu.Unlock()

	err := m.setAssetGroup(assetGroup, versionVector)

	if err != nil {
		go m.retrySetAssetGroup(assetGroup, versionVector)
	}

	return err
}

// retrySetAssetGroup sets the asset group at the refresh interval until it succeeds, a newer asset group is requested, or the monitor is stopped
func (m *Monitor) retrySetAssetGroup(assetGroup c.AssetGroup, versionVector int) {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():

			return
		case <-ticker.C:
			if m.isSuperseded(versionVector) {

				return
			}

			// Let an earlier attempt that timed out finish rather than setting symbols on the same monitors concurrently
			if m.attemptsInFlight.Load() > 0 {

				continue
			}

			err := m.setAssetGroup(assetGroup, versionVector)

			if err == nil {

				return
			}

			if m.logger != nil {
				m.logger.Printf("%v", err)
			}
		}
	}
}

// isSuperseded returns true if an asset group newer than the one with versionVector has been requested
func (m *Monitor) isSuperseded(versionVector int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return versionVector != m.requestedVersionVector
}

// setAssetGroup makes one attempt to set the symbols for each monitor and to send the asset group quote
func (m *Monitor) setAssetGroup(assetGroup c.AssetGroup, versionVector int) error {
	var wg sync.WaitGroup

	// Create a channel for timeout
	done := make(chan bool)
	// Create error channel for collecting errors from each monitor
	chanError := make(chan error, len(assetGroup.SymbolsBySource))
	// Create a slice to collect errors
	var errors []error

	// Count this attempt as in flight until every monitor has returned, which can be after a timeout
	m.attemptsInFlight.Add(1)

	// Concurrently set symbols for each monitor (execute a synchronous call to update quotes for each monitor)
	for _, symbolBySource := range assetGroup.SymbolsBySource {
		if monitor, exists := m.monitors[symbolBySource.Source]; exists {
			wg.Add(1)
			go func(mon c.Monitor, symbols []string) {
				defer wg.Done()
				err := mon.SetSymbols(symbols, versionVector)
				if err != nil {
					chanError <- err
				}
			}(monitor, symbolBySource.Symbols)
		}
	}

	// Wait for the waitgroup to finish in the background
	go func() {
		wg.Wait()
		m.attemptsInFlight.Add(-1)
		close(done)
	}()

	// Continue when the waitgroup is finished or a timeout is reached
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-done:
			// If there are any errors, return them
			if len(errors) > 0 {

				return fmt.Errorf("errors setting symbols on monitor(s): %v", errors)
			}

			goto Continue
		case err := <-chanError:
			errors = append(errors, err)
		case <-timeout:

			// If there are any errors, return them along with the timeout error
			return fmt.Errorf("timeout waiting for monitor(s) to set symbols. Additional non-timeout errors: %v", errors)
		}
	}
Continue:

	m.mu.Lock()
	defer m.mu.Unlock()

	// Do not replace an asset group that was requested after this one
	if versionVector != m.requestedVersionVector {

		return nil
	}

	// Update the versionVector so that any messages from the previous asset group can be ignored
	m.assetGroupVersionVector = versionVector
	m.assetGroup = assetGroup

	// Get asset quotes for all sources
	assetGroupQuote := m.GetAssetGroupQuote()

	// Run the callback in a goroutine to avoid blocking
	go m.onUpdateAssetGroupQuote(assetGroupQuote, versionVector)

	return nil
}

// SetOnUpdate sets the callback functions for when asset quotes are updated
func (m *Monitor) SetOnUpdate(config ConfigUpdateFns) error {

	if config.OnUpdateAssetQuote == nil || config.OnUpdateAssetGroupQuote == nil {
		return errors.New("onUpdateAssetQuote and onUpdateAssetGroupQuote must be set")
	}

	m.onUpdateAssetQuote = config.OnUpdateAssetQuote
	m.onUpdateAssetGroupQuote = config.OnUpdateAssetGroupQuote

	return nil
}

// Start starts all monitors
func (m *Monitor) Start() {

	m.monitorCurrencyRate.Start() //nolint:errcheck

	for _, monitor := range m.monitors {
		monitor.Start() //nolint:errcheck
	}

	go m.handleUpdates()
}

// GetAssetGroupQuote synchronously gets price quotes a group of assets across all sources
func (m *Monitor) GetAssetGroupQuote(ignoreCache ...bool) c.AssetGroupQuote {

	assetQuotesFromAllSources := make([]c.AssetQuote, 0)

	for _, symbolBySource := range m.assetGroup.SymbolsBySource {

		assetQuotes, _ := m.monitors[symbolBySource.Source].GetAssetQuotes(ignoreCache...)
		assetQuotesFromAllSources = append(assetQuotesFromAllSources, assetQuotes...)

	}

	return c.AssetGroupQuote{
		AssetQuotes: assetQuotesFromAllSources,
		AssetGroup:  m.assetGroup,
	}
}

// handleUpdates listens for asset quote updates and errors from monitors
func (m *Monitor) handleUpdates() {
	for {
		select {
		case <-m.ctx.Done():

			return
		case update := <-m.chanUpdateAssetQuote:

			m.mu.RLock()

			// Skip updates from previous asset groups
			if update.VersionVector != m.assetGroupVersionVector {
				m.mu.RUnlock()

				continue
			}
			m.mu.RUnlock()

			// Call the callback function for individual asset quote updates
			go m.onUpdateAssetQuote(update.Data.Symbol, update.Data, update.VersionVector)

		case err := <-m.chanError:
			// Log errors using the configured logger if one is set
			if m.logger != nil {
				m.logger.Printf("%v", err)
			}

		case currencyRates := <-m.chanUpdateCurrencyRates:
			// Set currency rates on each each monitor
			for _, monitor := range m.monitors {
				err := monitor.SetCurrencyRates(currencyRates)
				if err != nil {
					m.chanError <- err
				}
			}

			// Get asset quotes for all sources with new currency rates
			assetGroupQuote := m.GetAssetGroupQuote()

			// Callback with new asset quotes which include the new currency rates
			go m.onUpdateAssetGroupQuote(assetGroupQuote, m.assetGroupVersionVector)
		}
	}
}

// Stop stops all monitors and cancels the context
func (m *Monitor) Stop() {

	for _, monitor := range m.monitors {
		monitor.Stop() //nolint:errcheck
	}

	m.cancel()

}

// getRetryInterval returns how often to retry setting an asset group: the refresh interval, or five seconds when none is set
func getRetryInterval(refreshInterval int) time.Duration {
	if refreshInterval <= 0 {
		return 5 * time.Second
	}

	return time.Duration(refreshInterval) * time.Second
}
