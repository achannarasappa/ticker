package ui

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the command line interface and starts capturing input
func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {
		p := tea.NewProgram(NewModel(*dep, *ctx))

		p.EnableMouseCellMotion()
		p.EnterAltScreen()
		err := p.Start()
		p.ExitAltScreen()
		p.DisableMouseCellMotion()

		return err
	}

}
