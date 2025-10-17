package ui

import (
	"fmt"
	"sync"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v5/internal/asset"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	mon "github.com/achannarasappa/ticker/v5/internal/monitor"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/summary"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"

	util "github.com/achannarasappa/ticker/v5/internal/ui/util"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

//nolint:gochecknoglobals
var (
	styleLogo  = util.NewStyle("#ffffd7", "#ff8700", true)
	styleGroup = util.NewStyle("#8a8a8a", "#303030", false)
	styleHelp  = util.NewStyle("#4e4e4e", "", true)
)

const (
	footerHeight = 1
)

// Model for UI
type Model struct {
	ctx                c.Context
	ready              bool
	headerHeight       int
	versionVector      int
	requestInterval    int
	assets             []c.Asset
	assetQuotes        []c.AssetQuote
	assetQuotesLookup  map[string]int
	holdingSummary     asset.HoldingSummary
	viewport           viewport.Model
	watchlist          *watchlist.Model
	summary            *summary.Model
	lastUpdateTime     string
	groupSelectedIndex int
	groupMaxIndex      int
	groupSelectedName  string
	currentSort        string
	monitors           *mon.Monitor
	mu                 sync.RWMutex
}

type tickMsg struct {
	versionVector int
}

type SetAssetQuoteMsg struct {
	symbol        string
	assetQuote    c.AssetQuote
	versionVector int
}

type SetAssetGroupQuoteMsg struct {
	assetGroupQuote c.AssetGroupQuote
	versionVector   int
}

// NewModel is the constructor for UI model
func NewModel(dep c.Dependencies, ctx c.Context, monitors *mon.Monitor) *Model {

	groupMaxIndex := len(ctx.Groups) - 1

	return &Model{
		ctx:               ctx,
		headerHeight:      getVerticalMargin(ctx.Config),
		ready:             false,
		requestInterval:   ctx.Config.RefreshInterval,
		versionVector:     0,
		assets:            make([]c.Asset, 0),
		assetQuotes:       make([]c.AssetQuote, 0),
		assetQuotesLookup: make(map[string]int),
		holdingSummary:    asset.HoldingSummary{},
		watchlist: watchlist.NewModel(watchlist.Config{
			Sort:                  ctx.Config.Sort,
			Separate:              ctx.Config.Separate,
			ShowHoldings:          ctx.Config.ShowHoldings,
			ExtraInfoExchange:     ctx.Config.ExtraInfoExchange,
			ExtraInfoFundamentals: ctx.Config.ExtraInfoFundamentals,
			Styles:                ctx.Reference.Styles,
		}),
		summary:            summary.NewModel(ctx),
		groupMaxIndex:      groupMaxIndex,
		groupSelectedIndex: 0,
		groupSelectedName:  "       ",
		currentSort:        ctx.Config.Sort,
		monitors:           monitors,
	}
}

// Init is the initialization hook for bubbletea
func (m *Model) Init() tea.Cmd {
	(*m.monitors).Start()

	// Start renderer and set symbols in parallel
	return tea.Batch(
		tick(0),
		func() tea.Msg {
			err := (*m.monitors).SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector)

			if m.ctx.Config.Debug && err != nil {
				m.ctx.Logger.Println(err)
			}

			return nil
		},
	)
}

