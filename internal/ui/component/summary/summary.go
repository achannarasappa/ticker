package summary

import (
	"strings"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/ui/util"
	"github.com/muesli/reflow/ansi"
)

type Model struct {
	Width   int
	Summary position.PositionSummary
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{
		Width: 80,
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return ""
	}

	textChange := StyleNeutralFaded("Day Change: ") + quoteChangeText(m.Summary.DayChange, m.Summary.DayChangePercent) +
		StyleNeutralFaded(" • ") +
		StyleNeutralFaded("Change: ") + quoteChangeText(m.Summary.Change, m.Summary.ChangePercent)
	widthChange := ansi.PrintableRuneWidth(textChange)
	textValue := StyleNeutralFaded(" • ") +
		StyleNeutralFaded("Value: ") + ValueText(m.Summary.Value)
	widthValue := ansi.PrintableRuneWidth(textValue)
	textCost := StyleNeutralFaded(" • ") +
		StyleNeutralFaded("Cost: ") + ValueText(m.Summary.Cost)
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
					{Text: StyleLine(strings.Repeat("━", m.Width))},
				},
			},
		},
		GutterHorizontal: 1,
	})

}

func quoteChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return StyleNeutralFaded(ConvertFloatToString(change, false) + " (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return StylePrice(changePercent, "↑ "+ConvertFloatToString(change, false)+" ("+ConvertFloatToString(changePercent, false)+"%)")
	}

	return StylePrice(changePercent, "↓ "+ConvertFloatToString(change, false)+" ("+ConvertFloatToString(changePercent, false)+"%)")
}
