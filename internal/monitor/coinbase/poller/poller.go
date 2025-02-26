package poller

import (
	"context"
	"fmt"
	"time"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
)

type Poller struct {
	refreshInterval         time.Duration
	symbols                 []string
	isStarted               bool
	ctx                     context.Context
	cancel                  context.CancelFunc
	unaryAPI                *unary.UnaryAPI
	chanUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	chanUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
}

type PollerConfig struct {
	UnaryAPI                *unary.UnaryAPI
	ChanUpdateQuotePrice    chan c.MessageUpdate[c.QuotePrice]
	ChanUpdateQuoteExtended chan c.MessageUpdate[c.QuoteExtended]
}

func NewPoller(ctx context.Context, config PollerConfig) *Poller {
	ctx, cancel := context.WithCancel(ctx)

	return &Poller{
		refreshInterval:         0,
		isStarted:               false,
		ctx:                     ctx,
		cancel:                  cancel,
		unaryAPI:                config.UnaryAPI,
		chanUpdateQuotePrice:    config.ChanUpdateQuotePrice,
		chanUpdateQuoteExtended: config.ChanUpdateQuoteExtended,
	}
}

func (p *Poller) SetSymbols(symbols []string) {
	p.symbols = symbols
}

func (p *Poller) SetRefreshInterval(interval time.Duration) error {

	if p.isStarted {
		return fmt.Errorf("cannot set refresh interval while poller is started")
	}

	p.refreshInterval = interval
	return nil
}

func (p *Poller) Start() error {
	if p.isStarted {
		return fmt.Errorf("poller already started")
	}

	if p.refreshInterval <= 0 {
		return fmt.Errorf("refresh interval is not set")
	}

	p.isStarted = true

	// Start polling goroutine
	go func() {
		ticker := time.NewTicker(p.refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ticker.C:
				if len(p.symbols) == 0 {
					continue
				}
				assetQuotes, err := p.unaryAPI.GetAssetQuotes(p.symbols)
				if err != nil {
					// TODO: send error to error channel
					continue
				}

				for _, assetQuote := range assetQuotes {
					p.chanUpdateQuotePrice <- c.MessageUpdate[c.QuotePrice]{
						ID:   assetQuote.Meta.SymbolInSourceAPI,
						Data: assetQuote.QuotePrice,
					}

					p.chanUpdateQuoteExtended <- c.MessageUpdate[c.QuoteExtended]{
						ID:   assetQuote.Meta.SymbolInSourceAPI,
						Data: assetQuote.QuoteExtended,
					}
				}
			default:
			}
		}
	}()

	return nil
}
