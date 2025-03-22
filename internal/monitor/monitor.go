package monitor

import (
	"fmt"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
	monitorYahoo "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo"
)

// Monitor represents an overall monitor which manages API specific monitors
type Monitor struct {
	monitors  map[c.QuoteSource]c.Monitor
	chanError chan error
}

// ConfigMonitor represents the configuration for the main monitor
type ConfigMonitor struct {
	Reference c.Reference
	Config    c.Config
}

// ConfigUpdateFns represents the callback functions for when asset quotes are updated
type ConfigUpdateFns struct {
	OnUpdateAssetQuote  func(symbol string, assetQuote c.AssetQuote)
	OnUpdateAssetQuotes func(assetQuotes []c.AssetQuote)
}

// New creates a new instance of the Coinbase monitor
func NewMonitor(configMonitor ConfigMonitor) (*Monitor, error) {

	chanError := make(chan error, 5)

	var coinbase *monitorCoinbase.MonitorCoinbase
	coinbase = monitorCoinbase.NewMonitorCoinbase(
		monitorCoinbase.Config{
			UnaryURL:  "https://api.coinbase.com",
			ChanError: chanError,
		},
		monitorCoinbase.WithStreamingURL("wss://ws-feed.exchange.coinbase.com"),
		monitorCoinbase.WithRefreshInterval(time.Duration(configMonitor.Config.RefreshInterval)*time.Second),
	)

	var yahoo *monitorYahoo.MonitorYahoo
	yahoo = monitorYahoo.NewMonitorYahoo(
		monitorYahoo.Config{
			UnaryURL:  "https://query1.finance.yahoo.com",
			ChanError: chanError,
		},
		monitorYahoo.WithRefreshInterval(time.Duration(configMonitor.Config.RefreshInterval)*time.Second),
	)

	m := &Monitor{
		monitors: map[c.QuoteSource]c.Monitor{
			c.QuoteSourceCoinbase: coinbase,
			c.QuoteSourceYahoo:    yahoo,
		},
		chanError: chanError,
	}

	return m, nil
}

// SetSymbols sets the symbols for each monitor
func (m *Monitor) SetSymbols(assetGroup c.AssetGroup) {

	for _, symbolBySource := range assetGroup.SymbolsBySource {

		if monitor, exists := m.monitors[symbolBySource.Source]; exists {
			monitor.SetSymbols(symbolBySource.Symbols)
		}

	}
}

// SetOnUpdate sets the callback functions for when asset quotes are updated
func (m *Monitor) SetOnUpdate(config ConfigUpdateFns) error {

	if config.OnUpdateAssetQuote == nil || config.OnUpdateAssetQuotes == nil {
		return fmt.Errorf("onUpdateAssetQuote and onUpdateAssetQuotes must be set")
	}

	for _, monitor := range m.monitors {
		monitor.SetOnUpdateAssetQuote(config.OnUpdateAssetQuote)
		monitor.SetOnUpdateAssetQuotes(config.OnUpdateAssetQuotes)
	}

	return nil
}

// Start starts all monitors
func (m *Monitor) Start() {
	for _, monitor := range m.monitors {
		monitor.Start()
	}
}

// GetAssetGroupQuote synchronously gets price quotes a group of assets across all sources
func (m *Monitor) GetAssetGroupQuote(assetGroup c.AssetGroup) c.AssetGroupQuote {

	assetQuotesFromAllSources := make([]c.AssetQuote, 0)

	for _, symbolBySource := range assetGroup.SymbolsBySource {

		assetQuotes, _ := m.monitors[symbolBySource.Source].GetAssetQuotes(true)
		assetQuotesFromAllSources = append(assetQuotesFromAllSources, assetQuotes...)

	}

	return c.AssetGroupQuote{
		AssetQuotes: assetQuotesFromAllSources,
		AssetGroup:  assetGroup,
	}
}
