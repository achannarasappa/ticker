package watchlist

import (
	"fmt"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
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
	Context               c.Context
}

// NewModel returns a model with default values.
func NewModel(ctx c.Context) Model {
	return Model{
		Width:                 80,
		Context:               ctx,
		Separate:              ctx.Config.Separate,
		ExtraInfoExchange:     ctx.Config.ExtraInfoExchange,
		ExtraInfoFundamentals: ctx.Config.ExtraInfoFundamentals,
		Sorter:                NewSorter(ctx.Config.Sort),
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
					extraInfoHoldings(m.Context.Config.ShowHoldings, quote, m.Positions[quote.Symbol], m.Width),
					extraInfoFundamentals(m.ExtraInfoFundamentals, quote, m.Width),
					extraInfoExchange(m.ExtraInfoExchange, quote, m.Context.Config.Currency, m.Width),
				},
				"",
			),
		)
	}

	return strings.Join(items, separator(m.Separate, m.Width)) + "\n"
}

func separator(isSeparated bool, width int) string {
	if isSeparated {
		return "\n" + Line(
			width,
			Cell{
				Text: StyleLine(strings.Repeat("─", width)),
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
				Text:  StyleNeutral(ConvertFloatToString(q.Price, q.IsVariablePrecision)),
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
				Text:  valueChangeText(p.TotalChange, p.TotalChangePercent, q.IsVariablePrecision),
				Align: RightAlign,
			},
			Cell{
				Width: 25,
				Text:  quoteChangeText(q.Change, q.ChangePercent, q.IsVariablePrecision),
				Align: RightAlign,
			},
		),
	)
}

func extraInfoExchange(show bool, q Quote, targetCurrency string, width int) string {
	if !show {
		return ""
	}

	currencyText := q.Currency

	if targetCurrency != "" && targetCurrency != q.Currency {
		currencyText = q.Currency + " → " + targetCurrency
	}

	return "\n" + Line(
		width,
		Cell{
			Align: RightAlign,
			Text:  tagText(currencyText) + " " + tagText(exchangeDelayText(q.ExchangeDelay)) + " " + tagText(q.ExchangeName),
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
			Text:  dayRangeText(q.PriceDayHigh, q.PriceDayLow, q.IsVariablePrecision),
			Align: RightAlign,
		},
		Cell{
			Width: 15,
			Text:  StyleNeutralFaded("Prev Close: "),
			Align: RightAlign,
		},
		Cell{
			Width: 10,
			Text:  StyleNeutral(ConvertFloatToString(q.PricePrevClose, q.IsVariablePrecision)),
			Align: RightAlign,
		},
		Cell{
			Width: 15,
			Text:  StyleNeutralFaded("Open: "),
			Align: RightAlign,
		},
		Cell{
			Width: 10,
			Text:  StyleNeutral(ConvertFloatToString(q.PriceOpen, q.IsVariablePrecision)),
			Align: RightAlign,
		},
	)
}

func extraInfoHoldings(show bool, q Quote, p Position, width int) string {
	if (p == Position{} || !show) {
		return ""
	}

	return "\n" + Line(
		width,
		Cell{
			Text:  StyleNeutralFaded("Weight: "),
			Align: RightAlign,
		},
		Cell{
			Width: 7,
			Text:  StyleNeutral(ConvertFloatToString(p.Weight, q.IsVariablePrecision)) + "%",
			Align: RightAlign,
		},
		Cell{
			Width: 15,
			Text:  StyleNeutralFaded("Avg. Cost: "),
			Align: RightAlign,
		},
		Cell{
			Width: 10,
			Text:  StyleNeutral(ConvertFloatToString(p.AverageCost, q.IsVariablePrecision)),
			Align: RightAlign,
		},
		Cell{
			Width: 15,
			Text:  StyleNeutralFaded("Quantity: "),
			Align: RightAlign,
		},
		Cell{
			Width: 10,
			Text:  StyleNeutral(ConvertFloatToString(p.Quantity, q.IsVariablePrecision)),
			Align: RightAlign,
		},
	)
}

func dayRangeText(high float64, low float64, isVariablePrecision bool) string {
	if high == 0.0 || low == 0.0 {
		return ""
	}
	return StyleNeutralFaded("Day Range: ") + StyleNeutral(ConvertFloatToString(low, isVariablePrecision)+" - "+ConvertFloatToString(high, isVariablePrecision))
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
		return StyleNeutralFaded(" ●  ")
	}

	if !q.IsRegularTradingSession && q.IsActive {
		return StyleNeutralFaded(" ○  ")
	}

	return ""
}

func valueChangeText(change float64, changePercent float64, isVariablePrecision bool) string {
	if change == 0.0 {
		return ""
	}

	return quoteChangeText(change, changePercent, isVariablePrecision)
}

func quoteChangeText(change float64, changePercent float64, isVariablePrecision bool) string {
	if change == 0.0 {
		return StyleNeutralFaded("  " + ConvertFloatToString(change, isVariablePrecision) + "  (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	if change > 0.0 {
		return StylePricePositive(changePercent)("↑ " + ConvertFloatToString(change, isVariablePrecision) + "  (" + ConvertFloatToString(changePercent, false) + "%)")
	}

	return StylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change, isVariablePrecision) + " (" + ConvertFloatToString(changePercent, false) + "%)")
}
