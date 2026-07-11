package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	grid "github.com/achannarasappa/term-grid"
	"github.com/achannarasappa/ticker/v5/internal/asset"
	"github.com/achannarasappa/ticker/v5/internal/brokerage/snaptrade"
	"github.com/achannarasappa/ticker/v5/internal/cli"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	mon "github.com/achannarasappa/ticker/v5/internal/monitor"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/summary"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"
	"github.com/achannarasappa/ticker/v5/internal/updater"

	util "github.com/achannarasappa/ticker/v5/internal/ui/util"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/afero"
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
	dep                c.Dependencies
	ready              bool
	width              int
	height             int
	loadingAccounts    bool
	snapTradeLoading   map[string]bool
	snapTradeLoaded    map[string]bool
	configGroups       []c.AssetGroup // config-derived groups (source of truth for the non-SnapTrade tabs)
	snapTradeAccounts  []c.AssetGroup // full account roster (source of truth; loaded holdings replace entries)
	snapTradePrefs     snaptrade.Preferences
	modal              accountModal
	headerHeight       int
	versionVector      int
	requestInterval    int
	assets             []c.Asset
	assetQuotes        []c.AssetQuote
	assetQuotesLookup  map[string]int
	positionSummary    asset.PositionSummary
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
	version            string
	latestVersion      string
	releasesURL        string
	fs                 afero.Fs
}

// accountModal is the state for the SnapTrade account selection overlay
type accountModal struct {
	open      bool
	cursor    int
	items     []accountModalItem
	defaultID string
}

type accountModalItem struct {
	accountID string
	name      string
	shown     bool
}

type tickMsg struct {
	versionVector int
}

type updateCheckMsg string

type updateCheckTickMsg struct{}

type SetAssetQuoteMsg struct {
	symbol        string
	assetQuote    c.AssetQuote
	versionVector int
}

type SetAssetGroupQuoteMsg struct {
	assetGroupQuote c.AssetGroupQuote
	versionVector   int
}

// snapTradeAccountsMsg delivers the (fast) account list as placeholder groups
type snapTradeAccountsMsg struct {
	groups []c.AssetGroup
}

// snapTradeHoldingsMsg delivers one account's fully-loaded holdings group
type snapTradeHoldingsMsg struct {
	accountID string
	group     c.AssetGroup
}

