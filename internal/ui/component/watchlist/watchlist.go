package watchlist

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"
	. "github.com/achannarasappa/ticker/internal/sorter"
	. "github.com/achannarasappa/ticker/internal/ui/util"

	. "github.com/achannarasappa/ticker/internal/ui/util/text"
)

type Model struct {
	Width                 int
	Quotes                []Quote
	Positions             map[string]Position
	Separate              bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sorter                Sorter
}

// NewModel returns a model with default values.
func NewModel(separate bool, extraInfoExchange bool, extraInfoFundamentals bool, sort string) Model {
	return Model{
		Width:                 80,
		Separate:              separate,
		ExtraInfoExchange:     extraInfoExchange,
		ExtraInfoFundamentals: extraInfoFundamentals,
		Sorter:                NewSorter(sort),
	}
}

func (m Model) View() string {

	if m.Width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.Width)
	}

	quotes := m.Sorter(m.Quotes, m.Positions)
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

func item(q Quote, p Position, width int) string {

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
				Text:  valueChangeText(p.TotalChange, p.TotalChangePercent),
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

func extraInfoExchange(show bool, q Quote, width int) string {
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

func extraInfoFundamentals(show bool, q Quote, width int) string {
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

func marketStateText(q Quote) string {
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
