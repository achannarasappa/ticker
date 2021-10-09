package ui

import (
	"fmt"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/internal/asset"
	c "github.com/achannarasappa/ticker/internal/common"
	quote "github.com/achannarasappa/ticker/internal/quote/yahoo"
	"github.com/achannarasappa/ticker/internal/ui/component/summary"
	"github.com/achannarasappa/ticker/internal/ui/component/watchlist"

	util "github.com/achannarasappa/ticker/internal/ui/util"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	styleLogo = util.NewStyle("#ffffd7", "#ff8700", true)
	styleHelp = util.NewStyle("#4e4e4e", "", true)
)

const (
	footerHeight = 1
)

// Model for UI
type Model struct {
	ctx             c.Context
	ready           bool
	headerHeight    int
	getQuotes       func() []c.AssetQuote
	requestInterval int
	viewport        viewport.Model
	watchlist       watchlist.Model
	summary         summary.Model
	lastUpdateTime  string
}

func getTime() string {
	t := time.Now()
	return fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())
}

func (m Model) updateQuotes() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.requestInterval), func(t time.Time) tea.Msg {
		return quoteMsg{
			quotes: m.getQuotes(),
			time:   getTime(),
		}
	})
}

// NewModel is the constructor for UI model
func NewModel(dep c.Dependencies, ctx c.Context) Model {

	symbols := asset.GetSymbols(ctx.Config)

	return Model{
		ctx:             ctx,
		headerHeight:    getVerticalMargin(ctx.Config),
		ready:           false,
		requestInterval: ctx.Config.RefreshInterval,
		getQuotes:       quote.GetAssetQuotes(*dep.HttpClient, symbols),
		watchlist:       watchlist.NewModel(ctx),
		summary:         summary.NewModel(ctx),
	}
}

// Init is the initialization hook for bubbletea
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return quoteMsg{
			quotes: m.getQuotes(),
			time:   getTime(),
		}
	}
}

type quoteMsg struct {
	quotes []c.AssetQuote
	time   string
}

// Update hook for bubbletea
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, tea.Quit
		case "left", "h":
			m.watchlist.TabIndex -= 1
			if m.watchlist.TabIndex < 0 {
				m.watchlist.TabIndex = len(m.ctx.Config.Groups) - 1
			}
			m.viewport.SetContent(m.watchlist.View())
		case "right", "l":
			m.watchlist.TabIndex = (m.watchlist.TabIndex + 1) % len(m.ctx.Config.Groups)
			m.viewport.SetContent(m.watchlist.View())
		}

	case tea.WindowSizeMsg:
		m.watchlist.Width = msg.Width
		m.summary.Width = msg.Width
		viewportHeight := msg.Height - m.headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: viewportHeight}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		m.viewport.SetContent(m.watchlist.View())

	case quoteMsg:
		assets, holdingSummary := asset.GetAssets(m.ctx, msg.quotes)
		m.watchlist.Assets = assets
		m.lastUpdateTime = msg.time
		m.summary.Summary = holdingSummary
		if m.ready {
			m.viewport.SetContent(m.watchlist.View())
		}
		return m, m.updateQuotes()

	}

	m.viewport, _ = m.viewport.Update(msg)

	return m, nil
}

// View rendering hook for bubbletea
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	viewSummary := ""

	if m.ctx.Config.ShowSummary && m.ctx.Config.ShowHoldings {
		viewSummary += m.summary.View() + "\n"
	}

	return viewSummary + m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime)

}

func footer(width int, time string) string {

	if width < 80 {
		return styleLogo(" ticker ")
	}

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: width,
				Cells: []grid.Cell{
					{Text: styleLogo(" ticker "), Width: 9},
					{Text: styleHelp("q: exit ↑: scroll up ↓: scroll down"), Width: 35},
					{Text: styleHelp("↻  " + time), Align: grid.Right},
				},
			},
		},
	})

}

func getVerticalMargin(config c.Config) int {
	if config.ShowSummary && config.ShowHoldings {
		return 2
	}

	return 0
}
