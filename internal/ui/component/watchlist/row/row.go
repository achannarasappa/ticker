package row

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	u "github.com/achannarasappa/ticker/v5/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var lastID int64 //nolint:gochecknoglobals

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
	ID                    int
	Separate              bool
	ShowHoldings          bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Styles                c.Styles
	Asset                 *c.Asset
}

type UpdateAssetMsg *c.Asset

type FrameMsg int

// Model for watchlist row
type Model struct {
	id                   int
	width                int
	config               Config
	cellWidths           CellWidthsContainer
	frame                int
	priceStyle           lipgloss.Style
	priceChangeSegment   string
	priceNoChangeSegment string
	priceChangeDirection int
}

// New returns a model with default values
func New(config Config) *Model {

	var id int

	if config.ID != 0 {
		id = config.ID
	} else {
		id = nextID()
	}

	return &Model{
		id:                   id,
		width:                80,
		config:               config,
		priceNoChangeSegment: u.ConvertFloatToStringWithCommas(config.Asset.QuotePrice.Price, config.Asset.Meta.IsVariablePrecision),
		priceChangeSegment:   "",
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

	case UpdateAssetMsg:

		// If symbol has not changed and price has changed then start the price animation
		if m.config.Asset.Symbol == msg.Symbol && m.config.Asset.QuotePrice.Price != msg.QuotePrice.Price {
			// Reset color and frame on number change
			m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color(""))
			m.frame = 0

			oldPrice := u.ConvertFloatToStringWithCommas(m.config.Asset.QuotePrice.Price, m.config.Asset.Meta.IsVariablePrecision)
			newPrice := u.ConvertFloatToStringWithCommas(msg.QuotePrice.Price, msg.Meta.IsVariablePrecision)

			if msg.QuotePrice.Price > m.config.Asset.QuotePrice.Price {
				m.priceChangeDirection = 1
			} else if msg.QuotePrice.Price < m.config.Asset.QuotePrice.Price {
				m.priceChangeDirection = -1
			}

			// Find the last position where prices differ by iterating from right to left
			if len(oldPrice) == len(newPrice) {
				i := len(newPrice) - 1
				highestIndex := i
				for i >= 0 {
					if newPrice[i] != oldPrice[i] {
						highestIndex = i
					}
					i--
				}

				// Split the price into unchanged and changed segments
				m.priceNoChangeSegment = newPrice[:highestIndex]
				m.priceChangeSegment = newPrice[highestIndex:]
			} else {
				m.priceNoChangeSegment = ""
				m.priceChangeSegment = newPrice
			}

			m.config.Asset = msg

			return m, frameCmd(m.id)

		}

		// If symbol has changed or price has not changed then just update the asset
		m.config.Asset = msg
		m.priceNoChangeSegment = u.ConvertFloatToStringWithCommas(msg.QuotePrice.Price, msg.Meta.IsVariablePrecision)
		m.priceChangeSegment = ""

		return m, nil

	case FrameMsg:

		if m.id != int(msg) {
			return m, nil
		}

		if m.frame < 4 && m.priceChangeDirection > 0 {
			switch m.frame {
			case 0:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("154")).Background(lipgloss.Color("22"))
			case 1:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("154")).Background(lipgloss.Color("22"))
			case 2:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("157")).Background(lipgloss.Color("232"))
			case 3:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color(""))
			}

			m.frame++

			return m, frameCmd(m.id)
		}

		if m.frame < 4 && m.priceChangeDirection < 0 {
			switch m.frame {
			case 0:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.Color("52"))
			case 1:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.Color("52"))
			case 2:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Background(lipgloss.Color("232"))
			case 3:
				m.priceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color(""))
			}

			m.frame++

			return m, frameCmd(m.id)
		}

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
			Cells: m.buildCells(),
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

