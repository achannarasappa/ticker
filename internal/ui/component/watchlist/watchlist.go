package watchlist

import (
	"fmt"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	p "github.com/achannarasappa/ticker/internal/position"
	q "github.com/achannarasappa/ticker/internal/quote"
	s "github.com/achannarasappa/ticker/internal/sorter"
	u "github.com/achannarasappa/ticker/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
)

const (
	widthMarketState    = 5
	widthGutter         = 1
	widthLabel          = 15
	widthName           = 20
	widthPositionGutter = 2
	widthChangeStatic   = 12 // "↓ " + " (100.00%)" = 12 length
	widthRangeStatic    = 3  // " - " = 3 length
)

// Model for watchlist section
type Model struct {
	Width                 int
	Quotes                []q.Quote
	Positions             map[string]p.Position
	Separate              bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sorter                s.Sorter
	Context               c.Context
	styles                c.Styles
	cellWidths            cellWidthsContainer
}

type cellWidthsContainer struct {
	positionLength        int
	quoteLength           int
	WidthQuote            int
	WidthQuoteExtended    int
	WidthQuoteRange       int
	WidthPosition         int
	WidthPositionExtended int
	WidthVolumeMarketCap  int
}

// NewModel returns a model with default values
func NewModel(ctx c.Context) Model {
	return Model{
		Width:                 80,
		Context:               ctx,
		Separate:              ctx.Config.Separate,
		ExtraInfoExchange:     ctx.Config.ExtraInfoExchange,
		ExtraInfoFundamentals: ctx.Config.ExtraInfoFundamentals,
		Sorter:                s.NewSorter(ctx.Config.Sort),
		styles:                ctx.Reference.Styles,
	}
}

// View rendering hook for bubbletea
func (m Model) View() string {

	if m.Width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.Width)
	}

	if (m.cellWidths == cellWidthsContainer{}) {
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

	return grid.Render(grid.Grid{Rows: rows, GutterHorizontal: widthGutter})
}

