package watchlist

import (
	"fmt"
	"math"
	"ticker/internal/position"
	"ticker/internal/quote"
	. "ticker/internal/ui/util"

	. "ticker/internal/ui/util/text"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/novalagung/gubrak/v2"
)

var (
	styleNeutral       = NewStyle("#d4d4d4", "", false)
	styleNeutralBold   = NewStyle("#d4d4d4", "", true)
	styleNeutralFaded  = NewStyle("#616161", "", false)
	stylePricePositive = newStyleFromGradient("#C6FF40", "#779929")
	stylePriceNegative = newStyleFromGradient("#FF7940", "#994926")
)

const (
	maxPercentChangeColorGradient = 10
)

type Model struct {
	Width     int
	Quotes    []quote.Quote
	Positions map[string]position.Position
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{
		Width: 100,
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.Width)
	}

	quotes := sortQuotes(m.Quotes)
	items := ""
	for _, quote := range quotes {
		items += "\n" + item(quote, m.Positions[quote.Symbol], m.Width)
	}
	return items
}

func item(q quote.Quote, p position.Position, width int) string {

	return JoinLines(
		Line(
			width,
			Cell{
				Text: styleNeutralBold(q.Symbol),
			},
			Cell{
				Width: 5,
				Text:  marketStateText(q),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  valueText(p.Value),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  styleNeutral(ConvertFloatToString(q.Price) + " " + normalizeCurrency(q.Currency)),
				Align: RightAlign,
			},
		),
		Line(
			width,
			Cell{
				Text: styleNeutralFaded(q.ShortName),
			},
			Cell{
				Width: 25,
				Text:  valueChangeText(p.DayChange, p.DayChangePercent),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  quoteChangeText(q.Change, q.ChangePercent),
				Align: RightAlign,
			},
		),
	)
}

func marketStateText(q quote.Quote) string {
	if q.IsRegularTradingSession {
		return styleNeutralFaded(" ⦿  ")
	}

	if !q.IsRegularTradingSession && q.IsActive {
		return styleNeutralFaded(" ⦾  ")
	}

	return ""
}

func valueText(value float64) string {
	if value <= 0.0 {
		return ""
	}

	return styleNeutral(ConvertFloatToString(value))
}

func valueChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return ""
	}

	return quoteChangeText(change, changePercent)
}

func quoteChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return styleNeutralFaded("  " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return stylePricePositive(changePercent)("↑ " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	return stylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change) + " (" + ConvertFloatToString(changePercent) + "%)")
}

func normalizeCurrency(currency string) string {
	if currency == "GBp" {
		return "GBX"
	}

	return currency
}

func newStyleFromGradient(startColorHex string, endColorHex string) func(float64) func(string) string {
	c1, _ := colorful.Hex(startColorHex)
	c2, _ := colorful.Hex(endColorHex)

	return func(percent float64) func(string) string {
		normalizedPercent := getNormalizedPercentWithMax(percent, maxPercentChangeColorGradient)
		return NewStyle(c1.BlendHsv(c2, normalizedPercent).Hex(), "", false)
	}
}

// Normalize 0-100 percent with a maximum percent value
func getNormalizedPercentWithMax(percent float64, maxPercent float64) float64 {

	absolutePercent := math.Abs(percent)
	if absolutePercent >= maxPercent {
		return 1.0
	}
	return math.Abs(percent / maxPercent)

}

// Sort by change percent and keep all inactive quotes at the end
func sortQuotes(q []quote.Quote) []quote.Quote {
	if len(q) <= 0 {
		return q
	}

	activeQuotes, inactiveQuotes, _ := gubrak.
		From(q).
		Partition(func(v quote.Quote) bool {
			return v.IsActive
		}).
		ResultAndError()

	concatQuotes := gubrak.
		From(activeQuotes).
		OrderBy(func(v quote.Quote) float64 {
			return v.ChangePercent
		}, false).
		Concat(inactiveQuotes).
		Result()

	return (concatQuotes).([]quote.Quote)
}
