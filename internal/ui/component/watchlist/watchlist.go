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

	grid "github.com/achannarasappa/term-grid"
	// . "github.com/achannarasappa/ticker/internal/ui/util/text"
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
	rows := make([]grid.Row, 0)
	for _, quote := range quotes {

		position := m.Positions[quote.Symbol]
		showHoldings := m.Context.Config.ShowHoldings

		rows = append(
			rows,
			grid.Row{
				Width: m.Width,
				Cells: []grid.Cell{
					{Text: textName(quote)},
					{Text: textMarketState(quote), Width: 5, Align: grid.Right},
					{Text: textQuoteRangeLabels(quote), Width: 15, Align: grid.Right, VisibleMinWidth: 135},
					{Text: textQuoteRange(quote), Width: 15, Align: grid.Right, VisibleMinWidth: 120},
					{Text: textQuoteExtendedLabels(quote), Width: 15, Align: grid.Right, VisibleMinWidth: 105},
					{Text: textQuoteExtended(quote), Width: 7, Align: grid.Right, VisibleMinWidth: 90},
					{Text: textPosition(quote, position), Width: 25, Align: grid.Right},
					{Text: textQuote(quote), Width: 25, Align: grid.Right},
				},
			})

		if showHoldings {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: ""},
						{Text: "", Width: 5, Align: grid.Right},
						{Text: StyleNeutralFaded("Avg. Cost: "), Width: 15, Align: grid.Right, VisibleMinWidth: 135},
						{Text: StyleNeutral(ConvertFloatToString(position.AverageCost, quote.IsVariablePrecision)), Width: 15, Align: grid.Right, VisibleMinWidth: 120},
						{Text: StyleNeutralFaded("Quantity: "), Width: 15, Align: grid.Right, VisibleMinWidth: 105},
						{Text: StyleNeutral(ConvertFloatToString(position.Quantity, quote.IsVariablePrecision)), Width: 7, Align: grid.Right, VisibleMinWidth: 90},
						{Text: "", Width: 25, Align: grid.Right},
						{Text: "", Width: 25, Align: grid.Right},
					},
				})
		}

		if m.Separate {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: textSeparator(m.Width)},
					},
				})
		}

	}

	return grid.Render(grid.Grid{Rows: rows, GutterHorizontal: 1})
}

func textName(quote Quote) string {
	return StyleNeutralBold(quote.Symbol) +
		"\n" +
		StyleNeutralFaded(quote.ShortName)
}

func textQuote(quote Quote) string {
	return StyleNeutral(ConvertFloatToString(quote.Price, quote.IsVariablePrecision)) +
		"\n" +
		quoteChangeText(quote.Change, quote.ChangePercent, quote.IsVariablePrecision)
}

func textPosition(quote Quote, position Position) string {

	if position.TotalChange == 0.0 {
		return ""
	}

	return ValueText(position.Value) +
		StyleNeutralLight(
			" ("+
				ConvertFloatToString(position.Weight, quote.IsVariablePrecision)+"%"+
				")") +
		"\n" +
		quoteChangeText(position.TotalChange, position.TotalChangePercent, quote.IsVariablePrecision)
}

func textQuoteExtended(quote Quote) string {

	return StyleNeutral(ConvertFloatToString(quote.PricePrevClose, quote.IsVariablePrecision)) +
		"\n" +
		StyleNeutral(ConvertFloatToString(quote.PriceOpen, quote.IsVariablePrecision))

}

func textQuoteExtendedLabels(quote Quote) string {

	return StyleNeutralFaded("Prev. Close: ") +
		"\n" +
		StyleNeutralFaded("Open: ")
}

func textQuoteRange(quote Quote) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = ConvertFloatToString(quote.PriceDayLow, quote.IsVariablePrecision) +
			StyleNeutral(" - ") +
			ConvertFloatToString(quote.PriceDayHigh, quote.IsVariablePrecision) +
			"\n" +
			ConvertFloatToString(quote.PriceDayLow, quote.IsVariablePrecision) +
			StyleNeutral(" - ") +
			ConvertFloatToString(quote.PriceDayHigh, quote.IsVariablePrecision)
	}

	return textDayRange

}

func textQuoteRangeLabels(quote Quote) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = StyleNeutralFaded("Day Range: ") +
			"\n" +
			StyleNeutralFaded("52wk Range: ")
	}

	return textDayRange
}

func textSeparator(width int) string {
	return StyleLine(strings.Repeat("─", width))
}

// func extraInfoExchange(show bool, q Quote, targetCurrency string, width int) string {
// 	if !show {
// 		return ""
// 	}

// 	currencyText := q.Currency

// 	if targetCurrency != "" && targetCurrency != q.Currency {
// 		currencyText = q.Currency + " → " + targetCurrency
// 	}

// 	return "\n" + Line(
// 		width,
// 		Cell{
// 			Align: RightAlign,
// 			Text:  tagText(currencyText) + " " + tagText(exchangeDelayText(q.ExchangeDelay)) + " " + tagText(q.ExchangeName),
// 		},
// 	)
// }

func exchangeDelayText(delay float64) string {
	if delay <= 0 {
		return "Real-Time"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func tagText(text string) string {
	return StyleTagEnd(" ") + StyleTag(text) + StyleTagEnd(" ")
}

func textMarketState(q Quote) string {
	if q.IsRegularTradingSession {
		return StyleNeutralFaded(" ●  ")
	}

	if !q.IsRegularTradingSession && q.IsActive {
		return StyleNeutralFaded(" ○  ")
	}

	return ""
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
