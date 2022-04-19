package commands

import (
	"context"

	"github.com/beatlabs/ergo/release"

	"github.com/beatlabs/ergo/cli"

	"github.com/beatlabs/ergo/github"
	"github.com/spf13/cobra"
)

// defineDraftCommand defines the draft command.
func defineDraftCommand() *cobra.Command {
	var (
		releaseName      string
		releaseTag       string
		branchesString   string
		minor            bool
		major            bool
		suffix           string
		skipConfirmation bool
	)

	draftCmd := &cobra.Command{
		Use:   "draft",
		Short: "Create a draft release [github]",
		Long:  "Create a draft release on github comparing one target branch with the base branch",
	}

	draftCmd.Flags().StringVar(&releaseName, "releaseName", "", "Name for the release. If empty the tag name will be used")
	draftCmd.Flags().StringVar(&releaseTag, "releaseTag", "", "Tag for the release. If empty, current date in YYYY.MM.DD will be used")
	draftCmd.Flags().BoolVar(&minor, "minor", false, "The minor part of the tag.")
	draftCmd.Flags().BoolVar(&major, "major", false, "The major part of the tag.")
	draftCmd.Flags().StringVar(&suffix, "suffix", "", "The suffix of the tag.")
	draftCmd.Flags().StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	draftCmd.Flags().BoolVar(&skipConfirmation, "skip-confirmation", false, "Create the draft without asking for user confirmation.")

	draftCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return defineDraftCommandRun(releaseName, releaseTag, suffix, branchesString, major, minor, skipConfirmation)
	}

	return draftCmd
}

// defineDraftCommandRun defines the draft command run actions.
func defineDraftCommandRun(releaseName, releaseTag, suffix, branchesString string, major, minor, skipConfirmation bool) error {
	ctx := context.Background()

	if branchesString != "" {
		vipOpts.SetReleaseBranches(branchesString)
	}

	printer := cli.NewCLI()

	githubClient := github.NewGithubClient(ctx, opts.AccToken)
	host := github.NewRepositoryClient(opts.Organization, opts.RepoName, githubClient)

	version, err := release.NewVersion(host, opts.BaseBranch).NextVersion(ctx, releaseTag, suffix, major, minor)
	if err != nil {
		return err
	}

	if releaseName == "" {
		releaseName = version.Name
	}

	return release.NewDraft(
		printer,
		host,
		opts.BaseBranch,
		opts.ReleaseBodyPrefix,
		opts.ReleaseBranches,
		opts.ReleaseBodyBranches,
	).Create(ctx, releaseName, version.Name, skipConfirmation)
}
