package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/webmeisterei/lql_api/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  `Even I have a version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lql_api %s\n", version.Version)
	},
}
