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

		monitors.SetOnUpdate(mon.ConfigUpdateFns{
			OnUpdateAsset: func(symbol string, asset c.Asset) {
				p.Send(watchlist.SetAssetMsg{
					Symbol: symbol,
					Asset:  asset,
				})
			},
			OnUpdateAssets: func(assets []c.Asset) {
				p.Send(watchlist.SetAssetsMsg(assets))
			},
		})

		_, err := p.Run()

		return err
	}

}
