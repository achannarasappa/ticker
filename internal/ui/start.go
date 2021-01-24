package ui

import (
	"ticker/internal/cli"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
)

func Start(config *cli.Config) func() error {
	return func() error {
		client := resty.New()
		p := tea.NewProgram(NewModel(*config, client))

		p.EnableMouseCellMotion()
		p.EnterAltScreen()
		err := p.Start()
		p.ExitAltScreen()
		p.DisableMouseCellMotion()

		return err
	}

}
