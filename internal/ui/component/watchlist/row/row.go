package row

import (
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	u "github.com/achannarasappa/ticker/v4/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	WidthMarketState    = 5
	WidthGutter         = 1
	WidthLabel          = 15
	WidthName           = 20
	WidthPositionGutter = 2
	WidthChangeStatic   = 12 // "↓ " + " (100.00%)" = 12 length
	WidthRangeStatic    = 3  // " - " = 3 length
)

type SetCellWidthsMsg struct {
	Width      int
	CellWidths CellWidthsContainer
}

type CellWidthsContainer struct {
	PositionLength        int
	QuoteLength           int
	WidthQuote            int
	WidthQuoteExtended    int
	WidthQuoteRange       int
	WidthPosition         int
	WidthPositionExtended int
	WidthVolumeMarketCap  int
}

type Config struct {
	Separate              bool
	ShowHoldings          bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Styles                c.Styles
	Asset                 *c.Asset
}

type UpdateAssetMsg c.Asset

// Model for watchlist row
type Model struct {
	width      int
	config     Config
	cellWidths CellWidthsContainer
}

// New returns a model with default values
func New(config Config) *Model {
	return &Model{
		width:  80,
		config: config,
	}
}

// Init initializes the watchlist row
func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {

	switch msg := msg.(type) {
	case SetCellWidthsMsg:
		m.cellWidths = msg.CellWidths
		m.width = msg.Width
		return m, nil
	}

	return m, nil
}

// View rendering hook
func (m *Model) View() string {

	rows := []grid.Row{}

	rows = append(
		rows,
		grid.Row{
			Width: m.width,
			Cells: buildCells(m.config, m.cellWidths),
		})

	if m.config.ExtraInfoExchange {
		rows = append(
			rows,
			grid.Row{
				Width: m.width,
				Cells: []grid.Cell{
					{Text: textTags(m.config.Asset, m.config.Styles)},
				},
			})
	}

	if m.config.Separate {
		rows = append(
			rows,
			grid.Row{
				Width: m.width,
				Cells: []grid.Cell{
					{Text: textSeparator(m.width, m.config.Styles)},
				},
			})
	}

	return grid.Render(grid.Grid{Rows: rows, GutterHorizontal: WidthGutter})
}

