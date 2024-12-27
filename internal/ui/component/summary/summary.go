package summary

import (
	"strings"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v4/internal/asset"
	c "github.com/achannarasappa/ticker/v4/internal/common"

	u "github.com/achannarasappa/ticker/v4/internal/ui/util"
	"github.com/muesli/reflow/ansi"
)

// Model for summary section
type Model struct {
	Width   int
	Summary asset.HoldingSummary
	Context c.Context
	styles  c.Styles
}

// NewModel returns a model with default values
func NewModel(ctx c.Context) Model {
	return Model{
		Width:  80,
		styles: ctx.Reference.Styles,
	}
}

// View rendering hook for bubbletea
func (m Model) View() string {

	if m.Width < 80 {
		return ""
	}

	textChange := m.styles.TextLabel("Day Change: ") + quoteChangeText(m.Summary.DayChange.Amount, m.Summary.DayChange.Percent, m.styles) +
		m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Change: ") + quoteChangeText(m.Summary.TotalChange.Amount, m.Summary.TotalChange.Percent, m.styles)
	widthChange := ansi.PrintableRuneWidth(textChange)
	textValue := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Value: ") + m.styles.TextLabel(u.ConvertFloatToString(m.Summary.Value, false))
	widthValue := ansi.PrintableRuneWidth(textValue)
	textCost := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Cost: ") + m.styles.TextLabel(u.ConvertFloatToString(m.Summary.Cost, false))
	widthCost := ansi.PrintableRuneWidth(textValue)

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: m.Width,
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
				Width: m.Width,
				Cells: []grid.Cell{
					{Text: m.styles.TextLine(strings.Repeat("━", m.Width))},
				},
			},
		},
		GutterHorizontal: 1,
	})

}

func quoteChangeText(change float64, changePercent float64, styles c.Styles) string {
	if change == 0.0 {
		return styles.TextLabel(u.ConvertFloatToString(change, false) + " (" + u.ConvertFloatToString(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+u.ConvertFloatToString(change, false)+" ("+u.ConvertFloatToString(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+u.ConvertFloatToString(change, false)+" ("+u.ConvertFloatToString(changePercent, false)+"%)")
}
