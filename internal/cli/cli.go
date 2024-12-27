package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/achannarasappa/ticker/v4/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/achannarasappa/ticker/v4/internal/quote"
	yahooClient "github.com/achannarasappa/ticker/v4/internal/quote/yahoo/client"
	"github.com/achannarasappa/ticker/v4/internal/ui/util"

	"github.com/adrg/xdg"
	"github.com/go-resty/resty/v2"
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
	ShowHoldings          bool
	Proxy                 string
	Sort                  string
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

// Validate checks whether config is valid and returns an error if invalid or if an error was generated earlier
func Validate(config *c.Config, options *Options, prevErr *error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {

		if prevErr != nil && *prevErr != nil {
			return *prevErr
		}

		if len(config.Watchlist) == 0 && len(options.Watchlist) == 0 && len(config.Lots) == 0 && len(config.AssetGroup) == 0 {
			return errors.New("invalid config: No watchlist provided") //nolint:goerr113
		}

		return nil
	}
}

func GetDependencies() c.Dependencies {

	return c.Dependencies{
		Fs: afero.NewOsFs(),
		HttpClients: c.DependenciesHttpClients{
			Default:      resty.New(),
			Yahoo:        yahooClient.New(resty.New(), resty.New()),
			YahooSession: resty.New(),
		},
	}

}

// GetContext builds the context from the config and reference data
func GetContext(d c.Dependencies, config c.Config) (c.Context, error) {
	var (
		reference c.Reference
		groups    []c.AssetGroup
		err       error
	)

	err = yahooClient.RefreshSession(d.HttpClients.Yahoo, d.HttpClients.YahooSession)

	if err != nil {
		return c.Context{}, err
	}

	groups, err = getGroups(config, *d.HttpClients.Default)

	if err != nil {
		return c.Context{}, err
	}

	reference, err = getReference(config, groups, d.HttpClients.Yahoo)

	if err != nil {
		return c.Context{}, err
	}

	context := c.Context{
		Reference: reference,
		Config:    config,
		Groups:    groups,
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

func getReference(config c.Config, assetGroups []c.AssetGroup, client *resty.Client) (c.Reference, error) {

	currencyRates, err := quote.GetAssetGroupsCurrencyRates(client, assetGroups, config.Currency)
	if err != nil {
		return c.Reference{}, err
	}

	styles := util.GetColorScheme(config.ColorScheme)
	sourceToUnderlyingAssetSymbols, err := quote.GetAssetGroupUnderlyingAssetSymbols(client, assetGroups)

	if err != nil {
		return c.Reference{}, err
	}

	return c.Reference{
		CurrencyRates:                  currencyRates,
		SourceToUnderlyingAssetSymbols: sourceToUnderlyingAssetSymbols,
		Styles:                         styles,
	}, err

}

func GetConfig(dep c.Dependencies, configPath string, options Options) (c.Config, error) {

	config, err := readConfig(dep.Fs, configPath)

	if err != nil {
		return c.Config{}, err
	}

	if len(options.Watchlist) != 0 {
		config.Watchlist = strings.Split(strings.ReplaceAll(options.Watchlist, " ", ""), ",")
	}

	if len(config.Proxy) > 0 {
		dep.HttpClients.Default.SetProxy(config.Proxy)
		dep.HttpClients.Yahoo.SetProxy(config.Proxy)
	}

	config.RefreshInterval = getRefreshInterval(options.RefreshInterval, config.RefreshInterval)
	config.Separate = getBoolOption(options.Separate, config.Separate)
	config.ExtraInfoExchange = getBoolOption(options.ExtraInfoExchange, config.ExtraInfoExchange)
	config.ExtraInfoFundamentals = getBoolOption(options.ExtraInfoFundamentals, config.ExtraInfoFundamentals)
	config.ShowSummary = getBoolOption(options.ShowSummary, config.ShowSummary)
	config.ShowHoldings = getBoolOption(options.ShowHoldings, config.ShowHoldings)
	config.Proxy = getStringOption(options.Proxy, config.Proxy)
	config.Sort = getStringOption(options.Sort, config.Sort)

	return config, nil
}

func getConfigPath(fs afero.Fs, configPathOption string) (string, error) {
	var err error
	if configPathOption != "" {
		return configPathOption, nil
	}

	home, _ := homedir.Dir()

	v := viper.New()
	v.SetFs(fs)
	v.SetConfigType("yaml")
	v.AddConfigPath(home)
	v.AddConfigPath(".")
	v.AddConfigPath(xdg.ConfigHome)
	v.AddConfigPath(xdg.ConfigHome + "/ticker")
	v.SetConfigName(".ticker")
	err = v.ReadInConfig()

	if err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	return v.ConfigFileUsed(), nil
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

func getGroups(config c.Config, client resty.Client) ([]c.AssetGroup, error) {

	groups := make([]c.AssetGroup, 0)
	var configAssetGroups []c.ConfigAssetGroup

	tickerSymbolToSourceSymbol, err := symbol.GetTickerSymbols(client)

	if err != nil {
		return []c.AssetGroup{}, err
	}

	if len(config.Watchlist) > 0 || len(config.Lots) > 0 {
		configAssetGroups = append(configAssetGroups, c.ConfigAssetGroup{
			Name:      "default",
			Watchlist: config.Watchlist,
			Holdings:  config.Lots,
		})
	}

	configAssetGroups = append(configAssetGroups, config.AssetGroup...)

	for _, configAssetGroup := range configAssetGroups {

		symbols := make(map[string]bool)
		symbolsUnique := make(map[c.QuoteSource]c.AssetGroupSymbolsBySource)
		var assetGroupSymbolsBySource []c.AssetGroupSymbolsBySource

		for _, symbol := range configAssetGroup.Watchlist {
			if !symbols[symbol] {
				symbols[symbol] = true
				symbolAndSource := getSymbolAndSource(symbol, tickerSymbolToSourceSymbol)
				symbolsUnique = appendSymbol(symbolsUnique, symbolAndSource)
			}
		}

		for _, lot := range configAssetGroup.Holdings {
			if !symbols[lot.Symbol] {
				symbols[lot.Symbol] = true
				symbolAndSource := getSymbolAndSource(lot.Symbol, tickerSymbolToSourceSymbol)
				symbolsUnique = appendSymbol(symbolsUnique, symbolAndSource)
			}
		}

		for _, symbolsBySource := range symbolsUnique {
			assetGroupSymbolsBySource = append(assetGroupSymbolsBySource, symbolsBySource)
		}

		groups = append(groups, c.AssetGroup{
			ConfigAssetGroup: configAssetGroup,
			SymbolsBySource:  assetGroupSymbolsBySource,
		})

	}

	return groups, nil

}

func getSymbolAndSource(symbol string, tickerSymbolToSourceSymbol symbol.TickerSymbolToSourceSymbol) symbolSource {

	symbolUppercase := strings.ToUpper(symbol)

	if strings.HasSuffix(symbolUppercase, ".CG") {
		return symbolSource{
			source: c.QuoteSourceCoingecko,
			symbol: strings.ToLower(symbol)[:len(symbol)-3],
		}
	}

	if strings.HasSuffix(symbolUppercase, ".CC") {
		return symbolSource{
			source: c.QuoteSourceCoinCap,
			symbol: strings.ToLower(symbol)[:len(symbol)-3],
		}
	}

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
