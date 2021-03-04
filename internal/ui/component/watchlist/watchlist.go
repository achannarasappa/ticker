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

		rows = append(
			rows,
			grid.Row{
				Width: m.Width,
				Cells: buildCells(quote, position, m.Context.Config),
			})

		if m.Context.Config.ExtraInfoExchange {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: textTags(quote)},
					},
				})
		}

		if m.Context.Config.Separate {
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

func buildCells(quote Quote, position Position, config c.Config) []grid.Cell {

	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(quote)},
			{Text: textMarketState(quote), Width: 5, Align: grid.Right},
			{Text: textPosition(quote, position), Width: 25, Align: grid.Right},
			{Text: textQuote(quote), Width: 25, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(quote), Width: 20},
		{Text: ""},
		{Text: textMarketState(quote), Width: 5, Align: grid.Right},
	}

	widthMinTerm := 90
	cells := []grid.Cell{
		{Text: textPosition(quote, position), Width: 25, Align: grid.Right},
		{Text: textQuote(quote), Width: 25, Align: grid.Right},
	}

	if config.ShowHoldings {
		cells = append(
			[]grid.Cell{
				{Text: textPositionExtendedLabels(position), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 15},
				{Text: textPositionExtended(quote, position), Width: 7, Align: grid.Right, VisibleMinWidth: widthMinTerm},
			},
			cells...,
		)
		widthMinTerm += 30
	}

	if config.ExtraInfoFundamentals {
		cells = append(
			[]grid.Cell{
				{Text: textQuoteRangeLabels(quote), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 50},
				{Text: textQuoteRange(quote), Width: 20, Align: grid.Right, VisibleMinWidth: widthMinTerm + 30},
				{Text: textQuoteExtendedLabels(quote), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 15},
				{Text: textQuoteExtended(quote), Width: 7, Align: grid.Right, VisibleMinWidth: widthMinTerm},
			},
			cells...,
		)
	}

	cells = append(
		cellName,
		cells...,
	)

	return cells

}

func textName(quote Quote) string {

	if len(quote.ShortName) > 20 {
		quote.ShortName = quote.ShortName[:20]
	}

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

	positionValue := ""
	positionChange := ""

	if position.Value != 0.0 {
		positionValue = ValueText(position.Value) +
			StyleNeutralLight(
				" ("+
					ConvertFloatToString(position.Weight, quote.IsVariablePrecision)+"%"+
					")")
	}
	if position.TotalChange != 0.0 {
		positionChange = quoteChangeText(position.TotalChange, position.TotalChangePercent, quote.IsVariablePrecision)
	}

	return positionValue +
		"\n" +
		positionChange
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

func textPositionExtended(quote Quote, position Position) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return StyleNeutral(ConvertFloatToString(position.Quantity, quote.IsVariablePrecision)) +
		"\n" +
		StyleNeutral(ConvertFloatToString(position.AverageCost, quote.IsVariablePrecision))

}

func textPositionExtendedLabels(position Position) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return StyleNeutralFaded("Quantity:") +
		"\n" +
		StyleNeutralFaded("Avg. Cost:")
}

func textQuoteRange(quote Quote) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = ConvertFloatToString(quote.PriceDayLow, quote.IsVariablePrecision) +
			StyleNeutral(" - ") +
			ConvertFloatToString(quote.PriceDayHigh, quote.IsVariablePrecision) +
			"\n" +
			ConvertFloatToString(quote.FiftyTwoWeekLow, quote.IsVariablePrecision) +
			StyleNeutral(" - ") +
			ConvertFloatToString(quote.FiftyTwoWeekHigh, quote.IsVariablePrecision)
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

func textTags(q Quote) string {

	currencyText := q.Currency

	if q.CurrencyConverted != "" && q.CurrencyConverted != q.Currency {
		currencyText = q.Currency + " → " + q.CurrencyConverted
	}

	return formatTag(currencyText) + " " + formatTag(exchangeDelayText(q.ExchangeDelay)) + " " + formatTag(q.ExchangeName)
}

func exchangeDelayText(delay float64) string {
	if delay <= 0 {
		return "Real-Time"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func formatTag(text string) string {
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
