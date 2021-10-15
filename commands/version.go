package commands

import (
	"github.com/beatlabs/ergo/cli"
	"github.com/spf13/cobra"
)

// defineVersionCommand defines the version command.
func defineVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "the version of ergo",
		Long:  "the version of ergo",
		Run: func(cmd *cobra.Command, args []string) {
			cli.NewCLI().PrintColorizedLine("Version: ", version, cli.WarningType)
		},
	}
}
