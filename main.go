package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var baseBranch string
	var branchesString string
	var err error
	var diff []DiffCommitBranch

	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "Skip fetch. When set you may not be up to date with remote")
	flag.StringVar(&baseBranch, "baseBranch", "master", "Base branch for the comparison.")
	flag.StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	flag.Parse()

	repo, err := loadOrClone(repoURL, directory, "origin", skipFetch)
	if err != nil {
		fmt.Printf("Error loading repo:%s\n", err)
		return
	}

	if branchesString == "" {
		fmt.Printf("no branches to compare, use -branches\n")
		return
	}

	branches := strings.Split(branchesString, ",")
	for _, branch := range branches {
		ahead, behind, err := compareBranch(repo, baseBranch, branch, directory)
		if err != nil {
			fmt.Printf("error comparing %s %s:%s\n", baseBranch, branch, err)
			return
		}
		branchCommitDiff := DiffCommitBranch{
			branch:     branch,
			baseBranch: baseBranch,
			ahead:      ahead,
			behind:     behind,
		}
		diff = append(diff, branchCommitDiff)
	}

	printBranchCompare(diff)
}

func printBranchCompare(commitDiffBranches []DiffCommitBranch) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	fmt.Println()
	blue.Print("BASE: ")
	yellow.Println(commitDiffBranches[0].baseBranch)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Branch", "Behind", "Ahead")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, diffBranch := range commitDiffBranches {
		tbl.AddRow(diffBranch.branch, len(diffBranch.behind), len(diffBranch.ahead))
	}

	tbl.Print()
}
