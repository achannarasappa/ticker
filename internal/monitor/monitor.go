package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
	monitorPriceYahoo "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/monitor-price"
)

// Monitor represents an overall monitor which manages API specific monitors
type Monitor struct {
	monitors                map[c.QuoteSource]c.Monitor
	chanError               chan error
	chanUpdateAssetQuote    chan c.MessageUpdate[c.AssetQuote]
	onUpdateAssetQuote      func(symbol string, assetQuote c.AssetQuote, nonce int)
	onUpdateAssetGroupQuote func(assetGroupQuote c.AssetGroupQuote, nonce int)
	assetGroupNonce         int
	mu                      sync.RWMutex
	errorLogger             *log.Logger
	ctx                     context.Context
	cancel                  context.CancelFunc
}

// ConfigMonitor represents the configuration for the main monitor
type ConfigMonitor struct {
	RefreshInterval int
	ErrorLogger     *log.Logger
}

// ConfigUpdateFns represents the callback functions for when asset quotes are updated
type ConfigUpdateFns struct {
	OnUpdateAssetQuote      func(symbol string, assetQuote c.AssetQuote, nonce int)
	OnUpdateAssetGroupQuote func(assetGroupQuote c.AssetGroupQuote, nonce int)
}

// New creates a new instance of the Coinbase monitor
func NewMonitor(configMonitor ConfigMonitor) (*Monitor, error) {

	chanError := make(chan error, 5)
	chanUpdateAssetQuote := make(chan c.MessageUpdate[c.AssetQuote], 10)

	ctx, cancel := context.WithCancel(context.Background())

	var coinbase *monitorCoinbase.MonitorCoinbase
	coinbase = monitorCoinbase.NewMonitorCoinbase(
		monitorCoinbase.Config{
			Ctx:                  ctx,
			UnaryURL:             "https://api.coinbase.com",
			ChanError:            chanError,
			ChanUpdateAssetQuote: chanUpdateAssetQuote,
		},
		monitorCoinbase.WithStreamingURL("wss://ws-feed.exchange.coinbase.com"),
		monitorCoinbase.WithRefreshInterval(time.Duration(configMonitor.RefreshInterval)*time.Second),
	)

	var yahoo *monitorPriceYahoo.MonitorYahoo
	yahoo = monitorPriceYahoo.NewMonitorYahoo(
		monitorPriceYahoo.Config{
			Ctx:                  ctx,
			UnaryURL:             "https://query1.finance.yahoo.com",
			SessionRootURL:       "https://finance.yahoo.com",
			SessionCrumbURL:      "https://query2.finance.yahoo.com",
			SessionConsentURL:    "https://consent.yahoo.com",
			ChanError:            chanError,
			ChanUpdateAssetQuote: chanUpdateAssetQuote,
		},
		monitorPriceYahoo.WithRefreshInterval(time.Duration(configMonitor.RefreshInterval)*time.Second),
	)

	m := &Monitor{
		monitors: map[c.QuoteSource]c.Monitor{
			c.QuoteSourceCoinbase: coinbase,
			c.QuoteSourceYahoo:    yahoo,
		},
		chanUpdateAssetQuote:    chanUpdateAssetQuote,
		chanError:               chanError,
		onUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, nonce int) {},
		onUpdateAssetQuote:      func(symbol string, assetQuote c.AssetQuote, nonce int) {},
		errorLogger:             configMonitor.ErrorLogger,
		ctx:                     ctx,
		cancel:                  cancel,
	}

	return m, nil
}

// SetAssetGroup sets the asset group for the monitor
func (m *Monitor) SetAssetGroup(assetGroup c.AssetGroup, nonce int) {
	var wg sync.WaitGroup

	// Create a channel for timeout
	done := make(chan bool)

	// Concurrently set symbols for each monitor (execute a synchronous call to update quotes for each monitor)
	for _, symbolBySource := range assetGroup.SymbolsBySource {
		if monitor, exists := m.monitors[symbolBySource.Source]; exists {
			wg.Add(1)
			go func(mon c.Monitor, symbols []string) {
				defer wg.Done()
				err := mon.SetSymbols(symbols, nonce)
				if err != nil {
					m.chanError <- err
				}
				return
			}(monitor, symbolBySource.Symbols)
		}
	}

	// Wait for the waitgroup to finish in the background
	go func() {
		wg.Wait()
		close(done)
	}()

	// Continue when the waitgroup is finished or a timeout is reached
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Emit an error if not all monitors have completed setting symbols but continue
		// The monitors that have not completed may return stale asset quotes
		m.chanError <- fmt.Errorf("timeout waiting for monitors to set symbols on monitors")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update the nonce so that any messages from the previous asset group can be ignored
	m.assetGroupNonce = nonce

	// Get asset quotes for all sources
	assetGroupQuote := m.GetAssetGroupQuote(assetGroup)

	// Run the callback in a goroutine to avoid blocking
	go m.onUpdateAssetGroupQuote(assetGroupQuote, nonce)
}

// SetOnUpdate sets the callback functions for when asset quotes are updated
func (m *Monitor) SetOnUpdate(config ConfigUpdateFns) error {

	if config.OnUpdateAssetQuote == nil || config.OnUpdateAssetGroupQuote == nil {
		return fmt.Errorf("onUpdateAssetQuote and onUpdateAssetGroupQuote must be set")
	}

	m.onUpdateAssetQuote = config.OnUpdateAssetQuote
	m.onUpdateAssetGroupQuote = config.OnUpdateAssetGroupQuote

	return nil
}

// Start starts all monitors
func (m *Monitor) Start() {
	for _, monitor := range m.monitors {
		monitor.Start()
	}

	go m.handleUpdates()
}

// GetAssetGroupQuote synchronously gets price quotes a group of assets across all sources
func (m *Monitor) GetAssetGroupQuote(assetGroup c.AssetGroup, ignoreCache ...bool) c.AssetGroupQuote {

	assetQuotesFromAllSources := make([]c.AssetQuote, 0)

	for _, symbolBySource := range assetGroup.SymbolsBySource {

		assetQuotes, _ := m.monitors[symbolBySource.Source].GetAssetQuotes(ignoreCache...)
		assetQuotesFromAllSources = append(assetQuotesFromAllSources, assetQuotes...)

	}

	return c.AssetGroupQuote{
		AssetQuotes: assetQuotesFromAllSources,
		AssetGroup:  assetGroup,
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
			if update.Nonce != m.assetGroupNonce {
				m.mu.RUnlock()
				continue
			}
			m.mu.RUnlock()

			// Call the callback function for individual asset quote updates
			go m.onUpdateAssetQuote(update.Data.Symbol, update.Data, update.Nonce)

		case err := <-m.chanError:
			// Log errors using the configured logger if one is set
			if m.errorLogger != nil {
				m.errorLogger.Printf("%v", err)
			}
		}
	}
}

// Stop stops all monitors and cancels the context
func (m *Monitor) Stop() {

	for _, monitor := range m.monitors {
		monitor.Stop()
	}

	m.cancel()

}