// NewModel is the constructor for UI model
func NewModel(dep c.Dependencies, ctx c.Context, monitors *mon.Monitor, version string) *Model {

	groupMaxIndex := len(ctx.Groups) - 1

	return &Model{
		ctx:               ctx,
		dep:               dep,
		headerHeight:      getVerticalMargin(ctx.Config),
		ready:             false,
		loadingAccounts:   isSnapTradeConfigured(ctx.Config),
		snapTradeLoading:  make(map[string]bool),
		snapTradeLoaded:   make(map[string]bool),
		configGroups:      ctx.Groups,
		snapTradePrefs:    cli.LoadSnapTradePreferences(dep),
		requestInterval:   ctx.Config.RefreshInterval,
		versionVector:     0,
		assets:            make([]c.Asset, 0),
		assetQuotes:       make([]c.AssetQuote, 0),
		assetQuotesLookup: make(map[string]int),
		positionSummary:   asset.PositionSummary{},
		watchlist: watchlist.NewModel(watchlist.Config{
			Sort:                  ctx.Config.Sort,
			Separate:              ctx.Config.Separate,
			ShowPositions:         ctx.Config.ShowPositions,
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
		version:            version,
		releasesURL:        dep.GitHubReleasesURL,
		fs:                 dep.Fs,
	}
}

// Init is the initialization hook for bubbletea
func (m *Model) Init() tea.Cmd {
	(*m.monitors).Start()

	cmds := []tea.Cmd{
		tick(0),
		updateCheckTick(),
		func() tea.Msg {
			return updateCheckMsg(updater.Check(m.version, m.releasesURL, updater.CacheFilePath(), m.fs))
		},
	}

	// Load the first config-derived group's quotes (if any exist yet)
	if len(m.ctx.Groups) > 0 {
		cmds = append(cmds, func() tea.Msg {
			err := (*m.monitors).SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector)

			if m.ctx.Config.Debug && err != nil {
				m.ctx.Logger.Println(err)
			}

			return nil
		})
	}

	// Fetch the SnapTrade account list asynchronously (fast); holdings load lazily per tab
	if m.loadingAccounts {
		cmds = append(cmds, m.fetchSnapTradeAccounts())
	}

	return tea.Batch(cmds...)
}

// fetchSnapTradeAccounts returns a command that lists the brokerage accounts off
// the UI thread and delivers them as placeholder groups.
func (m *Model) fetchSnapTradeAccounts() tea.Cmd {
	return func() tea.Msg {
		groups, err := cli.FetchSnapTradeAccountGroups(m.dep, m.ctx.Config)
		if err != nil && m.ctx.Config.Debug && m.ctx.Logger != nil {
			m.ctx.Logger.Println(err)
		}

		return snapTradeAccountsMsg{groups: groups}
	}
}

// loadSnapTradeAccount returns a command that fetches one account's holdings.
func (m *Model) loadSnapTradeAccount(group c.AssetGroup) tea.Cmd {
	return func() tea.Msg {
		loaded, err := cli.LoadSnapTradeAccountGroup(m.dep, m.ctx.Config, group.SnapTradeAccountID, group.Name)
		if err != nil && m.ctx.Config.Debug && m.ctx.Logger != nil {
			m.ctx.Logger.Println(err)
		}

		return snapTradeHoldingsMsg{accountID: group.SnapTradeAccountID, group: loaded}
	}
}

// loadSelectedGroupCmd returns a command to lazily load the selected group's
// holdings when it is a SnapTrade account that hasn't been loaded yet.
func (m *Model) loadSelectedGroupCmd() tea.Cmd {
	if m.groupSelectedIndex < 0 || m.groupSelectedIndex > m.groupMaxIndex {
		return nil
	}

	group := m.ctx.Groups[m.groupSelectedIndex]
	if !group.IsSnapTrade || m.snapTradeLoaded[group.SnapTradeAccountID] || m.snapTradeLoading[group.SnapTradeAccountID] {
		return nil
	}

	m.snapTradeLoading[group.SnapTradeAccountID] = true

	return m.loadSnapTradeAccount(group)
}

func isSnapTradeConfigured(config c.Config) bool {
	return config.SnapTrade.ClientID != "" && config.SnapTrade.ConsumerKey != ""
}

// rebuildGroups derives the visible tab list (ctx.Groups) from the config groups
// plus the non-hidden SnapTrade accounts, and clamps the selection. Callers hold m.mu.
func (m *Model) rebuildGroups() {
	groups := make([]c.AssetGroup, 0, len(m.configGroups)+len(m.snapTradeAccounts))
	groups = append(groups, m.configGroups...)

	for _, account := range m.snapTradeAccounts {
		if !m.snapTradePrefs.IsHidden(account.SnapTradeAccountID) {
			groups = append(groups, account)
		}
	}

	m.ctx.Groups = groups
	m.groupMaxIndex = len(groups) - 1

	if m.groupSelectedIndex > m.groupMaxIndex {
		m.groupSelectedIndex = m.groupMaxIndex
	}
	if m.groupSelectedIndex < 0 {
		m.groupSelectedIndex = 0
	}
}

// defaultGroupIndex returns the index of the preferred default account in the
// visible groups, or -1 when no (visible) default is set.
func (m *Model) defaultGroupIndex() int {
	if m.snapTradePrefs.DefaultAccountID == "" {
		return -1
	}

	for i, group := range m.ctx.Groups {
		if group.IsSnapTrade && group.SnapTradeAccountID == m.snapTradePrefs.DefaultAccountID {
			return i
		}
	}

	return -1
}

// updateRosterAccount replaces the roster entry for a loaded account, preserving
// the display fields that only the account list carries. Callers hold m.mu.
func (m *Model) updateRosterAccount(group c.AssetGroup) {
	for i := range m.snapTradeAccounts {
		if m.snapTradeAccounts[i].SnapTradeAccountID == group.SnapTradeAccountID {
			group.SnapTradeInstitution = m.snapTradeAccounts[i].SnapTradeInstitution
			group.SnapTradeAccountNumber = m.snapTradeAccounts[i].SnapTradeAccountNumber

			if group.Name == "" {
				group.Name = m.snapTradeAccounts[i].Name
			}

			m.snapTradeAccounts[i] = group

			return
		}
	}
}

// updateModal handles key events while the account selector is open.
func (m *Model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.modal.cursor > 0 {
			m.modal.cursor--
		}
	case "down", "j":
		if m.modal.cursor < len(m.modal.items)-1 {
			m.modal.cursor++
		}
	case " ":
		if len(m.modal.items) > 0 {
			m.modal.items[m.modal.cursor].shown = !m.modal.items[m.modal.cursor].shown
		}
	case "d":
		if len(m.modal.items) > 0 {
			id := m.modal.items[m.modal.cursor].accountID
			if m.modal.defaultID == id {
				m.modal.defaultID = ""
			} else {
				m.modal.defaultID = id
			}
		}
	case "esc", "q":
		m.modal.open = false
	case "enter":
		return m.commitModal()
	}

	return m, nil
}

