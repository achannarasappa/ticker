/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	configPath      string
	config          cli.Config
	watchlist       string
	refreshInterval int
	rootCmd         = &cobra.Command{
		Use:   "ticker",
		Short: "Terminal stock ticker and stock gain/loss tracker",
		Args: cli.Validate(
			&config,
			afero.NewOsFs(),
			cli.Options{
				ConfigPath:      &configPath,
				RefreshInterval: &refreshInterval,
				Watchlist:       &watchlist,
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
	rootCmd.Flags().IntVarP(&refreshInterval, "interval", "i", 0, "Refresh interval in seconds")
}

func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		configPath = viper.ConfigFileUsed()
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
}
