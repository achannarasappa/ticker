package ui

import (
	"fmt"
	"sync"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v4/internal/asset"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	mon "github.com/achannarasappa/ticker/v4/internal/monitor"
	"github.com/achannarasappa/ticker/v4/internal/ui/component/summary"
	"github.com/achannarasappa/ticker/v4/internal/ui/component/watchlist"

	util "github.com/achannarasappa/ticker/v4/internal/ui/util"

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
	nonce              int
	requestInterval    int
	assets             []c.Asset
	assetQuotes        []c.AssetQuote
	assetQuotesLookup  map[string]int
	holdingSummary     asset.HoldingSummary
	viewport           viewport.Model
	watchlist          *watchlist.Model
	summary            summary.Model
	lastUpdateTime     string
	groupSelectedIndex int
	groupMaxIndex      int
	groupSelectedName  string
	monitors           *mon.Monitor
	mu                 sync.RWMutex
}

type tickMsg struct {
	nonce int
}

type SetAssetQuoteMsg struct {
	symbol     string
	assetQuote c.AssetQuote
	nonce      int
}

type SetAssetGroupQuoteMsg struct {
	assetGroupQuote c.AssetGroupQuote
	nonce           int
}

// NewModel is the constructor for UI model
func NewModel(dep c.Dependencies, ctx c.Context, monitors *mon.Monitor) *Model {

	groupMaxIndex := len(ctx.Groups) - 1

	w := watchlist.NewModel(ctx)

	return &Model{
		ctx:                ctx,
		headerHeight:       getVerticalMargin(ctx.Config),
		ready:              false,
		requestInterval:    ctx.Config.RefreshInterval,
		nonce:              0,
		assets:             make([]c.Asset, 0),
		assetQuotes:        make([]c.AssetQuote, 0),
		assetQuotesLookup:  make(map[string]int),
		holdingSummary:     asset.HoldingSummary{},
		watchlist:          w,
		summary:            summary.NewModel(ctx),
		groupMaxIndex:      groupMaxIndex,
		groupSelectedIndex: 0,
		groupSelectedName:  "default",
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
			(*m.monitors).SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.nonce)
			return nil
		},
	)
}

// Update hook for bubbletea
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.groupSelectedName = m.ctx.Groups[m.groupSelectedIndex].Name

			// Invalidate all previous ticks, incremental price updates, and full price updates
			m.nonce++

			m.mu.Unlock()

			// Set the new set of symbols in the monitors and initiate a request to refresh all price quotes
			// Eventually, SetAssetGroupQuoteMsg message will be sent with the new quotes once all of the HTTP request complete
			m.monitors.SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.nonce)

			return m, tickImmediate(m.nonce)
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.mu.Lock()
		defer m.mu.Unlock()

		m.watchlist.Width = msg.Width
		m.summary.Width = msg.Width
		viewportHeight := msg.Height - m.headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		return m, nil

	// Trigger component re-render if data has changed
	case tickMsg:

		m.mu.Lock()
		defer m.mu.Unlock()

		// Do not re-render if nonce has changed and do not start a new timer with this nonce
		if msg.nonce != m.nonce {
			return m, nil
		}

		m.watchlist.Update(watchlist.SetAssetsMsg(m.assets))
		m.lastUpdateTime = getTime()
		m.summary.Summary = m.holdingSummary
		if m.ready {
			m.viewport.SetContent(m.watchlist.View())
			m.viewport, _ = m.viewport.Update(msg)
		}

		return m, tick(msg.nonce)

	case SetAssetGroupQuoteMsg:

		m.mu.Lock()
		defer m.mu.Unlock()

		// Do not update the assets and holding summary if the nonce has changed
		if msg.nonce != m.nonce {
			return m, nil
		}

		assets, holdingSummary := asset.GetAssets(m.ctx, msg.assetGroupQuote)

		m.assets = assets
		m.holdingSummary = holdingSummary

		m.assetQuotes = msg.assetGroupQuote.AssetQuotes
		for i, assetQuote := range m.assetQuotes {
			m.assetQuotesLookup[assetQuote.Symbol] = i
		}

		return m, nil

	case SetAssetQuoteMsg:

		var i int
		var ok bool

		m.mu.Lock()
		defer m.mu.Unlock()

		if msg.nonce != m.nonce {
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

	viewSummary := ""

	if m.ctx.Config.ShowSummary && m.ctx.Config.ShowHoldings {
		viewSummary += m.summary.View() + "\n"
	}

	return viewSummary +
		m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime, m.groupSelectedName)

}

func footer(width int, time string, groupSelectedName string) string {

	if width < 80 {
		return styleLogo(" ticker ")
	}

	if len(groupSelectedName) > 12 {
		groupSelectedName = groupSelectedName[:12]
	}

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: width,
				Cells: []grid.Cell{
					{Text: styleLogo(" ticker "), Width: 8},
					{Text: styleGroup(" " + groupSelectedName + " "), Width: len(groupSelectedName) + 2, VisibleMinWidth: 95},
					{Text: styleHelp(" q: exit ↑: scroll up ↓: scroll down ⭾: change group"), Width: 52},
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

// Send a new tick message with the nonce 200ms from now
func tick(nonce int) tea.Cmd {
	return tea.Tick(time.Second/5, func(time.Time) tea.Msg {
		return tickMsg{
			nonce: nonce,
		}
	})
}

// Send a new tick message immediately
func tickImmediate(nonce int) tea.Cmd {

	return func() tea.Msg {
		return tickMsg{
			nonce: nonce,
		}
	}
}

func getTime() string {
	t := time.Now()

	return fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())
}
