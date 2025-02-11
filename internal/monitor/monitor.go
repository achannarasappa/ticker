package monitor

import (
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	monitorCoinbase "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

// Monitor represents a Coinbase market data monitor
type Monitor struct {
	monitors map[c.QuoteSource]c.Monitor
}

type ConfigMonitor struct {
	ClientHttp      *resty.Client
	ClientWebsocket *websocket.Conn
	Reference       c.Reference
	Config          c.Config
	OnUpdate        func()
}

// New creates a new instance of the Coinbase monitor
func NewMonitor(configMonitor ConfigMonitor) (*Monitor, error) {

	var coinbase *monitorCoinbase.MonitorCoinbase
	coinbase = monitorCoinbase.NewMonitorCoinbase(
		monitorCoinbase.Config{
			Client:   *configMonitor.ClientHttp,
			OnUpdate: configMonitor.OnUpdate,
		},
		monitorCoinbase.WithSymbolsUnderlying(configMonitor.Reference.SourceToUnderlyingAssetSymbols[c.QuoteSourceCoinbase]),
		monitorCoinbase.WithStreamingURL("ws-feed.exchange.coinbase.com"),
		monitorCoinbase.WithRefreshInterval(time.Duration(configMonitor.Config.RefreshInterval)*time.Second),
	)

	return &Monitor{
		monitors: map[c.QuoteSource]c.Monitor{
			c.QuoteSourceCoinbase: coinbase,
		},
	}, nil
}

// SetSymbols sets the symbols for each monitor
func (m *Monitor) SetSymbols(assetGroup c.AssetGroup) {

	for _, symbolBySource := range assetGroup.SymbolsBySource {

		if monitor, exists := m.monitors[symbolBySource.Source]; exists {
			monitor.SetSymbols(symbolBySource.Symbols)
		}

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
