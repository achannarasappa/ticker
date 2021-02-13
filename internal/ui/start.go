package ui

import (
	"github.com/achannarasappa/ticker/internal/cli"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-resty/resty/v2"
)

func Start(client *resty.Client, config *cli.Config, reference cli.Reference) func() error {
	return func() error {
		p := tea.NewProgram(NewModel(*config, client, reference))

		p.EnableMouseCellMotion()
		p.EnterAltScreen()
		err := p.Start()
		p.ExitAltScreen()
		p.DisableMouseCellMotion()

		return err
	}

}
