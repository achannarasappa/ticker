package ui

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	mon "github.com/achannarasappa/ticker/v4/internal/monitor"
	"github.com/achannarasappa/ticker/v4/internal/ui/component/watchlist"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the command line interface and starts capturing input
func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {

		monitors, _ := mon.NewMonitor(mon.ConfigMonitor{
			ClientHttp: dep.HttpClients.Default,
			Reference:  ctx.Reference,
			Config:     ctx.Config,
		})

		p := tea.NewProgram(
			NewModel(*dep, *ctx, monitors),
			tea.WithMouseCellMotion(),
			tea.WithAltScreen(),
		)

		monitors.SetOnUpdate(func(symbol string, quotePrice c.QuotePrice) {
			p.Send(watchlist.SetAssetQuotePriceMsg{
				Symbol:     symbol,
				QuotePrice: quotePrice,
			})
		})

		_, err := p.Run()

		return err
	}

}
