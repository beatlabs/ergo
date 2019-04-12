package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/taxibeat/ergo/config"
	"github.com/taxibeat/ergo/config/viper"
)

var (
	vipOpts *viper.Options
	opts    *config.Options
)

// Execute entry point for commands.
func Execute() {
	rootCommand := setUpCommand()
	defineRootCommandProperties(rootCommand)
	rootCommand.AddCommand(defineStatusCommand())
	rootCommand.AddCommand(defineTagCommand())
	rootCommand.AddCommand(defineVersionCommand())
	rootCommand.AddCommand(defineDraftCommand())
	rootCommand.AddCommand(defineDeployCommand())
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initConfig initializes the config and sets the global opts variable which is shared by all commands.
func initConfig() {
	var err error
	vipOpts = viper.NewOptions()

	opts, err = vipOpts.InitConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// defineRootCommandProperties defines the root command properties. This is the root command which is
// running in every "sub" command.
func defineRootCommandProperties(rootCmd *cobra.Command) {
	var (
		path           string
		baseBranch     string
		branchesString string
		repoName       string
		owner          string
	)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&path, "path", ".", "Location to store or retrieve from the repo")
	rootCmd.PersistentFlags().StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	rootCmd.PersistentFlags().StringVar(&baseBranch, "base", "", "Base branch for the comparison.")
	rootCmd.PersistentFlags().StringVar(&owner, "owner", "", "")
	rootCmd.PersistentFlags().StringVar(&repoName, "repo", "", "")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return persistentPreRunECommand(cmd.Name(), baseBranch, branchesString, repoName, owner)
	}

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Hola! type `ergo help`")
	}
}

// persistentPreRunECommand defines the root command actions before the run command.
func persistentPreRunECommand(cmdName, baseBranch, branchesString, repoName, owner string) error {
	var err error
	// commands not requiring a repo
	noRepoCmds := make(map[string]bool)
	noRepoCmds["help"] = true
	noRepoCmds["version"] = true

	if _, ok := noRepoCmds[cmdName]; ok {
		return nil
	}

	vipOpts.BaseBranch = baseBranch
	vipOpts.SetBranchesString(branchesString)
	vipOpts.RepoName = repoName
	vipOpts.Organization = owner
	vipOpts.RefreshConfig()
	opts, err = vipOpts.GetConfig()
	if err != nil {
		return fmt.Errorf("error Initializing ergo options %v", err)
	}

	return nil
}

// setUpCommand initializes and returns a new command object.
func setUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ergo",
		Short: "ergo is a tool that aims to help the daily developer workflow",
		Long: "Ergo helps to" +
			"\n * compare multiple branches" +
			"\n * push to multiple branches with time interval (useful for multiple release environments)" +
			"\n * minimize the browser interaction with github:" +
			"\n\t * draft a release" +
			"\n\t * update release notes",
	}
}
