package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/achannarasappa/ticker/internal/cli"
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/print"
	"github.com/achannarasappa/ticker/internal/ui"
)

//nolint:gochecknoglobals
var (
	// Version is a placeholder that is replaced at build time with a linker flag (-ldflags)
	Version      = "v0.0.0"
	configPath   string
	dep          c.Dependencies
	ctx          c.Context
	options      cli.Options
	optionsPrint print.Options
	err          error
	rootCmd      = &cobra.Command{
		Version: Version,
		Use:     "ticker",
		Short:   "Terminal stock ticker and stock gain/loss tracker",
		Args:    cli.Validate(&ctx, &options, &err),
		Run:     cli.Run(ui.Start(&dep, &ctx)),
	}
	printCmd = &cobra.Command{
		Use:   "print",
		Short: "Prints holdings",
		Args:  cli.Validate(&ctx, &options, &err),
		Run:   print.Run(&dep, &ctx, &optionsPrint),
	}
)

// Execute starts the CLI or prints an error is there is one
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() { //nolint: gochecknoinits
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&configPath, "config", "", "config file (default is $HOME/.ticker.yaml)")
	rootCmd.Flags().StringVarP(&options.Watchlist, "watchlist", "w", "", "comma separated list of symbols to watch")
	rootCmd.Flags().IntVarP(&options.RefreshInterval, "interval", "i", 0, "refresh interval in seconds")
	rootCmd.Flags().BoolVar(&options.Separate, "show-separator", false, "layout with separators between each quote")
	rootCmd.Flags().BoolVar(&options.ExtraInfoExchange, "show-tags", false, "display currency, exchange name, and quote delay for each quote")
	rootCmd.Flags().BoolVar(&options.ExtraInfoFundamentals, "show-fundamentals", false, "display open price, high, low, and volume for each quote")
	rootCmd.Flags().BoolVar(&options.ShowSummary, "show-summary", false, "display summary of total gain and loss for positions")
	rootCmd.Flags().BoolVar(&options.ShowHoldings, "show-holdings", false, "display average unit cost, quantity, portfolio weight")
	rootCmd.Flags().StringVar(&options.Proxy, "proxy", "", "proxy URL for requests (default is none)")
	rootCmd.Flags().StringVar(&options.Sort, "sort", "", "sort quotes on the UI. Set \"alpha\" to sort by ticker name. Set \"value\" to sort by position value. Keep empty to sort according to change percent")
	rootCmd.AddCommand(printCmd)
	printCmd.Flags().StringVar(&optionsPrint.Format, "format", "", "output format for printing holdings. Set \"csv\" to print as a CSV or \"json\" for JSON. Defaults to JSON.")
	printCmd.Flags().StringVar(&configPath, "config", "", "config file (default is $HOME/.ticker.yaml)")
}

func initConfig() {
	dep = cli.GetDependencies()
	ctx, err = cli.GetContext(dep, options, configPath)

}
