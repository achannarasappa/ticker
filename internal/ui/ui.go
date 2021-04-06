package ui

import (
	"fmt"
	"time"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/position"
	"github.com/achannarasappa/ticker/internal/quote"
	"github.com/achannarasappa/ticker/internal/ui/component/summary"
	"github.com/achannarasappa/ticker/internal/ui/component/watchlist"

	. "github.com/achannarasappa/ticker/internal/ui/util"
	. "github.com/achannarasappa/ticker/internal/ui/util/text"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	styleLogo = NewStyle("#ffffd7", "#ff8700", true)
	styleHelp = NewStyle("#4e4e4e", "", true)
)

const (
	footerHeight = 1
)

type Model struct {
	ctx             c.Context
	ready           bool
	headerHeight    int
	getQuotes       func() []quote.Quote
	getPositions    func([]quote.Quote) (map[string]position.Position, position.PositionSummary)
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
		return QuoteMsg{
			quotes: m.getQuotes(),
			time:   getTime(),
		}
	})
}

func NewModel(dep c.Dependencies, ctx c.Context) Model {

	aggregatedLots := position.GetLots(ctx.Config.Lots)
	symbols := position.GetSymbols(ctx.Config, aggregatedLots)

	return Model{
		ctx:             ctx,
		headerHeight:    getVerticalMargin(ctx.Config),
		ready:           false,
		requestInterval: ctx.Config.RefreshInterval,
		getQuotes:       quote.GetQuotes(ctx, *dep.HttpClient, symbols),
		getPositions:    position.GetPositions(ctx, aggregatedLots),
		watchlist:       watchlist.NewModel(ctx),
		summary:         summary.NewModel(ctx),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return QuoteMsg{
			quotes: m.getQuotes(),
			time:   getTime(),
		}
	}
}

type QuoteMsg struct {
	quotes []quote.Quote
	time   string
}

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

	case QuoteMsg:
		positions, positionSummary := m.getPositions(msg.quotes)
		m.watchlist.Quotes = msg.quotes
		m.watchlist.Positions = positions
		m.lastUpdateTime = msg.time
		m.summary.Summary = positionSummary
		if m.ready {
			m.viewport.SetContent(m.watchlist.View())
		}
		return m, m.updateQuotes()

	}

	m.viewport, _ = m.viewport.Update(msg)

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initalizing..."
	}

	viewSummary := ""

	if m.ctx.Config.ShowSummary {
		viewSummary += m.summary.View()
	}

	return viewSummary + "\n" +
		m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime)

}

func footer(width int, time string) string {

	if width < 80 {
		return styleLogo(" ticker ")
	}

	return Line(
		width,
		Cell{
			Width: 10,
			Text:  styleLogo(" ticker "),
		},
		Cell{
			Width: 36,
			Text:  styleHelp("q: exit ↑: scroll up ↓: scroll down"),
		},
		Cell{
			Text:  styleHelp("↻  " + time),
			Align: RightAlign,
		},
	)

}

func getVerticalMargin(config c.Config) int {
	if config.ShowSummary {
		return 2
	}

	return 0
}
