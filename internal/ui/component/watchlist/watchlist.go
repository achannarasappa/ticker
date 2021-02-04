package watchlist

import (
	"fmt"
	"strconv"
	"strings"
	"ticker/internal/position"
	"ticker/internal/quote"
	. "ticker/internal/ui/util"

	. "ticker/internal/ui/util/text"

	"github.com/novalagung/gubrak/v2"
)

type Model struct {
	Width                 int
	Quotes                []quote.Quote
	Positions             map[string]position.Position
	Separate              bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sort                  string
}

// NewModel returns a model with default values.
func NewModel(separate bool, extraInfoExchange bool, extraInfoFundamentals bool, sort string) Model {
	return Model{
		Width:                 80,
		Separate:              separate,
		ExtraInfoExchange:     extraInfoExchange,
		ExtraInfoFundamentals: extraInfoFundamentals,
		Sort:                  sort,
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.Width)
	}

	quotes := sortQuotes(m.Quotes, m.Sort)
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
				Text: StyleLine(strings.Repeat("⎯", width)),
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
				Text: StyleNeutralBold(q.Symbol),
			},
			Cell{
				Width: 5,
				Text:  marketStateText(q),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  ValueText(p.Value),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  StyleNeutral(ConvertFloatToString(q.Price)),
				Align: RightAlign,
			},
		),
		Line(
			width,
			Cell{
				Text: StyleNeutralFaded(q.ShortName),
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
			Text:  StyleNeutralFaded("Prev Close: ") + StyleNeutral(ConvertFloatToString(q.RegularMarketPreviousClose)),
		},
		Cell{
			Width: 20,
			Text:  StyleNeutralFaded("Open: ") + StyleNeutral(ConvertFloatToString(q.RegularMarketOpen)),
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
	return StyleNeutralFaded("Day Range: ") + StyleNeutral(dayRange)
}

func exchangeDelayText(delay float64) string {
	if delay <= 0 {
		return "Real-Time"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func tagText(text string) string {
	return StyleTagEnd(" ") + StyleTag(text) + StyleTagEnd(" ")
}

func marketStateText(q quote.Quote) string {
	if q.IsRegularTradingSession {
		return StyleNeutralFaded(" ⦿  ")
	}

	if !q.IsRegularTradingSession && q.IsActive {
		return StyleNeutralFaded(" ⦾  ")
	}

	return ""
}

func valueChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return ""
	}

	return quoteChangeText(change, changePercent)
}

func quoteChangeText(change float64, changePercent float64) string {
	if change == 0.0 {
		return StyleNeutralFaded("  " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return StylePricePositive(changePercent)("↑ " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	return StylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change) + " (" + ConvertFloatToString(changePercent) + "%)")
}

// Sort by `sort` parameter (Symbol or Change Percent).
// Keep all inactive quotes at the end
func sortQuotes(q []quote.Quote, sort string) []quote.Quote {
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
	appendOrderBy(quotesToShow, sort, inactiveQuotes)

	// Get the result from quotes
	concatQuotes := quotesToShow.
		Result()

	return (concatQuotes).([]quote.Quote)
}

func appendOrderBy(quotes gubrak.IChainable, sort string, inactiveQuotes interface{}) {

	switch strings.ToLower(sort) {
	case "alpha":
		quotes.Concat(inactiveQuotes)
		quotes.OrderBy(func(v quote.Quote) string {
			return v.Symbol
		})
	default:
		quotes.OrderBy(func(v quote.Quote) float64 {
			return v.ChangePercent
		}, false)
		quotes.Concat(inactiveQuotes)
	}
}
