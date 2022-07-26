package commands

import (
	"context"

	"github.com/beatlabs/ergo/cli"
	"github.com/beatlabs/ergo/github"
	"github.com/beatlabs/ergo/release"
	"github.com/spf13/cobra"
)

// defineDeployCommand defines the deploy command.
func defineDeployCommand() *cobra.Command {
	var (
		releaseOffset   string
		releaseInterval string
		allowForcePush  bool
		branchesString  string
		skipConfirm     bool
		publishDraft    bool
	)

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy base branch to target branches",
		Long:  "Deploy base branch to target branches",
	}

	deployCmd.Flags().StringVar(&releaseOffset, "releaseOffset", "1m", "Duration to wait before the first release ('5m', '1h25m', '30s')")
	deployCmd.Flags().StringVar(&releaseInterval, "releaseInterval", "25m", "Duration to wait between releases. ('5m', '1h25m', '30s')\n"+
		"You can do a non-linear interval by supplying more values: ('15m,10m,5m,5m,5m')")
	deployCmd.Flags().BoolVar(&allowForcePush, "force", false, "Allow force push if deploy branch has diverged from base")
	deployCmd.Flags().StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	deployCmd.Flags().BoolVar(&skipConfirm, "skip-confirmation", false, "Create the draft without asking for user confirmation.")
	deployCmd.Flags().BoolVar(&publishDraft, "publish-draft", false, "Publish the latest draft release before deployment.")

	deployCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return defineDeployCommandRun(releaseInterval, releaseOffset, branchesString, allowForcePush, skipConfirm, publishDraft)
	}

	return deployCmd
}

// defineDeployCommandRun defines the deploy command run actions.
func defineDeployCommandRun(releaseInterval, releaseOffset, branchesString string, allowForcePush, skipConfirm, publishDraft bool) error {
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
	).Do(ctx, releaseInterval, releaseOffset, allowForcePush, skipConfirm, publishDraft)
}
