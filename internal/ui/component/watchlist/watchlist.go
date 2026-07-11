package watchlist

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	s "github.com/achannarasappa/ticker/v5/internal/sorter"
	row "github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"
	u "github.com/achannarasappa/ticker/v5/internal/ui/util"

	tea "github.com/charmbracelet/bubbletea"
)

// Config represents the configuration for the watchlist component
type Config struct {
	Separate              bool
	ShowPositions         bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	Sort                  string
	Styles                c.Styles
}

// Model for watchlist section
type Model struct {
	width      int
	assets     []*c.Asset
	sorter     s.Sorter
	config     Config
	cellWidths row.CellWidthsContainer
	rows       []*row.Model
}

// Messages for replacing assets
type SetAssetsMsg []c.Asset

// Messages for updating assets
type UpdateAssetsMsg []c.Asset

// Messages for changing sort
type ChangeSortMsg string

// NewModel returns a model with default values
func NewModel(config Config) *Model {
	return &Model{
		width:  80,
		config: config,
		assets: make([]*c.Asset, 0),
		sorter: s.NewSorter(config.Sort),
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

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Convert []c.Asset to []*c.Asset
		assets := make([]*c.Asset, len(msg))

		for i := range msg {
			assets[i] = &msg[i]
		}

		assets = m.sorter(assets)

		for i, asset := range assets {
			if i < len(m.rows) {
				m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
				cmds = append(cmds, cmd)
			} else {
				m.rows = append(m.rows, row.New(row.Config{
					Separate:              m.config.Separate,
					ExtraInfoExchange:     m.config.ExtraInfoExchange,
					ExtraInfoFundamentals: m.config.ExtraInfoFundamentals,
					ShowPositions:         m.config.ShowPositions,
					Styles:                m.config.Styles,
					Asset:                 asset,
				}))
			}
		}

		if len(assets) < len(m.rows) {
			m.rows = m.rows[:len(assets)]
		}

		m.assets = assets

		// TODO: only set conditionally if all assets have changed
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, tea.Batch(cmds...)

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

	case row.FrameMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// TODO: send message to a specific row rather than all rows
		for i, r := range m.rows {
			m.rows[i], cmd = r.Update(msg)
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case ChangeSortMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Update the sorter with the new sort option
		m.config.Sort = string(msg)
		m.sorter = s.NewSorter(m.config.Sort)

		// Re-sort and update the assets
		assets := m.sorter(m.assets)
		m.assets = assets

		// Update rows with the new order
		for i, asset := range assets {
			m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {

	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	// Partition into options and everything else, preserving the sorted order within each.
	holdingsRows := make([]string, 0, len(m.rows))
	optionsRows := make([]string, 0)

	for i, row := range m.rows {
		if i < len(m.assets) && m.assets[i].Class == c.AssetClassOption {
			optionsRows = append(optionsRows, row.View())

			continue
		}

		holdingsRows = append(holdingsRows, row.View())
	}

	// When a group mixes holdings and options, render them as separate labeled lists.
	if len(holdingsRows) > 0 && len(optionsRows) > 0 {
		sections := make([]string, 0, len(m.rows)+3)
		sections = append(sections, m.sectionHeading("Holdings"))
		sections = append(sections, holdingsRows...)
		sections = append(sections, "")
		sections = append(sections, m.sectionHeading("Options"))
		sections = append(sections, optionsRows...)

		return strings.Join(sections, "\n")
	}

	rows := make([]string, 0, len(holdingsRows)+len(optionsRows))
	rows = append(rows, holdingsRows...)
	rows = append(rows, optionsRows...)

	return strings.Join(rows, "\n")

}

// sectionHeading renders a chip-style label followed by a full-width rule,
// used to visually separate the holdings and options lists within a group.
func (m *Model) sectionHeading(label string) string {

	chip := " " + strings.ToUpper(label) + " "

	ruleWidth := m.width - len(chip) - 1
	if ruleWidth < 0 {
		ruleWidth = 0
	}

	return m.config.Styles.TextHeader(chip) + " " + m.config.Styles.TextLine(strings.Repeat("─", ruleWidth))
}

func getCellWidths(assets []*c.Asset) row.CellWidthsContainer {

	cellMaxWidths := row.CellWidthsContainer{}

	for _, asset := range assets {
		var quoteLength int

		volumeMarketCapLength := len(u.ConvertFloatToStringWithCommas(asset.QuoteExtended.MarketCap, true))

		if asset.QuoteExtended.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(u.ConvertFloatToStringWithCommas(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision))
		}

		if asset.QuoteExtended.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(u.ConvertFloatToStringWithCommas(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision))
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

		if asset.Position != (c.Position{}) {
			positionLength := len(u.ConvertFloatToString(asset.Position.Value, asset.Meta.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(asset.Position.Quantity, asset.Meta.IsVariablePrecision))
			positionUnitCostLength := len(u.ConvertFloatToString(asset.Position.UnitCost, asset.Meta.IsVariablePrecision))

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

			if positionUnitCostLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionUnitCostLength
			}

		}

		// Options remap the fundamental columns and reuse position-extended for
		// contracts / total premium — size each to the values actually shown.
		if asset.Class == c.AssetClassOption {
			strikeLen := len(u.ConvertFloatToStringWithCommas(asset.QuoteOption.StrikePrice, asset.Meta.IsVariablePrecision))
			breakevenLen := len(u.ConvertFloatToStringWithCommas(asset.QuoteOption.BreakevenPrice, asset.Meta.IsVariablePrecision))
			diffLen := len(u.ConvertFloatToStringWithCommas(asset.QuoteOption.DiffToStrike, asset.Meta.IsVariablePrecision))
			premiumLen := len(u.ConvertFloatToStringWithCommas(asset.QuoteOption.Premium, asset.Meta.IsVariablePrecision))
			curPremiumLen := 1 // "-"
			if asset.QuoteOption.CurrentPremium != 0.0 {
				curPremiumLen = len(u.ConvertFloatToStringWithCommas(asset.QuoteOption.CurrentPremium, asset.Meta.IsVariablePrecision))
			}

			if strikeLen > cellMaxWidths.WidthVolumeMarketCap {
				cellMaxWidths.WidthVolumeMarketCap = strikeLen
			}
			if breakevenLen > cellMaxWidths.WidthVolumeMarketCap {
				cellMaxWidths.WidthVolumeMarketCap = breakevenLen
			}
			if diffLen > cellMaxWidths.WidthQuoteRange {
				cellMaxWidths.WidthQuoteRange = diffLen
			}
			if premiumLen > cellMaxWidths.WidthQuoteExtended {
				cellMaxWidths.WidthQuoteExtended = premiumLen
			}
			if curPremiumLen > cellMaxWidths.WidthQuoteExtended {
				cellMaxWidths.WidthQuoteExtended = curPremiumLen
			}

			totalPremium := math.Abs(asset.QuoteOption.Premium * asset.QuoteOption.Contracts * 100)
			contractsLength := len(strconv.FormatFloat(math.Abs(asset.QuoteOption.Contracts), 'f', 0, 64))
			totalPremiumLength := len(u.ConvertFloatToStringWithCommas(totalPremium, asset.Meta.IsVariablePrecision))

			if contractsLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = contractsLength
			}
			if totalPremiumLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = totalPremiumLength
			}
		}

	}

	return cellMaxWidths

}
