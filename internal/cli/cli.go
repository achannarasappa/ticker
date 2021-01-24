package cli

import (
	"errors"
	"fmt"
	"strings"
	"ticker/internal/position"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RefreshInterval int            `yaml:"interval"`
	Watchlist       []string       `yaml:"watchlist"`
	Lots            []position.Lot `yaml:"lots"`
}

type Options struct {
	ConfigPath      *string
	RefreshInterval *int
	Watchlist       *string
}

func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		err := uiStartFn()

		if err != nil {
			fmt.Println(fmt.Errorf("Unable to start UI: %w", err).Error())
		}
	}
}

func Validate(config *Config, fs afero.Fs, options Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		*config, err = read(fs, options, config)
		if err != nil {
			return fmt.Errorf("Invalid config: %w", err)
		}
		return nil
	}
}

func read(fs afero.Fs, options Options, configFile *Config) (Config, error) {

	var (
		err    error
		config Config
	)
	if *options.ConfigPath != "" {

		handle, err := fs.Open(*options.ConfigPath)

		if err != nil {
			return config, err
		}

		defer handle.Close()
		err = yaml.NewDecoder(handle).Decode(&config)

		if err != nil {
			return config, err
		}
	}

	if len(config.Watchlist) == 0 && len(*options.Watchlist) == 0 {
		return config, errors.New("No watchlist provided")
	}

	if len(*options.Watchlist) != 0 {
		config.Watchlist = strings.Split(strings.ReplaceAll(*options.Watchlist, " ", ""), ",")
	}

	config.RefreshInterval = getRefreshInterval(*options.RefreshInterval, configFile.RefreshInterval)

	return config, err

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
