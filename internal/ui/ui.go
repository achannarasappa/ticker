package ui

import (
	"fmt"
	"sort"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui/component/watchlist"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	color                = termenv.ColorProfile().Color
	footerHighlightStyle = termenv.Style{}.Foreground(color("#ffc27d")).Background(color("#f37329")).Bold().Styled
	helpStyle            = termenv.Style{}.Foreground(color("241")).Styled
)

const (
	footerHeight = 1
)

type Model struct {
	ready           bool
	requestQuotes   func([]string) []quote.Quote
	requestInterval int
	symbols         []string
	viewport        viewport.Model
	watchlist       watchlist.Model
}

func (m Model) updateQuotes() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.requestInterval), func(t time.Time) tea.Msg {
		return QuoteMsg{
			quotes: sortQuotes(m.requestQuotes(m.symbols)),
		}
	})
}

func NewModel(symbols []string, requestQuotes func([]string) []quote.Quote) Model {
	return Model{
		ready:           false,
		requestInterval: 3,
		requestQuotes:   requestQuotes,
		symbols:         symbols,
	}
}

func (m Model) Init() tea.Cmd {
	m.watchlist = watchlist.NewModel()
	return func() tea.Msg {
		return QuoteMsg{
			quotes: sortQuotes(m.requestQuotes(m.symbols)),
		}
	}
}

type QuoteMsg struct {
	quotes []quote.Quote
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
		verticalMargins := footerHeight

		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
			m.watchlist.Width = msg.Width
			m.viewport.SetContent(m.watchlist.View())
			m.ready = true
		} else {
			m.watchlist.Width = msg.Width
			m.viewport.Width = msg.Width
			m.viewport.SetContent(m.watchlist.View())
			m.viewport.Height = msg.Height - verticalMargins
		}

	case QuoteMsg:
		m.watchlist.Quotes = msg.quotes
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

	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer())
}

func footer() string {
	return footerHighlightStyle(" 🚀 ticker-tape ") + helpStyle(" q: exit")
}

func sortQuotes(q []quote.Quote) []quote.Quote {
	sort.Slice(q, func(i, j int) bool { return q[i].RegularMarketChangePercent < q[j].RegularMarketChangePercent })
	return q
}
