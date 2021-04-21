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

const (
	WIDTH_MARKET_STATE    = 5
	WIDTH_GUTTER          = 1
	WIDTH_LABEL           = 15
	WIDTH_NAME            = 20
	WIDTH_POSITION_GUTTER = 2
	WIDTH_CHANGE_STATIC   = 12 // "↓ " + " (100.00%)" = 12 length
	WIDTH_RANGE_STATIC    = 3  // " - " = 3 length
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
	cellWidths            CellWidths
}

type CellWidths struct {
	positionLength        int
	quoteLength           int
	WidthQuote            int
	WidthQuoteExtended    int
	WidthQuoteRange       int
	WidthPosition         int
	WidthPositionExtended int
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

	if (m.cellWidths == CellWidths{}) {
		m.cellWidths = getCellWidths(m.Quotes, m.Positions)
	}

	quotes := m.Sorter(m.Quotes, m.Positions)
	rows := make([]grid.Row, 0)
	for _, quote := range quotes {

		position := m.Positions[quote.Symbol]

		rows = append(
			rows,
			grid.Row{
				Width: m.Width,
				Cells: buildCells(quote, position, m.Context.Config, m.styles, m.cellWidths),
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

	return grid.Render(grid.Grid{Rows: rows, GutterHorizontal: WIDTH_GUTTER})
}

func getCellWidths(quotes []Quote, positions map[string]Position) CellWidths {

	cellMaxWidths := CellWidths{}

	for _, quote := range quotes {
		var quoteLength int

		if quote.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(ConvertFloatToString(quote.Price, quote.IsVariablePrecision))
		}

		if quote.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(ConvertFloatToString(quote.FiftyTwoWeekHigh, quote.IsVariablePrecision))
		}

		if quoteLength > cellMaxWidths.quoteLength {
			cellMaxWidths.quoteLength = quoteLength
			cellMaxWidths.WidthQuote = quoteLength + WIDTH_CHANGE_STATIC
			cellMaxWidths.WidthQuoteExtended = quoteLength
			cellMaxWidths.WidthQuoteRange = WIDTH_RANGE_STATIC + (quoteLength * 2)
		}

		if position, ok := positions[quote.Symbol]; ok {
			positionLength := len(ConvertFloatToString(position.Value, quote.IsVariablePrecision))
			positionQuantityLength := len(ConvertFloatToString(position.Quantity, quote.IsVariablePrecision))

			if positionLength > cellMaxWidths.positionLength {
				cellMaxWidths.positionLength = positionLength
				cellMaxWidths.WidthPosition = positionLength + WIDTH_CHANGE_STATIC + WIDTH_POSITION_GUTTER
			}

			if positionLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionLength
			}

			if positionQuantityLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionQuantityLength
			}

		}

	}

	return cellMaxWidths

}

