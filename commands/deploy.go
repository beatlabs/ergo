package commands

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/thebeatapp/ergo/cli"
	"github.com/thebeatapp/ergo/github"
	"github.com/thebeatapp/ergo/release"
)

// defineDeployCommand defines the deploy command.
func defineDeployCommand() *cobra.Command {
	var (
		releaseOffset   string
		releaseInterval string
		allowForcePush  bool
		branchesString  string
	)

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy base branch to target branches",
		Long:  "Deploy base branch to target branches",
	}

	deployCmd.Flags().StringVar(&releaseOffset, "releaseOffset", "1m", "Duration to wait before the first release ('5m', '1h25m', '30s')")
	deployCmd.Flags().StringVar(&releaseInterval, "releaseInterval", "25m", "Duration to wait between releases. ('5m', '1h25m', '30s')")
	deployCmd.Flags().BoolVar(&allowForcePush, "force", false, "Allow force push if deploy branch has diverged from base")
	deployCmd.Flags().StringVar(&branchesString, "branches", "", "Comma separated list of branches")

	deployCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return defineDeployCommandRun(releaseInterval, releaseOffset, branchesString, allowForcePush)
	}

	return deployCmd
}

// defineDeployCommandRun defines the deploy command run actions.
func defineDeployCommandRun(releaseInterval, releaseOffset, branchesString string, allowForcePush bool) error {
	ctx := context.Background()

	if branchesString != "" {
		vipOpts.SetReleaseBranches(branchesString)
	}

	printer := cli.NewCLI()

	githubClient := github.NewGithubClient(ctx, opts.AccToken)
	host := github.NewRepositoryClient(opts.Organization, opts.RepoName, githubClient)

	return release.NewDeploy(
		printer,
		host,
		opts.BaseBranch,
		opts.ReleaseBodyFind,
		opts.ReleaseBodyReplace,
		opts.ReleaseBranches,
		opts.ReleaseBodyBranches,
	).Do(ctx, releaseInterval, releaseOffset, allowForcePush)
}
