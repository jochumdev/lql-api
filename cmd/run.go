package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().StringP("htpasswd", "p", "/opt/sites/$SITE/etc/htpasswd", "htpasswd file")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [site]",
	Short: "Run the Proxy",
	Long:  `Run the Check_MK LQL API Proxy`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Check_MK LQL API Proxy v0.1 -- HEAD")
	},
}
