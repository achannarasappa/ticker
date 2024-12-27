package watchlist

import (
	"fmt"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	s "github.com/achannarasappa/ticker/v4/internal/sorter"
	u "github.com/achannarasappa/ticker/v4/internal/ui/util"

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
	Assets                []c.Asset
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
		m.cellWidths = getCellWidths(m.Assets)
	}

	assets := m.Sorter(m.Assets)
	rows := make([]grid.Row, 0)
	for _, asset := range assets {

		rows = append(
			rows,
			grid.Row{
				Width: m.Width,
				Cells: buildCells(asset, m.Context.Config, m.styles, m.cellWidths),
			})

		if m.Context.Config.ExtraInfoExchange {
			rows = append(
				rows,
				grid.Row{
					Width: m.Width,
					Cells: []grid.Cell{
						{Text: textTags(asset, m.styles)},
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

func getCellWidths(assets []c.Asset) cellWidthsContainer {

	cellMaxWidths := cellWidthsContainer{}

	for _, asset := range assets {
		var quoteLength int

		volumeMarketCapLength := len(u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true))

		if asset.QuoteExtended.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision))
		}

		if asset.QuoteExtended.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision))
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

		if asset.Holding != (c.Holding{}) {
			positionLength := len(u.ConvertFloatToString(asset.Holding.Value, asset.Meta.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(asset.Holding.Quantity, asset.Meta.IsVariablePrecision))

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

func buildCells(asset c.Asset, config c.Config, styles c.Styles, cellWidths cellWidthsContainer) []grid.Cell {

	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(asset, styles)},
			{Text: textMarketState(asset, styles), Width: widthMarketState, Align: grid.Right},
			{Text: textQuote(asset, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(asset, styles), Width: widthName},
		{Text: ""},
		{Text: textMarketState(asset, styles), Width: widthMarketState, Align: grid.Right},
	}

	cells := []grid.Cell{
		{Text: textQuote(asset, styles), Width: cellWidths.WidthQuote, Align: grid.Right},
	}

	widthMinTerm := widthName + widthMarketState + cellWidths.WidthQuote + (3 * widthGutter)

	if config.ShowHoldings {
		widthHoldings := widthMinTerm + cellWidths.WidthPosition + (3 * widthGutter) + cellWidths.WidthPositionExtended + widthLabel

		cells = append(
			[]grid.Cell{
				{
					Text:            textPositionExtendedLabels(asset, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthHoldings,
				},
				{
					Text:            textPositionExtended(asset, styles),
					Width:           cellWidths.WidthPositionExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + (2 * widthGutter) + cellWidths.WidthPositionExtended,
				},
				{
					Text:            textPosition(asset, styles),
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
					Text:            textVolumeMarketCapLabels(asset, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (6 * widthGutter) + (3 * widthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textVolumeMarketCap(asset),
					Width:           cellWidths.WidthVolumeMarketCap,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (5 * widthGutter) + (2 * widthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textQuoteRangeLabels(asset, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (4 * widthGutter) + (2 * widthLabel) + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteRange(asset, styles),
					Width:           cellWidths.WidthQuoteRange,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (3 * widthGutter) + widthLabel + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteExtendedLabels(asset, styles),
					Width:           widthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (2 * widthGutter) + widthLabel,
				},
				{
					Text:            textQuoteExtended(asset, styles),
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

func textName(asset c.Asset, styles c.Styles) string {

	if len(asset.Name) > 20 {
		asset.Name = asset.Name[:20]
	}

	return styles.TextBold(asset.Symbol) +
		"\n" +
		styles.TextLabel(asset.Name)
}

func textQuote(asset c.Asset, styles c.Styles) string {
	return styles.Text(u.ConvertFloatToString(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision)) +
		"\n" +
		quoteChangeText(asset.QuotePrice.Change, asset.QuotePrice.ChangePercent, asset.Meta.IsVariablePrecision, styles)
}

func textPosition(asset c.Asset, styles c.Styles) string {

	positionValue := ""
	positionChange := ""

	if asset.Holding.Value != 0.0 {
		positionValue = u.ValueText(asset.Holding.Value, styles) +
			styles.TextLight(
				" ("+
					u.ConvertFloatToString(asset.Holding.Weight, asset.Meta.IsVariablePrecision)+"%"+
					")")
	}
	if asset.Holding.TotalChange.Amount != 0.0 {
		positionChange = quoteChangeText(asset.Holding.TotalChange.Amount, asset.Holding.TotalChange.Percent, asset.Meta.IsVariablePrecision, styles)
	}

	return positionValue +
		"\n" +
		positionChange
}

func textQuoteExtended(asset c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract && asset.QuoteFutures.IndexPrice == 0.0 {
		return ""
	}

	if asset.Class == c.AssetClassFuturesContract {
		return styles.Text(u.ConvertFloatToString(asset.QuoteFutures.IndexPrice, asset.Meta.IsVariablePrecision)) +
			"\n" +
			styles.Text(u.ConvertFloatToString(asset.QuoteFutures.Basis, false)) + "%"
	}

	if asset.QuotePrice.PriceOpen == 0.0 {
		return styles.Text(u.ConvertFloatToString(asset.QuotePrice.PricePrevClose, asset.Meta.IsVariablePrecision)) +
			"\n"
	}

	return styles.Text(u.ConvertFloatToString(asset.QuotePrice.PricePrevClose, asset.Meta.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToString(asset.QuotePrice.PriceOpen, asset.Meta.IsVariablePrecision))

}

func textQuoteExtendedLabels(asset c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract && asset.QuoteFutures.IndexPrice == 0.0 {
		return ""
	}

	if asset.Class == c.AssetClassFuturesContract {
		return styles.TextLabel("Index Price:") +
			"\n" +
			styles.TextLabel("Basis:")
	}

	if asset.QuotePrice.PriceOpen == 0.0 {
		return styles.TextLabel("Prev. Close:") +
			"\n"
	}

	return styles.TextLabel("Prev. Close:") +
		"\n" +
		styles.TextLabel("Open:")
}

func textPositionExtended(asset c.Asset, styles c.Styles) string {

	if asset.Holding.Quantity == 0.0 {
		return ""
	}

	return styles.Text(u.ConvertFloatToString(asset.Holding.UnitCost, asset.Meta.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToString(asset.Holding.Quantity, asset.Meta.IsVariablePrecision))

}

func textPositionExtendedLabels(asset c.Asset, styles c.Styles) string {

	if asset.Holding.Quantity == 0.0 {
		return ""
	}

	return styles.TextLabel("Avg. Cost:") +
		"\n" +
		styles.TextLabel("Quantity:")
}

func textQuoteRange(asset c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract {

		if asset.QuotePrice.PriceDayHigh != 0.0 && asset.QuotePrice.PriceDayLow != 0.0 {
			return u.ConvertFloatToString(asset.QuotePrice.PriceDayLow, asset.Meta.IsVariablePrecision) +
				styles.Text(" - ") +
				u.ConvertFloatToString(asset.QuotePrice.PriceDayHigh, asset.Meta.IsVariablePrecision) +
				"\n" +
				asset.QuoteFutures.Expiry
		}

		return asset.QuoteFutures.Expiry

	}

	if asset.QuotePrice.PriceDayHigh != 0.0 && asset.QuotePrice.PriceDayLow != 0.0 {
		return u.ConvertFloatToString(asset.QuotePrice.PriceDayLow, asset.Meta.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToString(asset.QuotePrice.PriceDayHigh, asset.Meta.IsVariablePrecision) +
			"\n" +
			u.ConvertFloatToString(asset.QuoteExtended.FiftyTwoWeekLow, asset.Meta.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToString(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision)
	}

	return ""

}

func textQuoteRangeLabels(asset c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract {

		if asset.QuotePrice.PriceDayHigh != 0.0 && asset.QuotePrice.PriceDayLow != 0.0 {
			return styles.TextLabel("Day Range:") +
				"\n" +
				styles.TextLabel("Expiry:")
		}

		return styles.TextLabel("Expiry:")
	}

	if asset.QuotePrice.PriceDayHigh != 0.0 && asset.QuotePrice.PriceDayLow != 0.0 {
		return styles.TextLabel("Day Range:") +
			"\n" +
			styles.TextLabel("52wk Range:")
	}

	return ""
}

func textVolumeMarketCap(asset c.Asset) string {

	if asset.Class == c.AssetClassFuturesContract {
		return u.ConvertFloatToString(asset.QuoteFutures.OpenInterest, true) +
			"\n" +
			u.ConvertFloatToString(asset.QuoteExtended.Volume, true)
	}

	return u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true) +
		"\n" +
		u.ConvertFloatToString(asset.QuoteExtended.Volume, true)
}

func textVolumeMarketCapLabels(asset c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract {
		return styles.TextLabel("Open Interest:") +
			"\n" +
			styles.TextLabel("Volume:")
	}

	return styles.TextLabel("Market Cap:") +
		"\n" +
		styles.TextLabel("Volume:")
}

func textSeparator(width int, styles c.Styles) string {
	return styles.TextLine(strings.Repeat("─", width))
}

func textTags(asset c.Asset, styles c.Styles) string {

	currencyText := asset.Currency.FromCurrencyCode

	if asset.Currency.ToCurrencyCode != "" && asset.Currency.ToCurrencyCode != asset.Currency.FromCurrencyCode {
		currencyText = asset.Currency.FromCurrencyCode + " → " + asset.Currency.ToCurrencyCode
	}

	return formatTag(currencyText, styles) + " " + formatTag(exchangeDelayText(asset.Exchange.Delay), styles) + " " + formatTag(asset.Exchange.Name, styles)
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

func textMarketState(asset c.Asset, styles c.Styles) string {
	if asset.Exchange.IsRegularTradingSession {
		return styles.TextLabel(" ●  ")
	}

	if !asset.Exchange.IsRegularTradingSession && asset.Exchange.IsActive {
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