func (m *Model) buildCells() []grid.Cell {

	if !m.config.ExtraInfoFundamentals && !m.config.ShowHoldings {

		return []grid.Cell{
			{Text: textName(m.config.Asset, m.config.Styles)},
			{Text: textMarketState(m.config.Asset, m.config.Styles), Width: WidthMarketState, Align: grid.Right},
			{Text: textQuote(m.config.Asset, m.config.Styles, m.priceStyle, m.priceNoChangeSegment, m.priceChangeSegment), Width: m.cellWidths.WidthQuote, Align: grid.Right},
		}

	}

	cellName := []grid.Cell{
		{Text: textName(m.config.Asset, m.config.Styles), Width: WidthName},
		{Text: ""},
		{Text: textMarketState(m.config.Asset, m.config.Styles), Width: WidthMarketState, Align: grid.Right},
	}

	cells := []grid.Cell{
		{Text: textQuote(m.config.Asset, m.config.Styles, m.priceStyle, m.priceNoChangeSegment, m.priceChangeSegment), Width: m.cellWidths.WidthQuote, Align: grid.Right},
	}
	widthMinTerm := WidthName + WidthMarketState + m.cellWidths.WidthQuote + (3 * WidthGutter)

	if m.config.ShowHoldings {
		widthHoldings := widthMinTerm + m.cellWidths.WidthPosition + (3 * WidthGutter) + m.cellWidths.WidthPositionExtended + WidthLabel

		cells = append(
			[]grid.Cell{
				{
					Text:            textPositionExtendedLabels(m.config.Asset, m.config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthHoldings,
				},
				{
					Text:            textPositionExtended(m.config.Asset, m.config.Styles),
					Width:           m.cellWidths.WidthPositionExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthPosition + (2 * WidthGutter) + m.cellWidths.WidthPositionExtended,
				},
				{
					Text:            textPosition(m.config.Asset, m.config.Styles),
					Width:           m.cellWidths.WidthPosition,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthPosition + WidthGutter,
				},
			},
			cells...,
		)
		widthMinTerm = widthHoldings
	}

	if m.config.ExtraInfoFundamentals {
		cells = append(
			[]grid.Cell{
				{
					Text:            textVolumeMarketCapLabels(m.config.Asset, m.config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + (6 * WidthGutter) + (3 * WidthLabel) + m.cellWidths.WidthQuoteRange + m.cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textVolumeMarketCap(m.config.Asset),
					Width:           m.cellWidths.WidthVolumeMarketCap,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + (5 * WidthGutter) + (2 * WidthLabel) + m.cellWidths.WidthQuoteRange + m.cellWidths.WidthVolumeMarketCap,
				},
				{
					Text:            textQuoteRangeLabels(m.config.Asset, m.config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + (4 * WidthGutter) + (2 * WidthLabel) + m.cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteRange(m.config.Asset, m.config.Styles),
					Width:           m.cellWidths.WidthQuoteRange,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + (3 * WidthGutter) + WidthLabel + m.cellWidths.WidthQuoteRange,
				},
				{
					Text:            textQuoteExtendedLabels(m.config.Asset, m.config.Styles),
					Width:           WidthLabel,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + (2 * WidthGutter) + WidthLabel,
				},
				{
					Text:            textQuoteExtended(m.config.Asset, m.config.Styles),
					Width:           m.cellWidths.WidthQuoteExtended,
					Align:           grid.Right,
					VisibleMinWidth: widthMinTerm + m.cellWidths.WidthQuoteExtended + WidthGutter,
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

func textQuote(asset *c.Asset, styles c.Styles, priceStyle lipgloss.Style, priceNoChangeSegment string, priceChangeSegment string) string {
	return priceNoChangeSegment + priceStyle.Render(priceChangeSegment) +
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
					u.ConvertFloatToStringWithCommas(asset.Holding.Weight, asset.Meta.IsVariablePrecision)+"%"+
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
		return styles.Text(u.ConvertFloatToStringWithCommas(asset.QuoteFutures.IndexPrice, asset.Meta.IsVariablePrecision)) +
			"\n" +
			styles.Text(u.ConvertFloatToStringWithCommas(asset.QuoteFutures.Basis, false)) + "%"
	}

	if asset.QuotePrice.PriceOpen == 0.0 {
		return styles.Text(u.ConvertFloatToStringWithCommas(asset.QuotePrice.PricePrevClose, asset.Meta.IsVariablePrecision)) +
			"\n"
	}

	return styles.Text(u.ConvertFloatToStringWithCommas(asset.QuotePrice.PricePrevClose, asset.Meta.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToStringWithCommas(asset.QuotePrice.PriceOpen, asset.Meta.IsVariablePrecision))

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

	return styles.Text(u.ConvertFloatToStringWithCommas(asset.Holding.UnitCost, asset.Meta.IsVariablePrecision)) +
		"\n" +
		styles.Text(u.ConvertFloatToStringWithCommas(asset.Holding.Quantity, asset.Meta.IsVariablePrecision))

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
			return u.ConvertFloatToStringWithCommas(asset.QuotePrice.PriceDayLow, asset.Meta.IsVariablePrecision) +
				styles.Text(" - ") +
				u.ConvertFloatToStringWithCommas(asset.QuotePrice.PriceDayHigh, asset.Meta.IsVariablePrecision) +
				"\n" +
				asset.QuoteFutures.Expiry
		}

		return asset.QuoteFutures.Expiry

	}

	if asset.QuotePrice.PriceDayHigh != 0.0 && asset.QuotePrice.PriceDayLow != 0.0 {
		return u.ConvertFloatToStringWithCommas(asset.QuotePrice.PriceDayLow, asset.Meta.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToStringWithCommas(asset.QuotePrice.PriceDayHigh, asset.Meta.IsVariablePrecision) +
			"\n" +
			u.ConvertFloatToStringWithCommas(asset.QuoteExtended.FiftyTwoWeekLow, asset.Meta.IsVariablePrecision) +
			styles.Text(" - ") +
			u.ConvertFloatToStringWithCommas(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision)
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
		return u.ConvertFloatToStringWithCommas(asset.QuoteFutures.OpenInterest, true) +
			"\n" +
			u.ConvertFloatToStringWithCommas(asset.QuoteExtended.Volume, true)
	}

	return u.ConvertFloatToStringWithCommas(asset.QuoteExtended.MarketCap, true) +
		"\n" +
		u.ConvertFloatToStringWithCommas(asset.QuoteExtended.Volume, true)
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
		return styles.TextPrice(changePercent, "  "+u.ConvertFloatToStringWithCommas(change, isVariablePrecision)+" ("+u.ConvertFloatToStringWithCommas(changePercent, false)+"%)")
	}

	if change > 0.0 {
		return styles.TextPrice(changePercent, "↑ "+u.ConvertFloatToStringWithCommas(change, isVariablePrecision)+" ("+u.ConvertFloatToStringWithCommas(changePercent, false)+"%)")
	}

	return styles.TextPrice(changePercent, "↓ "+u.ConvertFloatToStringWithCommas(change, isVariablePrecision)+" ("+u.ConvertFloatToStringWithCommas(changePercent, false)+"%)")
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

func frameCmd(id int) tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return FrameMsg(id)
	})
}

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}
