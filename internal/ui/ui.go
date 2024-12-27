package ui

import (
	"fmt"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v4/internal/asset"
	c "github.com/achannarasappa/ticker/v4/internal/common"
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
	watchlist          watchlist.Model
	summary            summary.Model
	lastUpdateTime     string
	groupSelectedIndex int
	groupMaxIndex      int
	groupSelectedName  string
}

func getTime() string {
	t := time.Now()

	return fmt.Sprintf("%s %02d:%02d:%02d", t.Weekday().String(), t.Hour(), t.Minute(), t.Second())
}

func generateQuoteMsg(m Model, skipUpdate bool) func() tea.Msg {
	return func() tea.Msg {
		return quoteMsg{
			assetGroupIndex: m.groupSelectedIndex,
			assetGroupQuote: m.getQuotes(m.ctx.Groups[m.groupSelectedIndex]),
			skipUpdate:      skipUpdate,
			time:            getTime(),
		}
	}
}

func (m Model) updateQuotes() tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.requestInterval), func(_ time.Time) tea.Msg {
		return generateQuoteMsg(m, false)()
	})
}

// NewModel is the constructor for UI model
func NewModel(dep c.Dependencies, ctx c.Context) Model {

	groupMaxIndex := len(ctx.Groups) - 1

	return Model{
		ctx:                ctx,
		headerHeight:       getVerticalMargin(ctx.Config),
		ready:              false,
		requestInterval:    ctx.Config.RefreshInterval,
		getQuotes:          quote.GetAssetGroupQuote(dep, ctx.Reference),
		watchlist:          watchlist.NewModel(ctx),
		summary:            summary.NewModel(ctx),
		groupMaxIndex:      groupMaxIndex,
		groupSelectedIndex: 0,
		groupSelectedName:  "default",
	}
}

// Init is the initialization hook for bubbletea
func (m Model) Init() tea.Cmd {
	return generateQuoteMsg(m, false)
}

type quoteMsg struct {
	assetGroupIndex int
	assetGroupQuote c.AssetGroupQuote
	skipUpdate      bool
	time            string
}

// Update hook for bubbletea
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn,cyclop

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "tab", "shift+tab":
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
		m.watchlist.Width = msg.Width
		m.summary.Width = msg.Width
		viewportHeight := msg.Height - m.headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.Model{Width: msg.Width, Height: viewportHeight}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		m.viewport.SetContent(m.watchlist.View())

	case quoteMsg:
		// Update UI only if data matches current group
		if m.groupSelectedIndex == msg.assetGroupIndex {
			assets, holdingSummary := asset.GetAssets(m.ctx, msg.assetGroupQuote)
			m.watchlist.Assets = assets
			m.lastUpdateTime = msg.time
			m.summary.Summary = holdingSummary
			if m.ready {
				m.viewport.SetContent(m.watchlist.View())
			}
		}

		// Do not start a new timer to update
		if msg.skipUpdate {
			return m, nil
		}

		return m, m.updateQuotes()
	}

	m.viewport, _ = m.viewport.Update(msg)

	return m, nil
}

// View rendering hook for bubbletea
func (m Model) View() string {
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
