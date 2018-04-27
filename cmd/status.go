package cmd

import (
	"fmt"
	"strings"

	"github.com/dbaltas/ergo/repo"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of branches compared to baseBranch",
	Long:  `Prints the commits ahead and behind of status branches compared to a base branch`,
	Run: func(cmd *cobra.Command, args []string) {
		var diff []repo.DiffCommitBranch
		r, _ := getRepo()

		branches := strings.Split(branchesString, ",")
		for _, branch := range branches {
			ahead, behind, err := repo.CompareBranch(r, baseBranch, branch, directory)
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

		fmt.Println(viper.GetString("generic.base-branch"))
		fmt.Println(r)
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
