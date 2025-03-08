package ui

import (
	"fmt"
	"sync"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v4/internal/asset"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	mon "github.com/achannarasappa/ticker/v4/internal/monitor"
	quote "github.com/achannarasappa/ticker/v4/internal/quote"
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
	getQuotes          func(c.AssetGroup) c.AssetGroupQuote
	requestInterval    int
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

func getTime() string {
	t := time.Now()

	return fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())
}

func generateQuoteMsg(m *Model, skipUpdate bool) tea.Cmd {
	return func() tea.Msg {
		// Infer group change based on skipUpdate
		if skipUpdate {
			m.monitors.SetSymbols(m.ctx.Groups[m.groupSelectedIndex])
		}
		return quoteMsg{
			assetGroupIndex: m.groupSelectedIndex,
			assetGroupQuote: m.getQuotes(m.ctx.Groups[m.groupSelectedIndex]),
			skipUpdate:      skipUpdate,
			time:            getTime(),
		}
	}
}

func (m *Model) updateQuotes() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.requestInterval), func(_ time.Time) tea.Msg {
		return generateQuoteMsg(m, false)()
	})
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
		getQuotes:          quote.GetAssetGroupQuote(monitors, &dep),
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

	(*m.monitors).SetSymbols(m.ctx.Groups[m.groupSelectedIndex])

	return generateQuoteMsg(m, false)
}

type quoteMsg struct {
	assetGroupIndex int
	assetGroupQuote c.AssetGroupQuote
	skipUpdate      bool
	time            string
}

type SetAssetMsg struct {
	symbol        string
	asset         c.Asset
	assetGroupIdx int
}

type SetAssetsMsg struct {
	assets        []c.Asset
	assetGroupIdx int
}

// Update hook for bubbletea
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "tab", "shift+tab":
			m.mu.Lock()
			defer m.mu.Unlock()

			groupSelectedCursor := -1
			if msg.String() == "tab" {
				groupSelectedCursor = 1
			}

			m.groupSelectedIndex = (m.groupSelectedIndex + groupSelectedCursor + m.groupMaxIndex + 1) % (m.groupMaxIndex + 1)
			m.groupSelectedName = m.ctx.Groups[m.groupSelectedIndex].Name

			return m, generateQuoteMsg(m, true)
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

	case watchlist.SetAssetQuotePriceMsg:
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.ready {
			m.watchlist.Update(watchlist.SetAssetQuotePriceMsg{
				Symbol:     msg.Symbol,
				QuotePrice: msg.QuotePrice,
			})
			m.lastUpdateTime = getTime()
			m.viewport.SetContent(m.watchlist.View())
			m.viewport, _ = m.viewport.Update(msg)
		}

		return m, nil

	case quoteMsg:
		m.mu.Lock()
		defer m.mu.Unlock()

		// Update UI only if data matches current group
		if m.groupSelectedIndex == msg.assetGroupIndex {
			assets, holdingSummary := asset.GetAssets(m.ctx, msg.assetGroupQuote)
			m.watchlist.Update(watchlist.SetAssetsMsg(assets))
			m.lastUpdateTime = msg.time
			m.summary.Summary = holdingSummary
			if m.ready {
				m.viewport.SetContent(m.watchlist.View())
				m.viewport, _ = m.viewport.Update(msg)
			}
		}

		// Do not start a new timer to update
		if msg.skipUpdate {
			return m, nil
		}

		return m, m.updateQuotes()
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
