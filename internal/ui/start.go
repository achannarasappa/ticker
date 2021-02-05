package ui

import (
	"github.com/achannarasappa/ticker/internal/cli"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
)

func Start(config *cli.Config) func() error {
	return func() error {
		client := resty.New()
		if len(config.Proxy) > 0 {
			client.SetProxy(config.Proxy)
		}
		p := tea.NewProgram(NewModel(*config, client))

		p.EnableMouseCellMotion()
		p.EnterAltScreen()
		err := p.Start()
		p.ExitAltScreen()
		p.DisableMouseCellMotion()

		return err
	}

}
