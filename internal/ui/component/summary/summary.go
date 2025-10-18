package summary

import (
	"strings"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v5/internal/asset"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	tea "github.com/charmbracelet/bubbletea"

	u "github.com/achannarasappa/ticker/v5/internal/ui/util"
	"github.com/muesli/reflow/ansi"
)

// Model for summary section
type Model struct {
	width   int
	summary asset.HoldingSummary
	styles  c.Styles
}

type SetSummaryMsg asset.HoldingSummary

// NewModel returns a model with default values
func NewModel(ctx c.Context) *Model {
	return &Model{
		width:  80,
		styles: ctx.Reference.Styles,
	}
}

// Init initializes the summary component
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the summary component
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

		return m, nil
	case SetSummaryMsg:
		m.summary = asset.HoldingSummary(msg)

		return m, nil
	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {

	if m.width < 80 {
		return ""
	}

	textChange := m.styles.TextLabel("Day Change: ") + quoteChangeText(m.summary.DayChange.Amount, m.summary.DayChange.Percent, m.styles) +
		m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Change: ") + quoteChangeText(m.summary.TotalChange.Amount, m.summary.TotalChange.Percent, m.styles)
	widthChange := ansi.PrintableRuneWidth(textChange)
	textValue := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Value: ") + m.styles.TextLabel(u.ConvertFloatToStringWithCommas(m.summary.Value, false))
	widthValue := ansi.PrintableRuneWidth(textValue)
	textCost := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Cost: ") + m.styles.TextLabel(u.ConvertFloatToStringWithCommas(m.summary.Cost, false))
	widthCost := ansi.PrintableRuneWidth(textValue)

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: m.width,
				Cells: []grid.Cell{
					{
						Text:  textChange,
						Width: widthChange,
					},
					{
						Text:            textValue,
						Width:           widthValue,
						VisibleMinWidth: widthChange + widthValue,
					},
					{
						Text:            textCost,
						Width:           widthCost,
						VisibleMinWidth: widthChange + widthValue + widthCost,
					},
				},
			},
			{
				Width: m.width,
				Cells: []grid.Cell{
					{Text: m.styles.TextLine(strings.Repeat("━", m.width))},
				},
			},
		},
		GutterHorizontal: 1,
	})

}

func quoteChangeText(change float64, changePercent float64, styles c.Styles) string {
	if change == 0.0 {
		return styles.TextLabel(u.ConvertFloatToStringWithCommas(change, false) + " (" + u.ConvertFloatToStringWithCommas(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+u.ConvertFloatToStringWithCommas(change, false)+" ("+u.ConvertFloatToStringWithCommas(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+u.ConvertFloatToStringWithCommas(change, false)+" ("+u.ConvertFloatToStringWithCommas(changePercent, false)+"%)")
}
