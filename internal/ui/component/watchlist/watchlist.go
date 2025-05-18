package watchlist

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	s "github.com/achannarasappa/ticker/v4/internal/sorter"
	row "github.com/achannarasappa/ticker/v4/internal/ui/component/watchlist/row"
	u "github.com/achannarasappa/ticker/v4/internal/ui/util"

	tea "github.com/charmbracelet/bubbletea"
)

// Config represents the configuration for the watchlist component
type Config struct {
	Separate              bool
	ShowHoldings          bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sort                  string
	Styles                c.Styles
}

// Model for watchlist section
type Model struct {
	width          int
	assets         []*c.Asset
	assetsBySymbol map[string]*c.Asset
	sorter         s.Sorter
	config         Config
	cellWidths     row.CellWidthsContainer
	rows           []*row.Model
	rowsBySymbol   map[string]*row.Model
}

// Messages for replacing assets
type SetAssetsMsg []c.Asset

// Messages for updating assets
type UpdateAssetsMsg []c.Asset

// NewModel returns a model with default values
func NewModel(config Config) *Model {
	return &Model{
		width:          80,
		config:         config,
		assets:         make([]*c.Asset, 0),
		assetsBySymbol: make(map[string]*c.Asset),
		sorter:         s.NewSorter(config.Sort),
		rowsBySymbol:   make(map[string]*row.Model),
	}
}

// Init initializes the watchlist
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the watchlist
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetAssetsMsg:

		// Convert []c.Asset to []*c.Asset and update assetsBySymbol map
		assets := make([]*c.Asset, len(msg))
		assetsBySymbol := make(map[string]*c.Asset)

		for i := range msg {
			assets[i] = &msg[i]
			assetsBySymbol[msg[i].Symbol] = assets[i]
		}

		assets = m.sorter(assets)

		for i, asset := range assets {
			if i < len(m.rows) {
				m.rows[i], _ = m.rows[i].Update(row.UpdateAssetMsg(asset))
				m.rowsBySymbol[assets[i].Symbol] = m.rows[i]
			} else {
				m.rows = append(m.rows, row.New(row.Config{
					Separate:              m.config.Separate,
					ExtraInfoExchange:     m.config.ExtraInfoExchange,
					ExtraInfoFundamentals: m.config.ExtraInfoFundamentals,
					ShowHoldings:          m.config.ShowHoldings,
					Styles:                m.config.Styles,
					Asset:                 asset,
				}))
				m.rowsBySymbol[assets[i].Symbol] = m.rows[len(m.rows)-1]
			}
		}

		m.assets = assets
		m.assetsBySymbol = assetsBySymbol

		// TODO: only set conditionally if all assets have changed
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, nil

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}
		return m, nil
	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {

	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	rows := make([]string, 0)
	for _, row := range m.rows {
		rows = append(rows, row.View())
	}

	return strings.Join(rows, "\n")

}
func getCellWidths(assets []*c.Asset) row.CellWidthsContainer {

	cellMaxWidths := row.CellWidthsContainer{}

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

		if quoteLength > cellMaxWidths.QuoteLength {
			cellMaxWidths.QuoteLength = quoteLength
			cellMaxWidths.WidthQuote = quoteLength + row.WidthChangeStatic
			cellMaxWidths.WidthQuoteExtended = quoteLength
			cellMaxWidths.WidthQuoteRange = row.WidthRangeStatic + (quoteLength * 2)
		}

		if asset.Holding != (c.Holding{}) {
			positionLength := len(u.ConvertFloatToString(asset.Holding.Value, asset.Meta.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(asset.Holding.Quantity, asset.Meta.IsVariablePrecision))

			if positionLength > cellMaxWidths.PositionLength {
				cellMaxWidths.PositionLength = positionLength
				cellMaxWidths.WidthPosition = positionLength + row.WidthChangeStatic + row.WidthPositionGutter
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
