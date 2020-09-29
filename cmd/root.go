package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "lql_api_proxy",
		Short: "Check_MK LQL API Proxy",
		Long:  ``,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
}
