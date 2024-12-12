package cmd

import (
	"fmt"

	"github.com/jochumdev/lql-api/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  `Even I have a version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lql-api %s\n", version.Version)
	},
}