// commitModal persists the selector choices, re-derives the visible groups, and
// jumps to the (new) default account.
func (m *Model) commitModal() (tea.Model, tea.Cmd) {
	m.mu.Lock()

	hidden := make([]string, 0)
	for _, item := range m.modal.items {
		if !item.shown {
			hidden = append(hidden, item.accountID)
		}
	}

	defaultID := m.modal.defaultID
	for _, id := range hidden {
		if id == defaultID {
			defaultID = "" // can't start on a hidden account
		}
	}

	preferences := snaptrade.Preferences{HiddenAccounts: hidden, DefaultAccountID: defaultID}
	m.snapTradePrefs = preferences
	m.modal.open = false
	m.rebuildGroups()

	if index := m.defaultGroupIndex(); index >= 0 {
		m.groupSelectedIndex = index
	}

	m.versionVector++

	m.mu.Unlock()

	if err := cli.SaveSnapTradePreferences(m.dep, preferences); err != nil && m.ctx.Config.Debug && m.ctx.Logger != nil {
		m.ctx.Logger.Println(err)
	}

	if m.groupMaxIndex < 0 {
		return m, nil
	}

	m.monitors.SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector) //nolint:errcheck

	return m, tea.Batch(tickImmediate(m.versionVector), m.loadSelectedGroupCmd())
}

