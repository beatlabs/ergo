package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/dbaltas/ergo/repo"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// CommitMessageFormat How to format the commit message
type CommitMessageFormat int8

const (
	// Format Commit message for use in github release
	GithubRelease CommitMessageFormat = 0
)

func main() {
	var repoURL string
	var directory string
	var skipFetch bool
	var baseBranch string
	var branchesString string
	var inDetail bool
	var err error
	var diff []repo.DiffCommitBranch

	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", "", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "Skip fetch. When set you may not be up to date with remote")
	flag.StringVar(&baseBranch, "baseBranch", "master", "Base branch for the comparison.")
	flag.StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	flag.BoolVar(&inDetail, "inDetail", false, "When true, display commits difference in detail")
	flag.Parse()

	repository, err := repo.LoadOrClone(repoURL, directory, "origin", skipFetch)
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
		ahead, behind, err := repo.CompareBranch(repository, baseBranch, branch, directory)
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

	if inDetail {
		printDetail(diff)
		return
	}
	printBranchCompare(diff)
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
	lineSeparator := "\\r\\n"

	for _, diffBranch := range commitDiffBranches {
		for _, commit := range diffBranch.Behind {
			fmt.Printf("%s%s", repo.FormatMessage(commit, firstLinePrefix, nextLinePrefix, lineSeparator), lineSeparator)
		}
	}
}
