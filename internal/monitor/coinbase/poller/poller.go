package monitorCoinbase

import (
	"context"
	"fmt"
	"time"

	"github.com/achannarasappa/ticker/v4/internal/monitor/coinbase/unary"
)

type Poller struct {
	refreshInterval time.Duration
	symbols         []string
	isStarted       bool
	ctx             context.Context
	cancel          context.CancelFunc
	unaryAPI        *unary.UnaryAPI
}

func NewPoller(ctx context.Context, unaryAPI *unary.UnaryAPI) *Poller {
	ctx, cancel := context.WithCancel(ctx)

	return &Poller{
		refreshInterval: 0,
		isStarted:       false,
		ctx:             ctx,
		cancel:          cancel,
		unaryAPI:        unaryAPI,
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
			case <-ticker.C:
				if len(p.symbols) == 0 {
					continue
				}
				p.unaryAPI.GetAssetQuotes(p.symbols)
			case <-p.ctx.Done():
				return
			default:
			}
		}
	}()

	return nil
}
