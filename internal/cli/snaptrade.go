package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/adrg/xdg"

	"github.com/achannarasappa/ticker/v5/internal/brokerage/snaptrade"
	"github.com/achannarasappa/ticker/v5/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// snapTradeSession resolves the read-path credentials for SnapTrade requests.
// ok is false when SnapTrade isn't configured or the user isn't connected.
func snapTradeSession(d c.Dependencies, config c.Config) (client *snaptrade.Client, userID string, userSecret string, ok bool) {

	st := config.SnapTrade

	if st.ClientID == "" || st.ConsumerKey == "" {
		return nil, "", "", false
	}

	if !st.IsPersonal() {
		if st.UserID == "" {
			logSnapTradeWarning(config, "commercial keys require `user-id`", nil)

			return nil, "", "", false
		}

		secret, found, err := snaptrade.NewStore(d.Fs, xdg.DataHome).GetUserSecret(st.UserID)
		if err != nil || !found {
			logSnapTradeWarning(config, "not connected — run `ticker snaptrade connect`", err)

			return nil, "", "", false
		}

		userID, userSecret = st.UserID, secret
	}

	return snaptrade.New(d.SnapTradeBaseURL, st.ClientID, st.ConsumerKey), userID, userSecret, true
}

// FetchSnapTradeAccountGroups lists the connected brokerage accounts and returns a
// placeholder asset group per account (name + account id, no holdings yet). This is
// fast; holdings for each account are loaded lazily via LoadSnapTradeAccountGroup.
func FetchSnapTradeAccountGroups(d c.Dependencies, config c.Config) ([]c.AssetGroup, error) {

	client, userID, userSecret, ok := snapTradeSession(d, config)
	if !ok {
		return nil, nil
	}

	accounts, err := client.ListAccounts(userID, userSecret)
	if err != nil {
		logSnapTradeWarning(config, "unable to list accounts", err)

		return nil, err
	}

	groups := make([]c.AssetGroup, 0, len(accounts))

	for _, account := range accounts {
		groups = append(groups, c.AssetGroup{
			ConfigAssetGroup:       c.ConfigAssetGroup{Name: snaptrade.AccountName(account)},
			IsSnapTrade:            true,
			SnapTradeAccountID:     account.ID,
			SnapTradeInstitution:   account.InstitutionName,
			SnapTradeAccountNumber: account.Number,
		})
	}

	return groups, nil
}

// LoadSnapTradeAccountGroup fetches the positions and options for a single account
// and returns a fully-resolved asset group. Called on demand when an account's tab
// is first viewed so the first account displays without waiting for the rest.
func LoadSnapTradeAccountGroup(d c.Dependencies, config c.Config, accountID string, accountName string) (c.AssetGroup, error) {

	client, userID, userSecret, ok := snapTradeSession(d, config)
	if !ok {
		return c.AssetGroup{}, nil
	}

	positions, err := client.GetPositions(userID, userSecret, accountID)
	if err != nil {
		logSnapTradeWarning(config, "unable to get positions", err)

		return c.AssetGroup{}, err
	}

	// Options come from a separate endpoint; a failure there shouldn't hide holdings.
	optionPositions, err := client.ListOptionHoldings(userID, userSecret, accountID)
	if err != nil {
		logSnapTradeWarning(config, "unable to get option holdings", err)

		optionPositions = nil
	}

	// Equity symbols resolve to the default source without the CSV, so degrade to an
	// empty map if it can't be fetched rather than failing the whole account.
	tickerSymbolToSourceSymbol, err := symbol.GetTickerSymbols(d.SymbolsURL, nil)
	if err != nil {
		tickerSymbolToSourceSymbol = symbol.TickerSymbolToSourceSymbol{}
	}

	account := snaptrade.Account{ID: accountID, Name: accountName}
	group := buildAssetGroup(snaptrade.TransformToConfigAssetGroup(account, positions, optionPositions), tickerSymbolToSourceSymbol)
	group.IsSnapTrade = true
	group.SnapTradeAccountID = accountID

	return group, nil
}

// LoadSnapTradePreferences returns the persisted account display preferences
// (empty preferences if none are stored or on error).
func LoadSnapTradePreferences(d c.Dependencies) snaptrade.Preferences {
	preferences, _ := snaptrade.NewStore(d.Fs, xdg.DataHome).GetPreferences()

	return preferences
}

// SaveSnapTradePreferences persists the account display preferences.
func SaveSnapTradePreferences(d c.Dependencies, preferences snaptrade.Preferences) error {
	return snaptrade.NewStore(d.Fs, xdg.DataHome).SavePreferences(preferences)
}

// ConnectSnapTrade runs the one-time connection flow: register the user if
// needed, generate a connection-portal link, and open it in the browser. The
// browser opener is injected for testability.
func ConnectSnapTrade(d c.Dependencies, config c.Config, openBrowser func(string) error) error {

	st := config.SnapTrade

	var missing []string
	if st.ClientID == "" {
		missing = append(missing, "client-id")
	}
	if st.ConsumerKey == "" {
		missing = append(missing, "consumer-key")
	}
	if !st.IsPersonal() && st.UserID == "" {
		missing = append(missing, "user-id")
	}
	if len(missing) > 0 {
		return fmt.Errorf("snaptrade: missing required config under `snaptrade:` in .ticker.yml: %s", strings.Join(missing, ", ")) //nolint:goerr113
	}

	client := snaptrade.New(d.SnapTradeBaseURL, st.ClientID, st.ConsumerKey)

	userID, userSecret := "", ""

	// Commercial keys register (once) and persist a user secret. Personal keys are
	// already provisioned with their single user, so registration is skipped.
	if !st.IsPersonal() {
		store := snaptrade.NewStore(d.Fs, xdg.DataHome)

		secret, ok, err := store.GetUserSecret(st.UserID)
		if err != nil {
			return err
		}

		if !ok {
			secret, err = client.RegisterUser(st.UserID)
			if err != nil {
				return fmt.Errorf("snaptrade: register user: %w", err)
			}

			if err := store.SaveUserSecret(st.UserID, secret); err != nil {
				return fmt.Errorf("snaptrade: save user secret: %w", err)
			}
		}

		userID, userSecret = st.UserID, secret
	}

	redirectURI, err := client.GetLoginRedirectURI(userID, userSecret)
	if err != nil {
		return fmt.Errorf("snaptrade: generate connection link: %w", err)
	}

	fmt.Println("Open the following URL to connect your brokerage:")
	fmt.Println()
	fmt.Println("  " + redirectURI)
	fmt.Println()

	if openBrowser != nil {
		if err := openBrowser(redirectURI); err != nil {
			fmt.Fprintln(os.Stderr, "(unable to open browser automatically — open the URL above manually)")
		}
	}

	fmt.Println("After authorizing in the browser, run `ticker` to see your holdings.")

	return nil
}

// OpenBrowser opens targetURL in the user's default browser.
func OpenBrowser(targetURL string) error {

	var name string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		name = "open"
	case "windows":
		name = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		name = "xdg-open"
	}

	args = append(args, targetURL)

	return exec.Command(name, args...).Start()
}

func logSnapTradeWarning(config c.Config, message string, err error) {

	if !config.Debug {
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "snaptrade: %s: %v\n", message, err)

		return
	}

	fmt.Fprintf(os.Stderr, "snaptrade: %s\n", message)
}
