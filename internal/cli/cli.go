package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/achannarasappa/ticker/v5/internal/cache"
	"github.com/achannarasappa/ticker/v5/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/achannarasappa/ticker/v5/internal/ui/util"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Options to configured ticker behavior
type Options struct {
	RefreshInterval       int
	Watchlist             string
	Separate              bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	ShowSummary           bool
	ShowHoldings          bool // Deprecated: use ShowPositions instead, kept for backwards compatibility
	ShowPositions         bool // Preferred field name
	Sort                  string
	NoCache               bool
	Debug                 bool
}

type symbolSource struct {
	symbol string
	source c.QuoteSource
}

// Run starts the ticker UI
func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		err := uiStartFn()

		if err != nil {
			fmt.Println(fmt.Errorf("unable to start UI: %w", err).Error())
		}
	}
}

// validateLot validates a single lot and returns an error if invalid
func validateLot(lot c.Lot, groupName string, lotIndex int) error {
	if lot.Symbol == "" {
		return fmt.Errorf("invalid config: lot #%d in group '%s' has empty symbol", lotIndex+1, groupName) //nolint:goerr113
	}

	if lot.Quantity == 0 {
		return fmt.Errorf("invalid config: lot #%d for symbol '%s' in group '%s' has invalid quantity (cannot be zero, got %f)", lotIndex+1, lot.Symbol, groupName, lot.Quantity) //nolint:goerr113
	}

	if lot.UnitCost < 0 {
		return fmt.Errorf("invalid config: lot #%d for symbol '%s' in group '%s' has invalid unit_cost (must be zero or positive, got %f)", lotIndex+1, lot.Symbol, groupName, lot.UnitCost) //nolint:goerr113
	}

	if lot.FixedCost < 0 {
		return fmt.Errorf("invalid config: lot #%d for symbol '%s' in group '%s' has invalid fixed_cost (must be zero or positive, got %f)", lotIndex+1, lot.Symbol, groupName, lot.FixedCost) //nolint:goerr113
	}

	return nil
}