// Update hook for bubbletea
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "tab", "shift+tab":
			m.mu.Lock()

			groupSelectedCursor := -1
			if msg.String() == "tab" {
				groupSelectedCursor = 1
			}

			m.groupSelectedIndex = (m.groupSelectedIndex + groupSelectedCursor + m.groupMaxIndex + 1) % (m.groupMaxIndex + 1)

			// Invalidate all previous ticks, incremental price updates, and full price updates
			m.versionVector++

			m.mu.Unlock()

			// Set the new set of symbols in the monitors and initiate a request to refresh all price quotes
			// Eventually, SetAssetGroupQuoteMsg message will be sent with the new quotes once all of the HTTP request complete
			m.monitors.SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector) //nolint:errcheck

			return m, tickImmediate(m.versionVector)
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, tea.Quit
		case "up":
			m.viewport, cmd = m.viewport.Update(msg)

			return m, cmd
		case "down":
			m.viewport, cmd = m.viewport.Update(msg)

			return m, cmd
		case "pgup":
			m.viewport.PageUp()

			return m, nil
		case "pgdown":
			m.viewport.PageDown()

			return m, nil
		case "s":
			m.mu.Lock()

			// Cycle through sort options: default -> alpha -> value -> user -> default
			sortOptions := []string{"", "alpha", "value", "user"}
			currentIndex := -1
			for i, sortOpt := range sortOptions {
				if m.currentSort == sortOpt {
					currentIndex = i

					break
				}
			}

			// Move to next sort option
			nextIndex := (currentIndex + 1) % len(sortOptions)
			m.currentSort = sortOptions[nextIndex]

			m.mu.Unlock()

			// Update watchlist component with new sort
			m.watchlist, cmd = m.watchlist.Update(watchlist.ChangeSortMsg(m.currentSort))

			return m, cmd

		}

	case tea.WindowSizeMsg:

		var cmd tea.Cmd

		m.mu.Lock()
		defer m.mu.Unlock()

		viewportHeight := msg.Height - m.headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		// Forward window size message to watchlist and summary component
		m.watchlist, cmd = m.watchlist.Update(msg)
		m.summary, _ = m.summary.Update(msg)

		return m, cmd

	// Trigger component re-render if data has changed
	case tickMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		m.mu.Lock()
		defer m.mu.Unlock()

		// Do not re-render if versionVector has changed and do not start a new timer with this versionVector
		if msg.versionVector != m.versionVector {
			return m, nil
		}

		// Update watchlist and summary components
		m.watchlist, cmd = m.watchlist.Update(watchlist.SetAssetsMsg(m.assets))
		m.summary, _ = m.summary.Update(summary.SetSummaryMsg(m.holdingSummary))

		cmds = append(cmds, cmd)

		// Set the current tick time
		m.lastUpdateTime = getTime()

		// Update the viewport
		if m.ready {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		cmds = append(cmds, tick(msg.versionVector))

		return m, tea.Batch(cmds...)

	case SetAssetGroupQuoteMsg:

		m.mu.Lock()
		defer m.mu.Unlock()

		// Do not update the assets and holding summary if the versionVector has changed
		if msg.versionVector != m.versionVector {
			return m, nil
		}

		assets, holdingSummary := asset.GetAssets(m.ctx, msg.assetGroupQuote)

		m.assets = assets
		m.holdingSummary = holdingSummary

		m.assetQuotes = msg.assetGroupQuote.AssetQuotes
		for i, assetQuote := range m.assetQuotes {
			m.assetQuotesLookup[assetQuote.Symbol] = i
		}

		m.groupSelectedName = m.ctx.Groups[m.groupSelectedIndex].Name

		return m, nil

	case SetAssetQuoteMsg:

		var i int
		var ok bool

		m.mu.Lock()
		defer m.mu.Unlock()

		if msg.versionVector != m.versionVector {
			return m, nil
		}

		// Check if this symbol is in the lookup
		if i, ok = m.assetQuotesLookup[msg.symbol]; !ok {
			return m, nil
		}

		// Check if the index is out of bounds
		if i >= len(m.assetQuotes) {
			return m, nil
		}

		// Check if the symbol is the same
		if m.assetQuotes[i].Symbol != msg.symbol {
			return m, nil
		}

		// Update the asset quote and generate a new holding summary
		m.assetQuotes[i] = msg.assetQuote

		assetGroupQuote := c.AssetGroupQuote{
			AssetQuotes: m.assetQuotes,
			AssetGroup:  m.ctx.Groups[m.groupSelectedIndex],
		}

		assets, holdingSummary := asset.GetAssets(m.ctx, assetGroupQuote)

		m.assets = assets
		m.holdingSummary = holdingSummary

		return m, nil

	case row.FrameMsg:
		var cmd tea.Cmd
		m.watchlist, cmd = m.watchlist.Update(msg)

		return m, cmd
	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.ready {
		return "\n  Initializing..."
	}

	m.viewport.SetContent(m.watchlist.View())

	viewSummary := ""

	if m.ctx.Config.ShowSummary && m.ctx.Config.ShowHoldings {
		viewSummary += m.summary.View() + "\n"
	}

	return viewSummary +
		m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime, m.groupSelectedName, m.currentSort)

}

func footer(width int, time string, groupSelectedName string, currentSort string) string {

	if width < 80 {
		return styleLogo(" ticker ")
	}

	if len(groupSelectedName) > 12 {
		groupSelectedName = groupSelectedName[:12]
	}

	// Get display name for current sort
	sortDisplayName := "change"
	switch currentSort {
	case "alpha":
		sortDisplayName = "alpha"
	case "value":
		sortDisplayName = "value"
	case "user":
		sortDisplayName = "user"
	}

	helpText := " q: exit ↑: scroll up ↓: scroll down s: change sort (" + sortDisplayName + ") ⭾: change group"

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: width,
				Cells: []grid.Cell{
					{Text: styleLogo(" ticker "), Width: 8},
					{Text: styleGroup(" " + groupSelectedName + " "), Width: len(groupSelectedName) + 2, VisibleMinWidth: 95},
					{Text: styleHelp(helpText), Width: len(helpText)},
					{Text: styleHelp("↻  " + time), Align: grid.Right},
				},
			},
		},
	})

}

func getVerticalMargin(config c.Config) int {
	if config.ShowSummary && config.ShowHoldings {
		return 2
	}

	return 0
}

// Send a new tick message with the versionVector 200ms from now
func tick(versionVector int) tea.Cmd {
	return tea.Tick(time.Second/5, func(time.Time) tea.Msg {
		return tickMsg{
			versionVector: versionVector,
		}
	})
}

// Send a new tick message immediately
func tickImmediate(versionVector int) tea.Cmd {

	return func() tea.Msg {
		return tickMsg{
			versionVector: versionVector,
		}
	}
}

func getTime() string {
	t := time.Now()

	return fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())
}