func getCellWidths(quotes []q.Quote, positions map[string]p.Position) cellWidthsContainer {

	cellMaxWidths := cellWidthsContainer{}

	for _, quote := range quotes {
		var quoteLength int

		volumeMarketCapLength := len(u.ConvertFloatToString(quote.MarketCap, true))

		if quote.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(u.ConvertFloatToString(quote.Price, quote.IsVariablePrecision))
		}

		if quote.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(u.ConvertFloatToString(quote.FiftyTwoWeekHigh, quote.IsVariablePrecision))
		}

		if volumeMarketCapLength > cellMaxWidths.WidthVolumeMarketCap {
			cellMaxWidths.WidthVolumeMarketCap = volumeMarketCapLength
		}

		if quoteLength > cellMaxWidths.quoteLength {
			cellMaxWidths.quoteLength = quoteLength
			cellMaxWidths.WidthQuote = quoteLength + widthChangeStatic
			cellMaxWidths.WidthQuoteExtended = quoteLength
			cellMaxWidths.WidthQuoteRange = widthRangeStatic + (quoteLength * 2)
		}

		if position, ok := positions[quote.Symbol]; ok {
			positionLength := len(u.ConvertFloatToString(position.Value, quote.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(position.Quantity, quote.IsVariablePrecision))

			if positionLength > cellMaxWidths.positionLength {
				cellMaxWidths.positionLength = positionLength
				cellMaxWidths.WidthPosition = positionLength + widthChangeStatic + widthPositionGutter
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

func buildCells(quote q.Quote, position p.Position, config c.Config, styles c.Styles, cellWidths cellWidthsContainer) []grid.Cell {

	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(quote, styles)},
			{Text: textMarketState(quote, styles), Width: widthMarketState, Align: grid.Right},
			{Text: textQuote(quote, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(quote, styles), Width: widthName},
		{Text: ""},
		{Text: textMarketState(quote, styles), Width: widthMarketState, Align: grid.Right},
	}

	cells := []grid.Cell{
		{Text: textQuote(quote, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
	}

	widthMinTerm := widthName + widthMarketState + cellWidths.WidthQuote + (3 * widthGutter)

	if config.ShowHoldings {
		widthHoldings := widthMinTerm + cellWidths.WidthPosition + (3 * widthGutter) + cellWidths.WidthPositionExtended + widthLabel

		cells = append(
			[]grid.Cell{
				{
					Text:            textPositionExtendedLabels(position, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthHoldings,
				},
				{
					Text:            textPositionExtended(quote, position, styles),
					Width:           cellWidths.WidthPositionExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + (2 * widthGutter) + cellWidths.WidthPositionExtended,
				},
				{
					Text:            textPosition(quote, position, styles),
					Width:           cellWidths.WidthPosition,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + widthGutter,
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
					Text:            textVolumeMarketCapLabels(quote, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (6 * widthGutter) + (3 * widthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textVolumeMarketCap(quote, styles),
					Width:           cellWidths.WidthVolumeMarketCap,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (5 * widthGutter) + (2 * widthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textQuoteRangeLabels(quote, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (4 * widthGutter) + (2 * widthLabel) + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteRange(quote, styles),
					Width:           cellWidths.WidthQuoteRange,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (3 * widthGutter) + widthLabel + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteExtendedLabels(quote, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (2 * widthGutter) + widthLabel,
				},
				{
					Text:            textQuoteExtended(quote, styles),
					Width:           cellWidths.WidthQuoteExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + widthGutter,
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

func textName(quote q.Quote, styles c.Styles) string {

	if len(quote.ShortName) > 20 {
		quote.ShortName = quote.ShortName[:20]
	}

	return styles.TextBold(quote.Symbol) +
		"\n" +
		styles.TextLabel(quote.ShortName)
}

func textQuote(quote q.Quote, styles c.Styles) string {
	return styles.Text(u.ConvertFloatToString(quote.Price, quote.IsVariablePrecision)) +
		"\n" +
		quoteChangeText(quote.Change, quote.ChangePercent, quote.IsVariablePrecision, styles)
}

func textPosition(quote q.Quote, position p.Position, styles c.Styles) string {

	positionValue := ""
	positionChange := ""

	if position.Value != 0.0 {
		positionValue = u.ValueText(position.Value, styles) +
			styles.TextLight(
				" ("+
					u.ConvertFloatToString(position.Weight, quote.IsVariablePrecision)+"%"+
					")")
	}
	if position.TotalChange != 0.0 {
		positionChange = quoteChangeText(position.TotalChange, position.TotalChangePercent, quote.IsVariablePrecision, styles)
	}

	return positionValue +
		"\n" +
		positionChange
}

func textQuoteExtended(quote q.Quote, styles c.Styles) string {

	return styles.Text(u.ConvertFloatToString(quote.PricePrevClose, quote.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToString(quote.PriceOpen, quote.IsVariablePrecision))

}

func textQuoteExtendedLabels(quote q.Quote, styles c.Styles) string {

	return styles.TextLabel("Prev. Close:") +
		"\n" +
		styles.TextLabel("Open:")
}

func textPositionExtended(quote q.Quote, position p.Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.Text(u.ConvertFloatToString(position.AverageCost, quote.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToString(position.Quantity, quote.IsVariablePrecision))

}

func textPositionExtendedLabels(position p.Position, styles c.Styles) string {

	if position.Quantity == 0.0 {
		return ""
	}

	return styles.TextLabel("Avg. Cost:") +
		"\n" +
		styles.TextLabel("Quantity:")
}

func textQuoteRange(quote q.Quote, styles c.Styles) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = u.ConvertFloatToString(quote.PriceDayLow, quote.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToString(quote.PriceDayHigh, quote.IsVariablePrecision) +
			"\n" +
			u.ConvertFloatToString(quote.FiftyTwoWeekLow, quote.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToString(quote.FiftyTwoWeekHigh, quote.IsVariablePrecision)
	}

	return textDayRange

}

func textQuoteRangeLabels(quote q.Quote, styles c.Styles) string {

	textDayRange := ""

	if quote.PriceDayHigh != 0.0 && quote.PriceDayLow != 0.0 {
		textDayRange = styles.TextLabel("Day Range:") +
			"\n" +
			styles.TextLabel("52wk Range:")
	}

	return textDayRange
}

func textVolumeMarketCap(quote q.Quote, styles c.Styles) string {

	return u.ConvertFloatToString(quote.ResponseQuote.MarketCap, true) +
		"\n" +
		u.ConvertFloatToString(quote.ResponseQuote.RegularMarketVolume, true)
}

func textVolumeMarketCapLabels(quote q.Quote, styles c.Styles) string {

	return styles.TextLabel("Market Cap:") +
		"\n" +
		styles.TextLabel("Volume:")
}

func textSeparator(width int, styles c.Styles) string {
	return styles.TextLine(strings.Repeat("─", width))
}

func textTags(quote q.Quote, styles c.Styles) string {

	currencyText := quote.Currency

	if quote.CurrencyConverted != "" && quote.CurrencyConverted != quote.Currency {
		currencyText = quote.Currency + " → " + quote.CurrencyConverted
	}

	return formatTag(currencyText, styles) + " " + formatTag(exchangeDelayText(quote.ExchangeDelay), styles) + " " + formatTag(quote.ExchangeName, styles)
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

func textMarketState(quote q.Quote, styles c.Styles) string {
	if quote.IsRegularTradingSession {
		return styles.TextLabel(" ●  ")
	}

	if !quote.IsRegularTradingSession && quote.IsActive {
		return styles.TextLabel(" ○  ")
	}

	return ""
}

func quoteChangeText(change float64, changePercent float64, isVariablePrecision bool, styles c.Styles) string {
	if change == 0.0 {
		return styles.TextPrice(changePercent, "  "+u.ConvertFloatToString(change, isVariablePrecision)+" ("+u.ConvertFloatToString(changePercent, false)+"%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+u.ConvertFloatToString(change, isVariablePrecision)+" ("+u.ConvertFloatToString(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+u.ConvertFloatToString(change, isVariablePrecision)+" ("+u.ConvertFloatToString(changePercent, false)+"%)")
}
