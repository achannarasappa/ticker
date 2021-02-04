package ui

import (
	"fmt"
	"strings"
	"ticker/internal/cli"
	"ticker/internal/position"
	"ticker/internal/quote"
	"ticker/internal/ui/component/summary"
	"ticker/internal/ui/component/watchlist"
	"time"

	. "ticker/internal/ui/util"
	. "ticker/internal/ui/util/text"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
)

var (
	styleLogo = NewStyle("#ffc27d", "#f37329", true)
	styleHelp = NewStyle("#4e4e4e", "", true)
)

const (
	footerHeight = 1
)

type Model struct {
	ready           bool
	headerHeight    int
	getQuotes       func() []quote.Quote
	getPositions    func([]quote.Quote) map[string]position.Position
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

func NewModel(config cli.Config, client *resty.Client) Model {

	aggregatedLots := position.GetLots(config.Lots)
	symbols := position.GetSymbols(config.Watchlist, aggregatedLots)

	return Model{
		headerHeight:    getVerticalMargin(config),
		ready:           false,
		requestInterval: 3,
		getQuotes:       quote.GetQuotes(*client, symbols),
		getPositions:    position.GetPositions(aggregatedLots),
		watchlist:       watchlist.NewModel(config.Separate, config.ExtraInfoExchange, config.ExtraInfoFundamentals, config.Sort),
		summary:         summary.NewModel(),
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
		positions := m.getPositions(msg.quotes)
		m.watchlist.Quotes = msg.quotes
		m.watchlist.Positions = positions
		m.summary.Summary = position.GetPositionSummary(positions)
		m.lastUpdateTime = msg.time
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

	return strings.Join(
		[]string{
			m.summary.View(),
			m.viewport.View(),
			footer(m.viewport.Width, m.lastUpdateTime),
		},
		"\n",
	)
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
			Text:  styleHelp("⟳  " + time),
			Align: RightAlign,
		},
	)

}

func getVerticalMargin(config cli.Config) int {
	if config.ShowSummary {
		return 2
	}

	return 0
}
