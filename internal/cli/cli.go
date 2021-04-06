package cli

import (
	"errors"
	"fmt"
	"strings"

	. "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
	"github.com/achannarasappa/ticker/internal/position"
	"github.com/achannarasappa/ticker/internal/ui/util"

	"github.com/adrg/xdg"
	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

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

func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		err := uiStartFn()

		if err != nil {
			fmt.Println(fmt.Errorf("Unable to start UI: %w", err).Error())
		}
	}
}

func Validate(ctx *Context, options *Options, prevErr *error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		if prevErr != nil {
			return *prevErr
		}

		if len(ctx.Config.Watchlist) == 0 && len(options.Watchlist) == 0 && len(ctx.Config.Lots) == 0 {
			return errors.New("Invalid config: No watchlist provided")
		}

		return nil
	}
}

func GetContext(d Dependencies, options Options, configPath string) (Context, error) {
	var (
		reference Reference
		config    Config
		err       error
	)

	config, err = readConfig(d.Fs, configPath)

	if err != nil {
		return Context{}, err
	}

	config = getConfig(config, options, *d.HttpClient)
	reference, err = getReference(config, *d.HttpClient)

	if err != nil {
		return Context{}, err
	}

	context := Context{
		Reference: reference,
		Config:    config,
	}

	return context, nil
}

func readConfig(fs afero.Fs, configPathOption string) (Config, error) {
	var config Config
	configPath, err := getConfigPath(fs, configPathOption)

	if err != nil {
		return config, nil
	}
	handle, err := fs.Open(configPath)

	if err != nil {
		return config, fmt.Errorf("Invalid config: %w", err)
	}

	defer handle.Close()
	err = yaml.NewDecoder(handle).Decode(&config)

	if err != nil {
		return config, fmt.Errorf("Invalid config: %w", err)
	}

	return config, nil
}

func getReference(config Config, client resty.Client) (Reference, error) {

	aggregatedLots := position.GetLots(config.Lots)
	symbols := position.GetSymbols(config, aggregatedLots)

	currencyRates, err := currency.GetCurrencyRates(client, symbols, config.Currency)
	styles := util.GetColorScheme(config.ColorScheme)

	return Reference{
		CurrencyRates: currencyRates,
		Styles:        styles,
	}, err

}

func getConfig(config Config, options Options, client resty.Client) Config {

	if len(options.Watchlist) != 0 {
		config.Watchlist = strings.Split(strings.ReplaceAll(options.Watchlist, " ", ""), ",")
	}

	if len(config.Proxy) > 0 {
		client.SetProxy(config.Proxy)
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
		return "", fmt.Errorf("Invalid config: %w", err)
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