func buildCells(quote Quote, position Position, config c.Config, styles c.Styles, cellWidths CellWidths) []grid.Cell {

	if config.Compact {
		return []grid.Cell{
			{Text: styles.TextBold(quote.Symbol), Width:10, Align: grid.Left},
			{Text: genInstrumentName(quote, styles), Width: 30, Align: grid.Left},
			{Text: textMarketState(quote, styles), Width: 4, Align: grid.Left},
			{Text: genPrice(quote, styles), Width: 10, Align: grid.Left},
			{Text: genPriceChange(quote, styles), Width: 10, Align: grid.Left},
			{Text: genPriceChangePct(quote, styles), Width: 10, Align: grid.Left},
			{Text: ConvertMktcap(quote.MarketCap), Width: 8, Align: grid.Left},
		}
	}
	
	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(quote, styles)},
			{Text: textMarketState(quote, styles), Width: WIDTH_MARKET_STATE, Align: grid.Right},
			{Text: textQuote(quote, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(quote, styles), Width: WIDTH_NAME},
		{Text: ""},
		{Text: textMarketState(quote, styles), Width: WIDTH_MARKET_STATE, Align: grid.Right},
	}

	cells := []grid.Cell{
		{Text: textQuote(quote, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
	}

	widthMinTerm := WIDTH_NAME + WIDTH_MARKET_STATE + cellWidths.WidthQuote + (3 * WIDTH_GUTTER)

	if config.ShowHoldings {
		widthHoldings := widthMinTerm + cellWidths.WidthPosition + (3 * WIDTH_GUTTER) + cellWidths.WidthPositionExtended + WIDTH_LABEL

		cells = append(
			[]grid.Cell{
				{
					Text:            textPositionExtendedLabels(position, styles),
					Width:           WIDTH_LABEL,
					Align:           grid.Right,
					VisibleMinWidth: widthHoldings,
				},
				{
					Text:            textPositionExtended(quote, position, styles),
					Width:           cellWidths.WidthPositionExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + (2 * WIDTH_GUTTER) + cellWidths.WidthPositionExtended,
				},
				{
					Text:            textPosition(quote, position, styles),
					Width:           cellWidths.WidthPosition,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + WIDTH_GUTTER,
				},
			},
			cells...,
		)
		widthMinTerm = widthHoldings
	}

	if config.ExtraInfoFundamentals {
		cells = append(
			[]grid.Cell{
				{
					Text:            textQuoteRangeLabels(quote, styles),
					Width:           WIDTH_LABEL,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (4 * WIDTH_GUTTER) + (2 * WIDTH_LABEL) + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteRange(quote, styles),
					Width:           cellWidths.WidthQuoteRange,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (3 * WIDTH_GUTTER) + WIDTH_LABEL + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteExtendedLabels(quote, styles),
					Width:           WIDTH_LABEL,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (2 * WIDTH_GUTTER) + WIDTH_LABEL,
				},
				{
					Text:            textQuoteExtended(quote, styles),
					Width:           cellWidths.WidthQuoteExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + WIDTH_GUTTER,
				},
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

func genInstrumentName(quote Quote, styles c.Styles) string {
	name := quote.LongName
	if(name=="") {
		name = quote.ShortName
	}

	if len(name) > 30 {
		name = name[:30]
	}

	return styles.TextLabel(name)
}

func genPrice(quote Quote, styles c.Styles) string {
	return styles.Text(ConvertFloatToString(quote.Price, quote.IsVariablePrecision))
}

func genPriceChange(quote Quote, styles c.Styles) string {
	if(quote.Change == 0.0) {
		return ""
	}

	prefix := " "
	if quote.Change<0 {
		prefix = ""
	}

	return prefix + styles.Text(PriceToString(quote.Change))
}

func genPriceChangePct(quote Quote, styles c.Styles) string {
	return styles.Text(ConvertPercent(quote.ChangePercent))
}

func quoteChangeText(change float64, changePercent float64, isVariablePrecision bool, styles c.Styles) string {
	if change == 0.0 {
		return styles.TextPrice(changePercent, "  "+ConvertFloatToString(change, isVariablePrecision)+" ("+ConvertFloatToString(changePercent, false)+"%)")
	}
	
	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+ConvertFloatToString(change, isVariablePrecision)+" ("+ConvertFloatToString(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+ConvertFloatToString(change, isVariablePrecision)+" ("+ConvertFloatToString(changePercent, false)+"%)")
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

	return styles.TextLabel("Prev. Close:") +
		"\n" +
		styles.TextLabel("Open:")
}

func textPositionExtended(quote Quote, position Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.Text(ConvertFloatToString(position.AverageCost, quote.IsVariablePrecision)) +
		"\n" +
		styles.Text(ConvertFloatToString(position.Quantity, quote.IsVariablePrecision))

}

func textPositionExtendedLabels(position Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.TextLabel("Avg. Cost:") +
		"\n" +
		styles.TextLabel("Quantity:")
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
		textDayRange = styles.TextLabel("Day Range:") +
			"\n" +
			styles.TextLabel("52wk Range:")
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
