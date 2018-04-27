package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ergo",
	Long:  `Print the version of ergo`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.2.0")
	},
}
