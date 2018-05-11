package cmd

import (
	"fmt"

	"github.com/dbaltas/ergo/repo"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var detail bool

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.Flags().BoolVar(&detail, "detail", false, "Print commits in detail")
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of branches compared to baseBranch",
	Long:  `Prints the commits ahead and behind of status branches compared to a base branch`,
	Run: func(cmd *cobra.Command, args []string) {
		var diff []repo.DiffCommitBranch

		for _, branch := range branches {
			ahead, behind, err := r.CompareBranch(baseBranch, branch)
			if err != nil {
				fmt.Printf("error comparing %s %s:%s\n", baseBranch, branch, err)
				return
			}
			branchCommitDiff := repo.DiffCommitBranch{
				Branch:     branch,
				BaseBranch: baseBranch,
				Ahead:      ahead,
				Behind:     behind,
			}
			diff = append(diff, branchCommitDiff)
		}

		if detail {
			printDetail(diff)
			return
		}
		printBranchCompare(diff)
	},
}

func printBranchCompare(commitDiffBranches []repo.DiffCommitBranch) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	fmt.Println()
	blue.Print("BASE: ")
	yellow.Println(commitDiffBranches[0].BaseBranch)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Branch", "Behind", "Ahead")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, diffBranch := range commitDiffBranches {
		tbl.AddRow(diffBranch.Branch, len(diffBranch.Behind), len(diffBranch.Ahead))
	}

	tbl.Print()
}

func printDetail(commitDiffBranches []repo.DiffCommitBranch) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	fmt.Println()
	blue.Print("BASE: ")
	yellow.Println(commitDiffBranches[0].BaseBranch)

	firstLinePrefix := "- [ ] "
	nextLinePrefix := "     "
	lineSeparator := "\r\n"

	for _, diffBranch := range commitDiffBranches {
		for _, commit := range diffBranch.Behind {
			fmt.Printf("%s%s", repo.FormatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator), lineSeparator)
		}
	}
}