// modalView renders the account selector as a centered box.
func (m *Model) modalView() string {
	styles := m.ctx.Reference.Styles

	var body strings.Builder
	body.WriteString(styles.TextBold("Show accounts") + "\n\n")

	for i, item := range m.modal.items {
		cursor := "  "
		if i == m.modal.cursor {
			cursor = "❯ "
		}

		check := "[ ]"
		if item.shown {
			check = "[x]"
		}

		line := cursor + check + " " + item.name
		if item.accountID == m.modal.defaultID {
			line += "  ★ default"
		}

		if i == m.modal.cursor {
			line = styles.TextBold(line)
		} else {
			line = styles.Text(line)
		}

		body.WriteString(line + "\n")
	}

	body.WriteString("\n" + styles.TextLabel("↑/↓ move · space show/hide · d default · enter save · esc cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2).
		Render(body.String())

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	return box
}

// Update hook for bubbletea
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:maintidx,gocyclo
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		// While the account selector is open, all keys drive it
		if m.modal.open {
			return m.updateModal(msg)
		}

		switch msg.String() {

		case "a":

			// Open the SnapTrade account selector (no-op until accounts are loaded)
			if len(m.snapTradeAccounts) == 0 {
				return m, nil
			}

			m.mu.Lock()
			m.modal.items = m.modal.items[:0]
			for _, account := range m.snapTradeAccounts {
				m.modal.items = append(m.modal.items, accountModalItem{
					accountID: account.SnapTradeAccountID,
					name:      account.Name,
					shown:     !m.snapTradePrefs.IsHidden(account.SnapTradeAccountID),
				})
			}
			m.modal.defaultID = m.snapTradePrefs.DefaultAccountID
			m.modal.cursor = 0
			m.modal.open = true
			m.mu.Unlock()

			return m, nil

		case "tab", "shift+tab":
			m.mu.Lock()

			// No groups yet (e.g. SnapTrade still loading) — nothing to switch between
			if m.groupMaxIndex < 0 {
				m.mu.Unlock()

				return m, nil
			}

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

			// Lazily load this account's holdings the first time it is viewed
			return m, tea.Batch(tickImmediate(m.versionVector), m.loadSelectedGroupCmd())
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

		case "r":

			// Re-fetch SnapTrade holdings on demand. No-op when SnapTrade is not configured.
			if !isSnapTradeConfigured(m.ctx.Config) {
				return m, nil
			}

			m.mu.Lock()
			m.loadingAccounts = true
			m.snapTradeLoading = make(map[string]bool)
			m.snapTradeLoaded = make(map[string]bool)
			m.mu.Unlock()

			return m, m.fetchSnapTradeAccounts()

		}

	case tea.WindowSizeMsg:

		var cmd tea.Cmd

		m.mu.Lock()
		defer m.mu.Unlock()

		m.width = msg.Width
		m.height = msg.Height

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
		m.summary, _ = m.summary.Update(summary.SetSummaryMsg(m.positionSummary))
		m.summary, _ = m.summary.Update(summary.SetHeaderMsg(m.groupHeaderLabel()))

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

		// Do not update the assets and position summary if the versionVector has changed
		if msg.versionVector != m.versionVector {
			return m, nil
		}

		assets, positionSummary := asset.GetAssets(m.ctx, msg.assetGroupQuote)

		m.assets = assets
		m.positionSummary = positionSummary

		m.assetQuotes = msg.assetGroupQuote.AssetQuotes
		for i, assetQuote := range m.assetQuotes {
			m.assetQuotesLookup[assetQuote.Symbol] = i
		}

		return m, nil

	case snapTradeAccountsMsg:

		m.mu.Lock()

		m.snapTradeAccounts = msg.groups
		m.loadingAccounts = false
		m.rebuildGroups()

		// Start on the configured default account when one is set and visible
		if index := m.defaultGroupIndex(); index >= 0 {
			m.groupSelectedIndex = index
		}

		m.versionVector++

		m.mu.Unlock()

		if m.groupMaxIndex < 0 {
			return m, nil
		}

		m.monitors.SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector) //nolint:errcheck

		return m, tea.Batch(tickImmediate(m.versionVector), m.loadSelectedGroupCmd())

	case snapTradeHoldingsMsg:

		m.mu.Lock()

		m.snapTradeLoading[msg.accountID] = false

		selectedAccountID := ""
		if m.groupSelectedIndex >= 0 && m.groupSelectedIndex <= m.groupMaxIndex && m.ctx.Groups[m.groupSelectedIndex].IsSnapTrade {
			selectedAccountID = m.ctx.Groups[m.groupSelectedIndex].SnapTradeAccountID
		}

		// A populated group carries its account id; on failure the placeholder is kept
		if msg.group.SnapTradeAccountID != "" {
			m.snapTradeLoaded[msg.accountID] = true
			m.updateRosterAccount(msg.group)
			m.rebuildGroups()

			// Visibility is unchanged by a holdings load, but re-anchor selection to be safe
			if selectedAccountID != "" {
				for i, group := range m.ctx.Groups {
					if group.IsSnapTrade && group.SnapTradeAccountID == selectedAccountID {
						m.groupSelectedIndex = i

						break
					}
				}
			}
		}

		selected := msg.accountID == selectedAccountID

		if selected {
			m.versionVector++
		}

		m.mu.Unlock()

		// If the loaded account is on screen, fetch its quotes now
		if selected {
			m.monitors.SetAssetGroup(m.ctx.Groups[m.groupSelectedIndex], m.versionVector) //nolint:errcheck

			return m, tickImmediate(m.versionVector)
		}

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

		// Update the asset quote and generate a new position summary
		m.assetQuotes[i] = msg.assetQuote

		assetGroupQuote := c.AssetGroupQuote{
			AssetQuotes: m.assetQuotes,
			AssetGroup:  m.ctx.Groups[m.groupSelectedIndex],
		}

		assets, positionSummary := asset.GetAssets(m.ctx, assetGroupQuote)

		m.assets = assets
		m.positionSummary = positionSummary

		return m, nil

	case row.FrameMsg:
		var cmd tea.Cmd
		m.watchlist, cmd = m.watchlist.Update(msg)

		return m, cmd

	case updateCheckMsg:
		m.latestVersion = string(msg)

		return m, nil

	case updateCheckTickMsg:
		return m, tea.Batch(
			updateCheckTick(),
			func() tea.Msg {
				return updateCheckMsg(updater.Check(m.version, m.releasesURL, updater.CacheFilePath(), m.fs))
			},
		)
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

	if m.modal.open {
		return m.modalView()
	}

	m.viewport.SetContent(m.mainContent())

	viewSummary := ""

	if m.ctx.Config.ShowSummary {
		viewSummary += m.summary.View() + "\n"
	}

	return viewSummary +
		m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime, m.groupChipLabel(), m.currentSort, m.isLoading(), m.latestVersion)

}

