package commands

import (
	"context"
	"fmt"

	"github.com/thebeatapp/ergo/github"
	"github.com/thebeatapp/ergo/release"

	"github.com/spf13/cobra"
	"github.com/thebeatapp/ergo/cli"
)

// defineTagCommand defines the tag command.
func defineTagCommand() *cobra.Command {
	var (
		suffix string
		minor  bool
		major  bool
	)

	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Create a tag on branch",
		Long:  "",
	}

	tagCmd.Flags().StringVar(&suffix, "suffix", "", "The suffix of the tag.")
	tagCmd.Flags().BoolVar(&minor, "minor", false, "The minor part of the tag.")
	tagCmd.Flags().BoolVar(&major, "major", false, "The major part of the tag.")

	tagCmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		var versionArg string
		if len(args) > 0 {
			versionArg = args[0]
		}

		var prt = cli.NewCLI()

		githubClient := github.NewGithubClient(ctx, opts.AccToken)
		host := github.NewRepositoryClient(opts.Organization, opts.RepoName, githubClient)

		ver, err := release.NewVersion(host, opts.BaseBranch).NextVersion(ctx, versionArg, suffix, major, minor)
		if err != nil {
			return err
		}

		tag := release.NewTag(host)

		tagExists, err := tag.ExistsTagName(ctx, ver.Name)
		if err != nil {
			return err
		}

		for tagExists {
			existsVersionMessage := fmt.Sprintf("Tag %q already exists. "+
				"Please provide a new tag name e.g. 'ergo tag 1.2.1'", ver.Name)
			prt.PrintColorizedLine("", existsVersionMessage, cli.ErrorType)

			ver.Name, err = prt.Input()
			if err != nil {
				return err
			}

			tagExists, err = tag.ExistsTagName(ctx, ver.Name)
			if err != nil {
				return err
			}
		}

		confirmationMessage := fmt.Sprintf("Create tag %q on %v", ver.Name, ver.SHA)
		confirm, err := prt.Confirmation(confirmationMessage, "Aborting...", "Creating tag...")
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}

		newTag, err := tag.Create(ctx, ver)
		if err != nil {
			return err
		}

		prt.PrintColorizedLine("", fmt.Sprintf("Successfully created tag: %q", newTag.Name), cli.SuccessType)

		return nil
	}

	return tagCmd
}