// Validate checks whether config is valid and returns an error if invalid or if an error was generated earlier
func Validate(config *c.Config, options *Options, prevErr *error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {

		if prevErr != nil && *prevErr != nil {
			return *prevErr
		}

		if len(config.Watchlist) == 0 && len(options.Watchlist) == 0 && len(config.Lots) == 0 && len(config.AssetGroup) == 0 && config.SnapTrade.ClientID == "" {
			return errors.New("invalid config: No watchlist provided") //nolint:goerr113
		}

		if len(config.Currency) > 0 && (strings.ToUpper(config.Currency) != config.Currency || len(config.Currency) != 3) {
			return errors.New("invalid config: Display currency may only be an ISO 4217 major currency or blank (eg GBP not GBp; default: USD)") //nolint:goerr113
		}

		// Validate lots in config.Lots (default group)
		for i, lot := range config.Lots {
			if err := validateLot(lot, "default", i); err != nil {
				return err
			}
		}

		// Validate lots in config.AssetGroup
		for _, assetGroup := range config.AssetGroup {
			groupName := assetGroup.Name
			if groupName == "" {
				groupName = "unnamed"
			}
			// Use Lots (preferred), otherwise fall back to Holdings for backwards compatibility
			lots := assetGroup.Lots
			if len(lots) == 0 {
				lots = assetGroup.Holdings
			}
			for i, lot := range lots {
				if err := validateLot(lot, groupName, i); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func GetDependencies() c.Dependencies {
	return c.Dependencies{
		Fs:                               afero.NewOsFs(),
		SymbolsURL:                       "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv",
		GitHubReleasesURL:                "https://api.github.com/repos/achannarasappa/ticker/releases/latest",
		MonitorYahooBaseURL:              "https://query1.finance.yahoo.com",
		MonitorYahooSessionRootURL:       "https://finance.yahoo.com",
		MonitorYahooSessionCrumbURL:      "https://query2.finance.yahoo.com",
		MonitorYahooSessionConsentURL:    "https://consent.yahoo.com",
		MonitorPriceCoinbaseBaseURL:      "https://api.coinbase.com",
		MonitorPriceCoinbaseStreamingURL: "wss://ws-feed.exchange.coinbase.com",
		SnapTradeBaseURL:                 "https://api.snaptrade.com/api/v1",
	}
}

// GetContext builds the context from the config and reference data
func GetContext(d c.Dependencies, config c.Config) (c.Context, error) {
	var (
		reference c.Reference
		groups    []c.AssetGroup
		err       error
	)

	var logger *log.Logger

	if config.Debug {
		logger, err = getLogger(d)

		if err != nil {
			return c.Context{}, err
		}
	}

	// SnapTrade accounts and their holdings are fetched asynchronously by the UI
	// (see FetchSnapTradeAccountGroups / LoadSnapTradeAccountGroup) so a slow
	// brokerage API doesn't block startup; getGroups builds config-derived groups only.
	cache := cache.New(d.Fs, cache.FilePath(), cacheEnabled(config.Cache))

	groups, err = getGroups(config, d, cache)

	if err != nil {
		return c.Context{}, err
	}

	reference, err = getReference(config)

	if err != nil {
		return c.Context{}, err
	}

	context := c.Context{
		Reference: reference,
		Config:    config,
		Groups:    groups,
		Logger:    logger,
		Cache:     cache,
	}

	return context, err
}

func readConfig(fs afero.Fs, configPathOption string) (c.Config, error) {
	var config c.Config
	configPath, err := getConfigPath(fs, configPathOption)

	if err != nil {
		return config, nil //nolint:nilerr
	}
	handle, err := fs.Open(configPath)

	if err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}

	defer handle.Close()
	err = yaml.NewDecoder(handle).Decode(&config)

	if err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

func getReference(config c.Config) (c.Reference, error) {
	styles := util.GetColorScheme(config.ColorScheme)

	return c.Reference{
		Styles: styles,
	}, nil
}

func GetConfig(dep c.Dependencies, configPath string, options Options) (c.Config, error) {

	config, err := readConfig(dep.Fs, configPath)

	if err != nil {
		return c.Config{}, err
	}

	if len(options.Watchlist) != 0 {
		config.Watchlist = strings.Split(strings.ReplaceAll(options.Watchlist, " ", ""), ",")
	}

	config.RefreshInterval = getRefreshInterval(options.RefreshInterval, config.RefreshInterval)
	config.Separate = getBoolOption(options.Separate, config.Separate)
	config.ExtraInfoExchange = getBoolOption(options.ExtraInfoExchange, config.ExtraInfoExchange)
	config.ExtraInfoFundamentals = getBoolOption(options.ExtraInfoFundamentals, config.ExtraInfoFundamentals)
	config.ShowSummary = getBoolOption(options.ShowSummary, config.ShowSummary)
	// Merge ShowHoldings into ShowPositions with positions taking precedence
	// First check if Positions is set (CLI or config), then fall back to Holdings if not
	showPositionsFromCLI := options.ShowPositions
	showPositionsFromConfig := config.ShowPositions
	showHoldingsFromCLI := options.ShowHoldings
	showHoldingsFromConfig := config.ShowHoldings

	// Positions takes precedence: use it if set in CLI or config
	if showPositionsFromCLI || showPositionsFromConfig {
		config.ShowPositions = true
	} else {
		// Otherwise, fall back to Holdings
		config.ShowPositions = showHoldingsFromCLI || showHoldingsFromConfig
	}
	config.Sort = getStringOption(options.Sort, config.Sort)
	config.Cache = getCacheOption(options.NoCache, config.Cache)
	config.Debug = getBoolOption(options.Debug, config.Debug)

	return config, nil
}

// cacheEnabled reports whether the cache is enabled, defaulting to on when the
// config value is unset (nil).
func cacheEnabled(configValue *bool) bool {
	return configValue == nil || *configValue
}

// getCacheOption resolves whether the cache is enabled. The cache defaults to on
// and is disabled only by an explicit `cache: false` in config or the
// --no-cache flag, with the flag taking precedence.
func getCacheOption(noCacheFlag bool, configValue *bool) *bool {
	enabled := true

	if configValue != nil {
		enabled = *configValue
	}

	if noCacheFlag {
		enabled = false
	}

	return &enabled
}

func getConfigPath(fs afero.Fs, configPathOption string) (string, error) {
	if configPathOption != "" {
		return configPathOption, nil
	}

	home, _ := homedir.Dir()

	// Documented default is always ~/.config/ticker/config.yaml. On macOS,
	// adrg/xdg maps ConfigHome to ~/Library/Application Support, so that
	// platform path is searched as well. Legacy .ticker.yaml locations are
	// kept for backward compatibility.
	dotConfigTicker := home + "/.config/ticker"
	xdgTicker := xdg.ConfigHome + "/ticker"

	searches := []struct {
		name  string
		paths []string
	}{
		{name: "config", paths: uniquePaths(dotConfigTicker, xdgTicker)},
		{name: ".ticker", paths: uniquePaths(home, ".", home+"/.config", xdg.ConfigHome, dotConfigTicker, xdgTicker)},
	}

	var err error

	for _, search := range searches {
		v := viper.New()
		v.SetFs(fs)
		v.SetConfigType("yaml")
		for _, path := range search.paths {
			v.AddConfigPath(path)
		}
		v.SetConfigName(search.name)

		if err = v.ReadInConfig(); err == nil {
			return v.ConfigFileUsed(), nil
		}
	}

	return "", fmt.Errorf("invalid config: %w", err)
}

// uniquePaths returns paths with empty strings and duplicates removed,
// preserving order so search precedence stays stable.
func uniquePaths(paths ...string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))

	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	return out
}

func getRefreshInterval(optionsRefreshInterval int, configRefreshInterval int) int {

	if optionsRefreshInterval > 0 {
		return optionsRefreshInterval
	}

	if configRefreshInterval > 0 {
		return configRefreshInterval
	}

	return 5
}

func getBoolOption(cliValue bool, configValue bool) bool {

	if cliValue {
		return cliValue
	}

	if configValue {
		return configValue
	}

	return false
}

func getStringOption(cliValue string, configValue string) string {

	if cliValue != "" {
		return cliValue
	}

	if configValue != "" {
		return configValue
	}

	return ""
}

func getGroups(config c.Config, d c.Dependencies, cache c.Cache) ([]c.AssetGroup, error) {

	groups := make([]c.AssetGroup, 0)
	var configAssetGroups []c.ConfigAssetGroup

	tickerSymbolToSourceSymbol, err := symbol.GetTickerSymbols(d.SymbolsURL, cache)

	if err != nil {
		return []c.AssetGroup{}, err
	}

	if len(config.Watchlist) > 0 || len(config.Lots) > 0 {
		configAssetGroups = append(configAssetGroups, c.ConfigAssetGroup{
			Name:      "default",
			Watchlist: config.Watchlist,
			Lots:      config.Lots,
		})
	}

	configAssetGroups = append(configAssetGroups, config.AssetGroup...)

	for _, configAssetGroup := range configAssetGroups {
		groups = append(groups, buildAssetGroup(configAssetGroup, tickerSymbolToSourceSymbol))
	}

	return groups, nil

}

// buildAssetGroup resolves each symbol in a config group to its quote source,
// producing a runtime AssetGroup. Shared by config-derived groups and SnapTrade groups.
func buildAssetGroup(configAssetGroup c.ConfigAssetGroup, tickerSymbolToSourceSymbol symbol.TickerSymbolToSourceSymbol) c.AssetGroup {

	symbols := make(map[string]bool)
	symbolsUnique := make(map[c.QuoteSource]c.AssetGroupSymbolsBySource)
	var assetGroupSymbolsBySource []c.AssetGroupSymbolsBySource

	appendUnique := func(sym string) {
		if !symbols[sym] {
			symbols[sym] = true
			symbolsUnique = appendSymbol(symbolsUnique, getSymbolAndSource(sym, tickerSymbolToSourceSymbol))
		}
	}

	// Prefer Lots; fall back to the deprecated Holdings field for backwards compatibility
	lots := configAssetGroup.Lots
	mergedConfigAssetGroup := configAssetGroup
	if len(lots) == 0 {
		lots = configAssetGroup.Holdings
		mergedConfigAssetGroup.Lots = lots
	}

	for _, sym := range configAssetGroup.Watchlist {
		appendUnique(sym)
	}

	for _, lot := range lots {
		appendUnique(lot.Symbol)
	}

	for _, option := range configAssetGroup.Options {
		appendUnique(option.Symbol)
	}

	for _, symbolsBySource := range symbolsUnique {
		assetGroupSymbolsBySource = append(assetGroupSymbolsBySource, symbolsBySource)
	}

	return c.AssetGroup{
		ConfigAssetGroup: mergedConfigAssetGroup,
		SymbolsBySource:  assetGroupSymbolsBySource,
	}
}

func getLogger(d c.Dependencies) (*log.Logger, error) {
	// Create log file with current date
	currentTime := time.Now()
	logFileName := fmt.Sprintf("ticker-log-%s.log", currentTime.Format("2006-01-02"))
	logFile, err := d.Fs.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return log.New(logFile, "", log.LstdFlags), nil
}

func getSymbolAndSource(symbol string, tickerSymbolToSourceSymbol symbol.TickerSymbolToSourceSymbol) symbolSource {

	symbolUppercase := strings.ToUpper(symbol)

	if strings.HasSuffix(symbolUppercase, ".CB") {

		symbol = strings.ToUpper(symbol)[:len(symbol)-3]

		// Futures contracts on Coinbase Derivatives Exchange are implicitly USD-denominated
		if strings.HasSuffix(symbol, "-CDE") {
			return symbolSource{
				source: c.QuoteSourceCoinbase,
				symbol: symbol,
			}
		}

		return symbolSource{
			source: c.QuoteSourceCoinbase,
			symbol: symbol + "-USD",
		}
	}

	if strings.HasSuffix(symbolUppercase, ".X") {

		if tickerSymbolToSource, exists := tickerSymbolToSourceSymbol[symbolUppercase]; exists {

			return symbolSource{
				source: tickerSymbolToSource.Source,
				symbol: tickerSymbolToSource.SourceSymbol,
			}

		}

	}

	return symbolSource{
		source: c.QuoteSourceYahoo,
		symbol: symbolUppercase,
	}

}

func appendSymbol(symbolsUnique map[c.QuoteSource]c.AssetGroupSymbolsBySource, symbolAndSource symbolSource) map[c.QuoteSource]c.AssetGroupSymbolsBySource {

	if symbolsBySource, ok := symbolsUnique[symbolAndSource.source]; ok {

		symbolsBySource.Symbols = append(symbolsBySource.Symbols, symbolAndSource.symbol)

		symbolsUnique[symbolAndSource.source] = symbolsBySource

		return symbolsUnique
	}

	symbolsUnique[symbolAndSource.source] = c.AssetGroupSymbolsBySource{
		Source:  symbolAndSource.source,
		Symbols: []string{symbolAndSource.symbol},
	}

	return symbolsUnique

}
