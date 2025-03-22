package monitor

import (
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
	monitorYahoo "github.com/achannarasappa/ticker/v4/internal/monitor/yahoo"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

// Monitor represents a Coinbase market data monitor
type Monitor struct {
	monitors  map[c.QuoteSource]c.Monitor
	chanError chan error
}

type ConfigMonitor struct {
	ClientHttp      *resty.Client
	ClientWebsocket *websocket.Conn
	Reference       c.Reference
	Config          c.Config
}

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

func (m *Monitor) SetOnUpdate(config ConfigUpdateFns) {
	for _, monitor := range m.monitors {
		monitor.SetOnUpdateAssetQuote(config.OnUpdateAssetQuote)
		monitor.SetOnUpdateAssetQuotes(config.OnUpdateAssetQuotes)
	}
}

func (m *Monitor) GetMonitor(source c.QuoteSource) c.Monitor {
	return m.monitors[source]
}

func (m *Monitor) Start() {
	for _, monitor := range m.monitors {
		monitor.Start()
	}
}
