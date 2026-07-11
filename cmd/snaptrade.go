package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/achannarasappa/ticker/v5/internal/cli"
)

//nolint:gochecknoglobals
var (
	snaptradeCmd = &cobra.Command{
		Use:   "snaptrade",
		Short: "Manage SnapTrade brokerage connections",
	}
	snaptradeConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect a brokerage account (e.g. Robinhood) via SnapTrade",
		Run: func(_ *cobra.Command, _ []string) {
			if connectErr := cli.ConnectSnapTrade(dep, config, cli.OpenBrowser); connectErr != nil {
				fmt.Println(connectErr)
				os.Exit(1)
			}
		},
	}
)
