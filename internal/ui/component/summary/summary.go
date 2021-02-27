package summary

import (
	"strings"

	"github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/ui/util"
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

	return strings.Join([]string{
		StyleNeutralFaded("Day:"),
		quoteChangeText(m.Summary.DayChange, m.Summary.DayChangePercent),
		StyleNeutralFaded("•"),
		StyleNeutralFaded("Change:"),
		quoteChangeText(m.Summary.Change, m.Summary.ChangePercent),
		StyleNeutralFaded("•"),
		StyleNeutralFaded("Value:"),
		ValueText(m.Summary.Value),
	}, " ") + "\n" + StyleLine(strings.Repeat("━", m.Width))

}

func quoteChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return StyleNeutralFaded(ConvertFloatToString(change, false) + " (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return StylePricePositive(changePercent)("↑ " + ConvertFloatToString(change, false) + " (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	return StylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change, false) + " (" + ConvertFloatToString(changePercent, false) + "%)")
}