func buildCells(config Config, cellWidths CellWidthsContainer) []grid.Cell {

	if !config.ExtraInfoFundamentals && !config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(config.Asset, config.Styles)},
			{Text: textMarketState(config.Asset, config.Styles), Width: WidthMarketState, Align: grid.Right},
			{Text: textQuote(config.Asset, config.Styles), Width: cellWidths.WidthQuote, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(config.Asset, config.Styles), Width: WidthName},
		{Text: ""},
		{Text: textMarketState(config.Asset, config.Styles), Width: WidthMarketState, Align: grid.Right},
	}

	cells := []grid.Cell{
		{Text: textQuote(config.Asset, config.Styles), Width: cellWidths.WidthQuote, Align: grid.Right},
	}
	widthMinTerm := WidthName + WidthMarketState + cellWidths.WidthQuote + (3 * WidthGutter)

	if config.ShowHoldings {
		widthHoldings := widthMinTerm + cellWidths.WidthPosition + (3 * WidthGutter) + cellWidths.WidthPositionExtended + WidthLabel

		cells = append(
			[]grid.Cell{
				{
					Text:            textPositionExtendedLabels(config.Asset, config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthHoldings,
				},
				{
					Text:            textPositionExtended(config.Asset, config.Styles),
					Width:           cellWidths.WidthPositionExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + (2 * WidthGutter) + cellWidths.WidthPositionExtended,
				},
				{
					Text:            textPosition(config.Asset, config.Styles),
					Width:           cellWidths.WidthPosition,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthPosition + WidthGutter,
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
					Text:            textVolumeMarketCapLabels(config.Asset, config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (6 * WidthGutter) + (3 * WidthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textVolumeMarketCap(config.Asset),
					Width:           cellWidths.WidthVolumeMarketCap,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (5 * WidthGutter) + (2 * WidthLabel) + cellWidths.WidthQuoteRange + cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textQuoteRangeLabels(config.Asset, config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (4 * WidthGutter) + (2 * WidthLabel) + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteRange(config.Asset, config.Styles),
					Width:           cellWidths.WidthQuoteRange,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (3 * WidthGutter) + WidthLabel + cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteExtendedLabels(config.Asset, config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + (2 * WidthGutter) + WidthLabel,
				},
				{
					Text:            textQuoteExtended(config.Asset, config.Styles),
					Width:           cellWidths.WidthQuoteExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + cellWidths.WidthQuoteExtended + WidthGutter,
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

func textName(asset *c.Asset, styles c.Styles) string {

	if len(asset.Name) > 20 {
		asset.Name = asset.Name[:20]
	}

	return styles.TextBold(asset.Symbol) +
		"\n" +
		styles.TextLabel(asset.Name)
}

func textQuote(asset *c.Asset, styles c.Styles) string {
	return styles.Text(u.ConvertFloatToString(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision)) +
		"\n" +
		quoteChangeText(asset.QuotePrice.Change, asset.QuotePrice.ChangePercent, asset.Meta.IsVariablePrecision, styles)
}

func textPosition(asset *c.Asset, styles c.Styles) string {

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

func textQuoteExtended(asset *c.Asset, styles c.Styles) string {

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

func textQuoteExtendedLabels(asset *c.Asset, styles c.Styles) string {

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

func textPositionExtended(asset *c.Asset, styles c.Styles) string {

	if asset.Holding.Quantity == 0.0 {
		return ""
	}

	return styles.Text(u.ConvertFloatToString(asset.Holding.UnitCost, asset.Meta.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToString(asset.Holding.Quantity, asset.Meta.IsVariablePrecision))

}

func textPositionExtendedLabels(asset *c.Asset, styles c.Styles) string {

	if asset.Holding.Quantity == 0.0 {
		return ""
	}

	return styles.TextLabel("Avg. Cost:") +
		"\n" +
		styles.TextLabel("Quantity:")
}

func textQuoteRange(asset *c.Asset, styles c.Styles) string {

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

func textQuoteRangeLabels(asset *c.Asset, styles c.Styles) string {

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

func textVolumeMarketCap(asset *c.Asset) string {

	if asset.Class == c.AssetClassFuturesContract {
		return u.ConvertFloatToString(asset.QuoteFutures.OpenInterest, true) +
			"\n" +
			u.ConvertFloatToString(asset.QuoteExtended.Volume, true)
	}

	return u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true) +
		"\n" +
		u.ConvertFloatToString(asset.QuoteExtended.Volume, true)
}
func textVolumeMarketCapLabels(asset *c.Asset, styles c.Styles) string {

	if asset.Class == c.AssetClassFuturesContract {
		return styles.TextLabel("Open Interest:") +
			"\n" +
			styles.TextLabel("Volume:")
	}

	return styles.TextLabel("Market Cap:") +
		"\n" +
		styles.TextLabel("Volume:")
}

func textMarketState(asset *c.Asset, styles c.Styles) string {
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

func textSeparator(width int, styles c.Styles) string {
	return styles.TextLine(strings.Repeat("─", width))
}

func textTags(asset *c.Asset, styles c.Styles) string {

	currencyText := asset.Currency.FromCurrencyCode

	if asset.Currency.ToCurrencyCode != "" && asset.Currency.ToCurrencyCode != asset.Currency.FromCurrencyCode {
		currencyText = asset.Currency.FromCurrencyCode + " → " + asset.Currency.ToCurrencyCode
	}

	return formatTag(currencyText, styles) + " " + formatTag(exchangeDelayText(asset.Exchange.Delay, asset.Exchange.DelayText), styles) + " " + formatTag(asset.Exchange.Name, styles)
}

func exchangeDelayText(delay float64, delayText string) string {

	if delayText != "" {
		return delayText
	}

	if delay <= 0 {
		return "Live"
	}

	return "Delayed " + strconv.FormatFloat(delay, 'f', 0, 64) + "min"
}

func formatTag(text string, style c.Styles) string {
	return style.Tag(" " + text + " ")
}
