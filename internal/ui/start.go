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
			RefreshInterval: ctx.Config.RefreshInterval,
			TargetCurrency:  ctx.Config.Currency,
			ConfigMonitorsYahoo: mon.ConfigMonitorsYahoo{
				BaseURL:           dep.MonitorYahooBaseURL,
				SessionRootURL:    dep.MonitorYahooSessionRootURL,
				SessionCrumbURL:   dep.MonitorYahooSessionCrumbURL,
				SessionConsentURL: dep.MonitorYahooSessionConsentURL,
			},
			ConfigMonitorPriceCoinbase: mon.ConfigMonitorPriceCoinbase{
				BaseURL:      dep.MonitorPriceCoinbaseBaseURL,
				StreamingURL: dep.MonitorPriceCoinbaseStreamingURL,
			},
		})

		p := tea.NewProgram(
			NewModel(*dep, *ctx, monitors),
			tea.WithMouseCellMotion(),
			tea.WithAltScreen(),
		)

		monitors.SetOnUpdate(mon.ConfigUpdateFns{
			OnUpdateAssetQuote: func(symbol string, assetQuote c.AssetQuote, nonce int) {
				p.Send(SetAssetQuoteMsg{
					symbol:     symbol,
					assetQuote: assetQuote,
					nonce:      nonce,
				})
				return
			},
			OnUpdateAssetGroupQuote: func(assetGroupQuote c.AssetGroupQuote, nonce int) {
				p.Send(SetAssetGroupQuoteMsg{
					assetGroupQuote: assetGroupQuote,
					nonce:           nonce,
				})
				return
			},
		})

		_, err := p.Run()

		return err
	}

}
