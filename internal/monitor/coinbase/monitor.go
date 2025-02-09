package monitorCoinbase

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/achannarasappa/ticker/v4/internal/common"
	poller "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/poller"
	streamer "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/streamer"
	unary "github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
	resty "github.com/go-resty/resty/v2"
)

type MonitorCoinbase struct {
	unaryAPI    *unary.UnaryAPI
	streamer    *streamer.Streamer
	poller      *poller.Poller
	unary       *unary.UnaryAPI
	input       input
	symbols     []string
	assetQuotes []common.AssetQuote
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

type unaryAPI struct {
	symbols []string
	client  resty.Client
}

type input struct {
	symbols           []string
	symbolsUnderlying []string
}

// Config contains the required configuration for the Coinbase monitor
type Config struct {
	Client   resty.Client
	OnUpdate func()
}

// Option defines an option for configuring the monitor
type Option func(*MonitorCoinbase)

func NewMonitorCoinbase(config Config, opts ...Option) *MonitorCoinbase {

	ctx, cancel := context.WithCancel(context.Background())

	unaryAPI := unary.NewUnaryAPI(config.Client)

	monitor := &MonitorCoinbase{
		streamer: streamer.NewStreamer(ctx),
		poller:   poller.NewPoller(ctx, unaryAPI),
		unaryAPI: unaryAPI,
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range opts {
		opt(monitor)
	}

	return monitor
}

// WithSymbolsUnderlying sets the underlying symbols for the monitor
func WithSymbolsUnderlying(symbols []string) Option {
	return func(m *MonitorCoinbase) {
		m.input.symbolsUnderlying = symbols
	}
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

func (m *MonitorCoinbase) GetAssetQuotes(ignoreCache ...bool) []common.AssetQuote {

	if len(ignoreCache) > 0 && ignoreCache[0] {
		return m.unaryAPI.GetAssetQuotes(m.symbols).AssetQuotes
	}

	return m.assetQuotes
}

func (m *MonitorCoinbase) SetSymbols(symbols []string) {

	symbolsMerged := make([]string, 0)
	symbolsMerged = append(symbolsMerged, m.input.symbolsUnderlying...)
	symbolsMerged = append(symbolsMerged, symbols...)
	slices.Sort(symbolsMerged)
	symbolsMerged = slices.Compact(symbolsMerged)

	m.input.symbols = symbols
	m.symbols = symbolsMerged

	// Execute one unary API call to get data not sent by streaming API and set initial prices
	assetQuotesIndexed := m.unaryAPI.GetAssetQuotes(m.symbols) // TODO: update to return and handle error
	m.assetQuotes = assetQuotesIndexed.AssetQuotes             // TODO: Add saving indexed quotes

	// Coinbase steaming API for CBE (spot) only and not CDE (futures)
	m.streamer.SetSymbolsAndUpdateSubscriptions(symbols) // TODO: update to return and handle error

}

// Start the monitor
func (m *MonitorCoinbase) Start() error {

	var err error

	err = m.streamer.Start()
	if err != nil {
		return err
	}

	err = m.poller.Start()
	if err != nil {
		return err
	}

	return nil
}

func (m *MonitorCoinbase) Stop() error {

	m.cancel()
	return nil
}
