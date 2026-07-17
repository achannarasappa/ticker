package poller

import (
	"context"
	"errors"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/achannarasappa/ticker/v5/internal/monitor/coinpaprika/unary"
)

// Poller represents a poller for the CoinPaprika API
type Poller struct {
	refreshInterval      time.Duration
	coinIds              []string
	isStarted            bool
	ctx                  context.Context
	cancel               context.CancelFunc
	unaryAPI             *unary.UnaryAPI
	chanUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
	chanError            chan error
	versionVector        int
}

// PollerConfig represents the configuration for the poller
type PollerConfig struct {
	UnaryAPI             *unary.UnaryAPI
	ChanUpdateAssetQuote chan c.MessageUpdate[c.AssetQuote]
	ChanError            chan error
}

// NewPoller creates a new poller
func NewPoller(ctx context.Context, config PollerConfig) *Poller {
	ctx, cancel := context.WithCancel(ctx)

	return &Poller{
		refreshInterval:      0,
		isStarted:            false,
		ctx:                  ctx,
		cancel:               cancel,
		unaryAPI:             config.UnaryAPI,
		chanUpdateAssetQuote: config.ChanUpdateAssetQuote,
		chanError:            config.ChanError,
		versionVector:        0,
	}
}

// SetSymbols sets the CoinPaprika coin ids to poll
func (p *Poller) SetSymbols(coinIds []string, versionVector int) {
	p.coinIds = coinIds
	p.versionVector = versionVector
}

// SetRefreshInterval sets the refresh interval for the poller
func (p *Poller) SetRefreshInterval(interval time.Duration) error {

	if p.isStarted {
		return errors.New("cannot set refresh interval while poller is started")
	}

	p.refreshInterval = interval

	return nil
}

// Start starts the poller
func (p *Poller) Start() error {
	if p.isStarted {
		return errors.New("poller already started")
	}

	if p.refreshInterval <= 0 {
		return errors.New("refresh interval is not set")
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
				// Skip making a HTTP request if no coin ids are set
				if len(p.coinIds) == 0 {

					continue
				}

				versionVector := p.versionVector

				// Make a HTTP request to get the asset quotes
				assetQuotes, _, err := p.unaryAPI.GetAssetQuotes(p.coinIds)

				if err != nil {
					p.chanError <- err

					continue
				}

				// Send the asset quotes to the update channel
				for _, assetQuote := range assetQuotes {
					p.chanUpdateAssetQuote <- c.MessageUpdate[c.AssetQuote]{
						ID:            assetQuote.Meta.SymbolInSourceAPI,
						Data:          assetQuote,
						VersionVector: versionVector,
					}
				}
			}
		}
	}()

	return nil
}

// Stop stops the poller
func (p *Poller) Stop() error {
	p.cancel()

	return nil
}
