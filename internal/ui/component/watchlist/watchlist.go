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
	styles                c.Styles
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
		styles:                ctx.Reference.Styles,
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
				Cells: buildCells(quote, position, m.Context.Config, m.styles),
			})

		if m.Context.Config.ExtraInfoExchange {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: textTags(quote, m.styles)},
					},
				})
		}

		if m.Context.Config.Separate {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: textSeparator(m.Width, m.styles)},
					},
				})
		}

	}

	return grid.Render(grid.Grid{Rows: rows, GutterHorizontal: 1})
}

func buildCells(quote Quote, position Position, config c.Config, styles c.Styles) []grid.Cell {

	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(quote, styles)},
			{Text: textMarketState(quote, styles), Width: 5, Align: grid.Right},
			{Text: textPosition(quote, position, styles), Width: 25, Align: grid.Right},
			{Text: textQuote(quote, styles), Width: 25, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(quote, styles), Width: 20},
		{Text: ""},
		{Text: textMarketState(quote, styles), Width: 5, Align: grid.Right},
	}

	widthMinTerm := 90
	cells := []grid.Cell{
		{Text: textPosition(quote, position, styles), Width: 25, Align: grid.Right},
		{Text: textQuote(quote, styles), Width: 25, Align: grid.Right},
	}

	if config.ShowHoldings {
		cells = append(
			[]grid.Cell{
				{Text: textPositionExtendedLabels(position, styles), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 15},
				{Text: textPositionExtended(quote, position, styles), Width: 10, Align: grid.Right, VisibleMinWidth: widthMinTerm},
			},
			cells...,
		)
		widthMinTerm += 30
	}

	if config.ExtraInfoFundamentals {
		cells = append(
			[]grid.Cell{
				{Text: textQuoteRangeLabels(quote, styles), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 50},
				{Text: textQuoteRange(quote, styles), Width: 20, Align: grid.Right, VisibleMinWidth: widthMinTerm + 30},
				{Text: textQuoteExtendedLabels(quote, styles), Width: 15, Align: grid.Right, VisibleMinWidth: widthMinTerm + 15},
				{Text: textQuoteExtended(quote, styles), Width: 7, Align: grid.Right, VisibleMinWidth: widthMinTerm},
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

func textName(quote Quote, styles c.Styles) string {

	if len(quote.ShortName) > 20 {
		quote.ShortName = quote.ShortName[:20]
	}

	return styles.TextBold(quote.Symbol) +
		"\n" +
		styles.TextLabel(quote.ShortName)
}

func textQuote(quote Quote, styles c.Styles) string {
	return styles.Text(ConvertFloatToString(quote.Price, quote.IsVariablePrecision)) +
		"\n" +
		quoteChangeText(quote.Change, quote.ChangePercent, quote.IsVariablePrecision, styles)
}

func textPosition(quote Quote, position Position, styles c.Styles) string {

	positionValue := ""
	positionChange := ""

	if position.Value != 0.0 {
		positionValue = ValueText(position.Value, styles) +
			styles.TextLight(
				" ("+
					ConvertFloatToString(position.Weight, quote.IsVariablePrecision)+"%"+
					")")
	}
	if position.TotalChange != 0.0 {
		positionChange = quoteChangeText(position.TotalChange, position.TotalChangePercent, quote.IsVariablePrecision, styles)
	}

	return positionValue +
		"\n" +
		positionChange
}

func textQuoteExtended(quote Quote, styles c.Styles) string {

	return styles.Text(ConvertFloatToString(quote.PricePrevClose, quote.IsVariablePrecision)) +
		"\n" +
		styles.Text(ConvertFloatToString(quote.PriceOpen, quote.IsVariablePrecision))

}

func textQuoteExtendedLabels(quote Quote, styles c.Styles) string {

	return styles.TextLabel("Prev. Close: ") +
		"\n" +
		styles.TextLabel("Open: ")
}

func textPositionExtended(quote Quote, position Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.Text(ConvertFloatToString(position.Quantity, quote.IsVariablePrecision)) +
		"\n" +
		styles.Text(ConvertFloatToString(position.AverageCost, quote.IsVariablePrecision))

}

func textPositionExtendedLabels(position Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.TextLabel("Quantity:") +
		"\n" +
		styles.TextLabel("Avg. Cost:")
}

func textQuoteRange(quote Quote, styles c.Styles) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = ConvertFloatToString(quote.PriceDayLow, quote.IsVariablePrecision) +
			styles.Text(" - ") +
			ConvertFloatToString(quote.PriceDayHigh, quote.IsVariablePrecision) +
			"\n" +
			ConvertFloatToString(quote.FiftyTwoWeekLow, quote.IsVariablePrecision) +
			styles.Text(" - ") +
			ConvertFloatToString(quote.FiftyTwoWeekHigh, quote.IsVariablePrecision)
	}

	return textDayRange

}

func textQuoteRangeLabels(quote Quote, styles c.Styles) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = styles.TextLabel("Day Range: ") +
			"\n" +
			styles.TextLabel("52wk Range: ")
	}

	return textDayRange
}

func textSeparator(width int, styles c.Styles) string {
	return styles.TextLine(strings.Repeat("─", width))
}

func textTags(q Quote, styles c.Styles) string {

	currencyText := q.Currency

	if q.CurrencyConverted != "" && q.CurrencyConverted != q.Currency {
		currencyText = q.Currency + " → " + q.CurrencyConverted
	}

	return formatTag(currencyText, styles) + " " + formatTag(exchangeDelayText(q.ExchangeDelay), styles) + " " + formatTag(q.ExchangeName, styles)
}

func exchangeDelayText(delay float64) string {
	if delay <= 0 {
		return "Real-Time"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func formatTag(text string, style c.Styles) string {
	return style.Tag(" " + text + " ")
}

func textMarketState(q Quote, styles c.Styles) string {
	if q.IsRegularTradingSession {
		return styles.TextLabel(" ●  ")
	}

	if !q.IsRegularTradingSession && q.IsActive {
		return styles.TextLabel(" ○  ")
	}

	return ""
}

func quoteChangeText(change float64, changePercent float64, isVariablePrecision bool, styles c.Styles) string {
	if change == 0.0 {
		return styles.TextPrice(changePercent, "  "+ConvertFloatToString(change, isVariablePrecision)+"  ("+ConvertFloatToString(changePercent, false)+"%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+ConvertFloatToString(change, isVariablePrecision)+"  ("+ConvertFloatToString(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+ConvertFloatToString(change, isVariablePrecision)+" ("+ConvertFloatToString(changePercent, false)+"%)")
}
