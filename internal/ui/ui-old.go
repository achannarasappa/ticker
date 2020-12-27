package ui

// import (
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	"ticker-tape/internal/quote"
// 	. "ticker-tape/internal/ui/component"

// 	"github.com/charmbracelet/bubbles/viewport"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/muesli/reflow/ansi"
// 	"github.com/muesli/termenv"
// )

// var (
// 	color                   = termenv.ColorProfile().Color
// 	footerHighlightStyle    = termenv.Style{}.Foreground(color("#ffc27d")).Background(color("#f37329")).Bold().Styled
// 	helpStyle               = termenv.Style{}.Foreground(color("241")).Styled
// 	quoteNeutralStyle       = termenv.Style{}.Foreground(color("#d4d4d4")).Bold().Styled
// 	quoteNeutralFadedStyle  = termenv.Style{}.Foreground(color("#7e8087")).Styled
// 	quotePricePositiveStyle = termenv.Style{}.Foreground(color("#d1ff82")).Styled
// 	quotePriceNegativeStyle = termenv.Style{}.Foreground(color("#ff8c82")).Styled
// )

// const (
// 	footerHeight = 1
// )

// type Model struct {
// 	content  string
// 	ready    bool
// 	viewport viewport.Model
// }

// func (m Model) Init() tea.Cmd {
// 	return nil
// }

// func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := message.(type) {

// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c":
// 			fallthrough
// 		case "esc":
// 			fallthrough
// 		case "q":
// 			return m, tea.Quit
// 		}

// 	case tea.WindowSizeMsg:
// 		verticalMargins := footerHeight

// 		if !m.ready {
// 			m.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
// 			m.viewport.SetContent(m.content)
// 			m.ready = true
// 		} else {
// 			m.viewport.Width = msg.Width
// 			m.viewport.Height = msg.Height - verticalMargins
// 		}

// 	}

// 	return m, nil
// }

// func (m Model) View() string {
// 	if !m.ready {
// 		return "\n  Initalizing..."
// 	}

// 	quotes := []quote.Quote{
// 		{Symbol: "AAPL", ShortName: "Apple, Inc.", RegularMarketPrice: 1000.1, RegularMarketChange: 10.1, RegularMarketChangePercent: 1.1},
// 		{Symbol: "ABNB", ShortName: "AirBnB, Inc.", RegularMarketPrice: 645.1, RegularMarketChange: 4.1, RegularMarketChangePercent: 0.9},
// 	}

// 	m.viewport.SetContent(quoteSummaries(quotes, m.viewport.Width))
// 	return fmt.Sprintf("%s\n%s", m.viewport.View(), footer(m.viewport.Width))
// }

// func footer(elementWidth int) string {
// 	return footerHighlightStyle(" 🚀 ticker-tape ") + helpStyle(" q: exit")
// }

// func quoteSummaries(q []quote.Quote, elementWidth int) string {
// 	quoteSummaries := ""
// 	for _, quote := range q {
// 		quoteSummaries = quoteSummaries + "\n\n" + quoteSummary(quote, elementWidth)
// 	}
// 	return quoteSummaries
// }

// func quoteSummary(q quote.Quote, elementWidth int) string {

// 	firstLine := lineWithGap(
// 		quoteNeutralStyle(q.Symbol),
// 		quoteNeutralStyle(convertFloatToString(q.RegularMarketPrice)),
// 		elementWidth,
// 	)
// 	secondLine := lineWithGap(
// 		quoteNeutralFadedStyle(q.ShortName),
// 		quotePricePositiveStyle("↑ "+convertFloatToString(q.RegularMarketChange)+" ("+convertFloatToString(q.RegularMarketChangePercent)+"%)"),
// 		elementWidth,
// 	)

// 	return fmt.Sprintf("%s\n%s", firstLine, secondLine)
// }

// // util
// func lineWithGap(leftText string, rightText string, elementWidth int) string {
// 	innerGapWidth := elementWidth - ansi.PrintableRuneWidth(leftText) - ansi.PrintableRuneWidth(rightText)
// 	return leftText + strings.Repeat(" ", innerGapWidth) + rightText
// }

// func convertFloatToString(f float64) string {
// 	return strconv.FormatFloat(f, 'f', 2, 64)
// }
