package ui

import (
	c "github.com/achannarasappa/ticker/internal/common"
	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the command line interface and starts capturing input
func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {
		opts := []tea.ProgramOption{
			tea.WithMouseCellMotion(),
			tea.WithAltScreen(),
		}
		p := tea.NewProgram(NewModel(*dep, *ctx), opts...)
		_, err := p.Run()

		return err
	}

}
