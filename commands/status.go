package commands

import (
	"context"
	"strconv"

	"github.com/thebeatapp/ergo/github"

	"github.com/spf13/cobra"
	"github.com/thebeatapp/ergo"
	"github.com/thebeatapp/ergo/cli"
)

// defineStatusCommand defines the status command.
func defineStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "the status of branches compared to base branch",
		Long:  "Prints the commits ahead and behind of status branches compared to a base branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			githubClient := github.NewGithubClient(ctx, opts.AccToken)
			host := github.NewRepositoryClient(opts.Organization, opts.RepoName, githubClient)

			diff, err := host.DiffCommits(ctx, opts.Branches, opts.BaseBranch)
			if err != nil {
				return err
			}
			printBranchCompare(diff, host.GetRepoName())
			return nil
		},
	}
}

// printBranchCompare prints the status.
func printBranchCompare(commitDiffBranches []*ergo.StatusReport, repoName string) {
	prt := cli.NewCLI()
	prt.PrintColorizedLine("REPO: ", repoName, cli.WarningType)
	if len(commitDiffBranches) > 0 {
		prt.PrintColorizedLine("BASE: ", commitDiffBranches[0].BaseBranch, cli.WarningType)
	}
	headers := []string{"Branch", "Behind", "Ahead"}
	var body [][]string
	for _, diff := range commitDiffBranches {
		behind := strconv.Itoa(len(diff.Behind))
		ahead := strconv.Itoa(len(diff.Ahead))
		row := []string{diff.Branch, behind, ahead}
		body = append(body, row)
	}
	prt.PrintTable(headers, body)
}
