package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"ticker/internal/cli"
	"ticker/internal/ui"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	configPath            string
	config                cli.Config
	watchlist             string
	refreshInterval       int
	separate              bool
	extraInfoExchange     bool
	extraInfoFundamentals bool
	sortQuotesBy          string
	rootCmd               = &cobra.Command{
		Use:   "ticker",
		Short: "Terminal stock ticker and stock gain/loss tracker",
		Args: cli.Validate(
			&config,
			afero.NewOsFs(),
			cli.Options{
				ConfigPath:            &configPath,
				RefreshInterval:       &refreshInterval,
				Watchlist:             &watchlist,
				Separate:              &separate,
				ExtraInfoExchange:     &extraInfoExchange,
				ExtraInfoFundamentals: &extraInfoFundamentals,
				SortQuotesBy:          &sortQuotesBy,
			},
		),
		Run: cli.Run(ui.Start(&config)),
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&configPath, "config", "", "config file (default is $HOME/.ticker.yaml)")
	rootCmd.Flags().StringVarP(&watchlist, "watchlist", "w", "", "comma separated list of symbols to watch")
	rootCmd.Flags().IntVarP(&refreshInterval, "interval", "i", 0, "refresh interval in seconds")
	rootCmd.Flags().BoolVar(&separate, "show-separator", false, "layout with separators between each quote")
	rootCmd.Flags().BoolVar(&extraInfoExchange, "show-tags", false, "display currency, exchange name, and quote delay for each quote")
	rootCmd.Flags().BoolVar(&extraInfoFundamentals, "show-fundamentals", false, "display open price, high, low, and volume for each quote")
	rootCmd.Flags().StringVar(&sortQuotesBy, "sort-quotes-by", "ChangePercent", "sort the quotes according to [Symbol, ExchangeName, ChangePercent]. default sort is by ChangePercent")
}

func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ticker")
	}

	viper.ReadInConfig()
	configPath = viper.ConfigFileUsed()
}
