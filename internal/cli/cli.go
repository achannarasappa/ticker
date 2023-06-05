package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/achannarasappa/ticker/internal/cli/symbol"
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/quote"
	yahooClient "github.com/achannarasappa/ticker/internal/quote/yahoo/client"
	"github.com/achannarasappa/ticker/internal/ui/util"

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
	return func(cmd *cobra.Command, args []string) {
		err := uiStartFn()

		if err != nil {
			fmt.Println(fmt.Errorf("unable to start UI: %w", err).Error())
		}
	}
}

// Validate checks whether config is valid and returns an error if invalid or if an error was generated earlier
func Validate(ctx *c.Context, options *Options, prevErr *error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		if prevErr != nil && *prevErr != nil {
			return *prevErr
		}

		if len(ctx.Config.Watchlist) == 0 && len(options.Watchlist) == 0 && len(ctx.Config.Lots) == 0 && len(ctx.Config.AssetGroup) == 0 {
			return errors.New("invalid config: No watchlist provided") //nolint:goerr113
		}

		return nil
	}
}

func GetDependencies() c.Dependencies {

	client := yahooClient.New(resty.New(), resty.New())
	yahooClient.RefreshSession(client, resty.New())

	return c.Dependencies{
		Fs: afero.NewOsFs(),
		HttpClients: c.DependenciesHttpClients{
			Default: resty.New(),
			Yahoo:   client,
		},
	}

}

// GetContext builds the context from the config and reference data
func GetContext(d c.Dependencies, options Options, configPath string) (c.Context, error) {
	var (
		reference c.Reference
		config    c.Config
		groups    []c.AssetGroup
		err       error
	)

	config, err = readConfig(d.Fs, configPath)

	if err != nil {
		return c.Context{}, err
	}

	config = getConfig(config, options, &d.HttpClients)
	groups, err = getGroups(config, *d.HttpClients.Default)

	if err != nil {
		return c.Context{}, err
	}

	reference, err = getReference(config, groups, *d.HttpClients.Yahoo)

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

func getReference(config c.Config, assetGroups []c.AssetGroup, client resty.Client) (c.Reference, error) {

	currencyRates, err := quote.GetAssetGroupsCurrencyRates(client, assetGroups, config.Currency)
	styles := util.GetColorScheme(config.ColorScheme)

	return c.Reference{
		CurrencyRates: currencyRates,
		Styles:        styles,
	}, err

}

func getConfig(config c.Config, options Options, httpClients *c.DependenciesHttpClients) c.Config {

	if len(options.Watchlist) != 0 {
		config.Watchlist = strings.Split(strings.ReplaceAll(options.Watchlist, " ", ""), ",")
	}

	if len(config.Proxy) > 0 {
		httpClients.Default.SetProxy(config.Proxy)
		httpClients.Yahoo.SetProxy(config.Proxy)
	}

	config.RefreshInterval = getRefreshInterval(options.RefreshInterval, config.RefreshInterval)
	config.Separate = getBoolOption(options.Separate, config.Separate)
	config.ExtraInfoExchange = getBoolOption(options.ExtraInfoExchange, config.ExtraInfoExchange)
	config.ExtraInfoFundamentals = getBoolOption(options.ExtraInfoFundamentals, config.ExtraInfoFundamentals)
	config.ShowSummary = getBoolOption(options.ShowSummary, config.ShowSummary)
	config.ShowHoldings = getBoolOption(options.ShowHoldings, config.ShowHoldings)
	config.Proxy = getStringOption(options.Proxy, config.Proxy)
	config.Sort = getStringOption(options.Sort, config.Sort)

	return config
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
			symbol: strings.TrimSuffix(strings.ToLower(symbol), ".cg"),
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
