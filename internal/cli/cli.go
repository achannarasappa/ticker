package cli

import (
	"errors"
	"fmt"
	"strings"
	"ticker/internal/position"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RefreshInterval       int            `yaml:"interval"`
	Watchlist             []string       `yaml:"watchlist"`
	Lots                  []position.Lot `yaml:"lots"`
	Separate              bool           `yaml:"show-separator"`
	ExtraInfoExchange     bool           `yaml:"show-tags"`
	ExtraInfoFundamentals bool           `yaml:"show-fundamentals"`
  Proxy                 string         `yaml:"proxy"`
}

type Options struct {
	RefreshInterval       *int
	Watchlist             *string
	Separate              *bool
	ExtraInfoExchange     *bool
	ExtraInfoFundamentals *bool
  Proxy                 *string
}

func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		err := uiStartFn()

		if err != nil {
			fmt.Println(fmt.Errorf("Unable to start UI: %w", err).Error())
		}
	}
}

func Validate(config *Config, fs afero.Fs, options Options, prevErr error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {

		if prevErr != nil {
			return prevErr
		}

		if len(config.Watchlist) == 0 && len(*options.Watchlist) == 0 {
			return errors.New("Invalid config: No watchlist provided")
		}

		if len(*options.Watchlist) != 0 {
			config.Watchlist = strings.Split(strings.ReplaceAll(*options.Watchlist, " ", ""), ",")
		}

		*config = mergeConfig(*config, options)

		return nil
	}
}

func ReadConfig(fs afero.Fs, configPathOption string) (Config, error) {
	var config Config
	configPath, err := getConfigPath(fs, configPathOption)

	if err != nil {
		return config, err
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

func mergeConfig(config Config, options Options) Config {
	config.RefreshInterval = getRefreshInterval(*options.RefreshInterval, config.RefreshInterval)
	config.Separate = getBoolOption(*options.Separate, config.Separate)
	config.ExtraInfoExchange = getBoolOption(*options.ExtraInfoExchange, config.ExtraInfoExchange)
	config.ExtraInfoFundamentals = getBoolOption(*options.ExtraInfoFundamentals, config.ExtraInfoFundamentals)
  config.Proxy = getProxy(*options.Proxy, config.Proxy)

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

func getProxy(optionsProxy string, configProxy string) string {

  if len(optionsProxy) > 0 {
    return optionsProxy
  }

  if len(configProxy) > 0 {
    return configProxy
  }

  return ""
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
