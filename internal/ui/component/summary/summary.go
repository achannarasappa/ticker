package summary

import (
	"strings"

	grid "github.com/achannarasappa/term-grid"
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/position"

	. "github.com/achannarasappa/ticker/internal/ui/util"
	"github.com/muesli/reflow/ansi"
)

type Model struct {
	Width   int
	Summary position.PositionSummary
	Context c.Context
	styles  c.Styles
}

// NewModel returns a model with default values.
func NewModel(ctx c.Context) Model {
	return Model{
		Width:  80,
		styles: ctx.Reference.Styles,
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return ""
	}

	textChange := m.styles.TextLabel("Day Change: ") + quoteChangeText(m.Summary.DayChange, m.Summary.DayChangePercent, m.styles) +
		m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Change: ") + quoteChangeText(m.Summary.Change, m.Summary.ChangePercent, m.styles)
	widthChange := ansi.PrintableRuneWidth(textChange)
	textValue := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Value: ") + ValueText(m.Summary.Value, m.styles)
	widthValue := ansi.PrintableRuneWidth(textValue)
	textCost := m.styles.TextLabel(" • ") +
		m.styles.TextLabel("Cost: ") + ValueText(m.Summary.Cost, m.styles)
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
		return styles.TextLabel(ConvertFloatToString(change, false) + " (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+ConvertFloatToString(change, false)+" ("+ConvertFloatToString(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+ConvertFloatToString(change, false)+" ("+ConvertFloatToString(changePercent, false)+"%)")
}
