package watchlist

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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
	styleLine          = NewStyle("#3a3a3a", "", false)
	styleTag           = NewStyle("#d4d4d4", "#3a3a3a", false)
	styleTagEnd        = NewStyle("#3a3a3a", "#3a3a3a", false)
	stylePricePositive = newStyleFromGradient("#C6FF40", "#779929")
	stylePriceNegative = newStyleFromGradient("#FF7940", "#994926")
)

const (
	maxPercentChangeColorGradient = 10
)

type Model struct {
	Width                 int
	Quotes                []quote.Quote
	Positions             map[string]position.Position
	Separate              bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	SortQuotesBy          string
}

// NewModel returns a model with default values.
func NewModel(separate bool, extraInfoExchange bool, extraInfoFundamentals bool, sortQuotesBy string) Model {
	return Model{
		Width:                 80,
		Separate:              separate,
		ExtraInfoExchange:     extraInfoExchange,
		ExtraInfoFundamentals: extraInfoFundamentals,
		SortQuotesBy:          sortQuotesBy,
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.Width)
	}

	quotes := sortQuotes(m.Quotes, m.SortQuotesBy)
	items := make([]string, 0)
	for _, quote := range quotes {
		items = append(
			items,
			strings.Join(
				[]string{
					item(quote, m.Positions[quote.Symbol], m.Width),
					extraInfoFundamentals(m.ExtraInfoFundamentals, quote, m.Width),
					extraInfoExchange(m.ExtraInfoExchange, quote, m.Width),
				},
				"",
			),
		)
	}

	return strings.Join(items, separator(m.Separate, m.Width))
}

func separator(isSeparated bool, width int) string {
	if isSeparated {
		return "\n" + Line(
			width,
			Cell{
				Text: styleLine(strings.Repeat("⎯", width)),
			},
		) + "\n"
	}

	return "\n"
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
				Text:  styleNeutral(ConvertFloatToString(q.Price)),
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

func extraInfoExchange(show bool, q quote.Quote, width int) string {
	if !show {
		return ""
	}
	return "\n" + Line(
		width,
		Cell{
			Text: tagText(q.Currency) + " " + tagText(exchangeDelayText(q.ExchangeDelay)) + " " + tagText(q.ExchangeName),
		},
	)
}

func extraInfoFundamentals(show bool, q quote.Quote, width int) string {
	if !show {
		return ""
	}

	return "\n" + Line(
		width,
		Cell{
			Width: 25,
			Text:  styleNeutralFaded("Prev Close: ") + styleNeutral(ConvertFloatToString(q.RegularMarketPreviousClose)),
		},
		Cell{
			Width: 20,
			Text:  styleNeutralFaded("Open: ") + styleNeutral(ConvertFloatToString(q.RegularMarketOpen)),
		},
		Cell{
			Text: dayRangeText(q.RegularMarketDayRange),
		},
	)
}

func dayRangeText(dayRange string) string {
	if len(dayRange) <= 0 {
		return ""
	}
	return styleNeutralFaded("Day Range: ") + styleNeutral(dayRange)
}

func exchangeDelayText(delay float64) string {
	if delay <= 0 {
		return "Real-Time"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func tagText(text string) string {
	return styleTagEnd(" ") + styleTag(text) + styleTagEnd(" ")
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

// Sort by `sortQuotesBy` parameter (Symbol or Change Percent).
// Keep all inactive quotes at the end
func sortQuotes(q []quote.Quote, sortQuotesBy string) []quote.Quote {
	if len(q) <= 0 {
		return q
	}

	activeQuotes, inactiveQuotes, _ := gubrak.
		From(q).
		Partition(func(v quote.Quote) bool {
			return v.IsActive
		}).
		ResultAndError()

	quotesToShow := gubrak.
		From(activeQuotes)

	// Append the orderBy functionality
	appendOrderBy(quotesToShow, sortQuotesBy, inactiveQuotes)

	// Get the result from quotes
	concatQuotes := quotesToShow.
		Result()

	return (concatQuotes).([]quote.Quote)
}

func appendOrderBy(quotes gubrak.IChainable, sortQuotesBy string, inactiveQuotes interface{}) {

	switch strings.ToLower(sortQuotesBy) {
	case "symbol":
		quotes.Concat(inactiveQuotes)
		quotes.OrderBy(func(v quote.Quote) string {
			return v.Symbol
		})
	case "changepercent":
		quotes.OrderBy(func(v quote.Quote) float64 {
			return v.ChangePercent
		}, false)
		quotes.Concat(inactiveQuotes)
	}
}
