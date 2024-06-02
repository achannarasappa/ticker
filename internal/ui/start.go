package ui

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the command line interface and starts capturing input
func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {

		p := tea.NewProgram(
			NewModel(*dep, *ctx),
			tea.WithMouseCellMotion(),
			tea.WithAltScreen(),
		)

		_, err := p.Run()

		return err
	}

}
