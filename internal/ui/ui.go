package ui

import (
	"fmt"
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
			quotes: m.requestQuotes(m.symbols),
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

// func temp(symbols []string) []quote.Quote {
// 	return []quote.Quote{
// 		{Symbol: "AAPL", ShortName: "Apple, Inc.", RegularMarketPrice: 1000.1, RegularMarketChange: 10.1, RegularMarketChangePercent: 1.1},
// 		{Symbol: "ABNB", ShortName: "AirBnB, Inc.", RegularMarketPrice: 645.1, RegularMarketChange: 4.1, RegularMarketChangePercent: 0.9},
// 	}
// }

func (m Model) Init() tea.Cmd {
	m.watchlist = watchlist.NewModel()
	return func() tea.Msg {
		return QuoteMsg{
			quotes: m.requestQuotes(m.symbols),
		}
	}
}

type QuoteMsg struct {
	quotes []quote.Quote
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {

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

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initalizing..."
	}

	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer(m.viewport.Width))
}

func footer(elementWidth int) string {
	return footerHighlightStyle(" 🚀 ticker-tape ") + helpStyle(" q: exit")
}
