package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"ticker/internal/cli"
	"ticker/internal/ui"
)

var (
	configPath            string
	config                cli.Config
	watchlist             string
	refreshInterval       int
	separate              bool
	extraInfoExchange     bool
	extraInfoFundamentals bool
	proxy                 string
	sort                  string
	err                   error
	rootCmd               = &cobra.Command{
		Use:   "ticker",
		Short: "Terminal stock ticker and stock gain/loss tracker",
		Args: cli.Validate(
			&config,
			afero.NewOsFs(),
			cli.Options{
				RefreshInterval:       &refreshInterval,
				Watchlist:             &watchlist,
				Separate:              &separate,
				ExtraInfoExchange:     &extraInfoExchange,
				ExtraInfoFundamentals: &extraInfoFundamentals,
				Proxy:                 &proxy,
				Sort:                  &sort,
			},
			err,
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
	rootCmd.Flags().StringVar(&proxy, "proxy", "", "proxy URL for requests (default is none)")
	rootCmd.Flags().StringVar(&sort, "sort", "", "sort the quotes according to [Symbol, ChangePercent]. default sort is by ChangePercent")
}

func initConfig() {
	config, err = cli.ReadConfig(afero.NewOsFs(), configPath)
}
