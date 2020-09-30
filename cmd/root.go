package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "lql-api",
		Short: "Check_MK LQL API Client/Server",
		Long:  ``,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()
}
