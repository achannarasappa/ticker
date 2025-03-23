package ui

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	mon "github.com/achannarasappa/ticker/v4/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the command line interface and starts capturing input
func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {

		monitors, _ := mon.NewMonitor(mon.ConfigMonitor{
			Reference: ctx.Reference,
			Config:    ctx.Config,
		})

		p := tea.NewProgram(
			NewModel(*dep, *ctx, monitors),
			tea.WithMouseCellMotion(),
			tea.WithAltScreen(),
		)

		monitors.SetOnUpdate(mon.ConfigUpdateFns{
			OnUpdateAssetQuote: func(symbol string, assetQuote c.AssetQuote) {
				p.Send(SetAssetQuoteMsg{
					symbol:     symbol,
					assetQuote: assetQuote,
				})
				return
			},
			OnUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote) {
				p.Send(SetAssetGroupQuoteMsg{
					assetGroupQuote: assetGroupQuote,
				})
				return
			},
		})

		_, err := p.Run()

		return err
	}

}