// groupChipLabel is the short label shown in the footer chip: the brokerage source
// for SnapTrade accounts (e.g. "Robinhood"), otherwise the group name.
func (m *Model) groupChipLabel() string {
	if m.groupSelectedIndex < 0 || m.groupSelectedIndex > m.groupMaxIndex {
		return ""
	}

	group := m.ctx.Groups[m.groupSelectedIndex]
	if group.IsSnapTrade && group.SnapTradeInstitution != "" {
		return group.SnapTradeInstitution
	}

	return group.Name
}

// groupHeaderLabel is the fuller label shown top-right for a SnapTrade account:
// the account name plus a masked last-4, where there is room for it.
func (m *Model) groupHeaderLabel() string {
	if m.groupSelectedIndex < 0 || m.groupSelectedIndex > m.groupMaxIndex {
		return ""
	}

	group := m.ctx.Groups[m.groupSelectedIndex]
	if !group.IsSnapTrade {
		return ""
	}

	label := group.Name

	digits := digitsOnly(group.SnapTradeAccountNumber)
	if len(digits) >= 4 {
		last4 := digits[len(digits)-4:]
		if !strings.Contains(group.Name, last4) {
			label += " ••" + last4
		}
	}

	return label
}

func digitsOnly(value string) string {
	var builder strings.Builder

	for _, r := range value {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// mainContent renders the watchlist, or a status message while accounts/holdings load.
func (m *Model) mainContent() string {

	label := func(text string) string {
		return "\n  " + m.ctx.Reference.Styles.TextLabel(text)
	}

	if len(m.ctx.Groups) == 0 {
		if m.loadingAccounts {
			return label("Loading accounts…")
		}

		return label("No watchlist or holdings configured")
	}

	selected := m.ctx.Groups[m.groupSelectedIndex]

	if selected.IsSnapTrade {
		if m.snapTradeLoading[selected.SnapTradeAccountID] {
			return label("Loading positions…")
		}

		if m.snapTradeLoaded[selected.SnapTradeAccountID] && len(m.assets) == 0 {
			return label("No holdings in this account")
		}
	}

	return m.watchlist.View()
}

// isLoading reports whether the account list or the selected account is still loading.
func (m *Model) isLoading() bool {
	if m.loadingAccounts {
		return true
	}

	if m.groupSelectedIndex < 0 || m.groupSelectedIndex > m.groupMaxIndex {
		return false
	}

	selected := m.ctx.Groups[m.groupSelectedIndex]

	return selected.IsSnapTrade && m.snapTradeLoading[selected.SnapTradeAccountID]
}

func footer(width int, time string, groupSelectedName string, currentSort string, loading bool, latestVersion string) string {

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

	baseHelpText := " q: exit ↑: scroll up ↓: scroll down ⭾: change group"
	sortHelpText := " s: change sort (" + sortDisplayName + ") a: accounts r: refresh"

	rightText := "↻  " + time
	if latestVersion != "" {
		rightText = "↑ " + latestVersion + " available"
	}
	if loading {
		rightText = "⟳ loading positions… " + rightText
	}

	// Calculate minimum width for sort help text to appear
	// Longest sort text is "s: change sort (change)" = 24 characters
	// Minimum width needed: logo(8) + max group(14) + base help(52) + sort help(24) + time(12) = 110
	const sortHelpMinWidth = 114

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: width,
				Cells: []grid.Cell{
					{Text: styleLogo(" ticker "), Width: 8},
					{Text: styleGroup(" " + groupSelectedName + " "), Width: len(groupSelectedName) + 2, VisibleMinWidth: 95},
					{Text: styleHelp(baseHelpText), Width: 52},
					{Text: styleHelp(sortHelpText), Width: len(sortHelpText), VisibleMinWidth: sortHelpMinWidth},
					{Text: styleHelp(rightText), Align: grid.Right},
				},
			},
		},
	})

}

func getVerticalMargin(config c.Config) int {
	if config.ShowSummary {
		return 2
	}

	return 0
}

func updateCheckTick() tea.Cmd {
	return tea.Tick(3*time.Hour, func(time.Time) tea.Msg {
		return updateCheckTickMsg{}
	})
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
