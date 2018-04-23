package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/dbaltas/ergo/github"
	"github.com/dbaltas/ergo/repo"
	"github.com/spf13/viper"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

// CommitMessageFormat How to format the commit message
type CommitMessageFormat int8

const (
	// GithubRelease Format Commit message for use in github release
	GithubRelease CommitMessageFormat = 0
)

func main() {
	var cfgFile string
	var repoURL string
	var directory string
	var skipFetch bool
	var baseBranch string
	var branchesString string
	var createReleaseCmd bool
	var releaseInterval string
	var releaseOffset string

	//command flags
	var statusCmd bool
	var pendingCmd bool
	var deployCmd bool

	var err error
	var diff []repo.DiffCommitBranch

	flag.StringVar(&cfgFile, "cfgFile", "", "config file (default is $HOME/.ergo.yaml)")
	flag.StringVar(&repoURL, "repoUrl", "", "git repo Url. ssh and https supported")
	flag.StringVar(&directory, "directory", ".", "Location to store or retrieve from the repo")
	flag.BoolVar(&skipFetch, "skipFetch", false, "Skip fetch. When set you may not be up to date with remote")
	flag.StringVar(&baseBranch, "baseBranch", "", "Base branch for the comparison.")
	flag.StringVar(&branchesString, "branches", "", "Comma separated list of branches")
	flag.StringVar(&releaseInterval, "releaseInterval", "5m", "Duration to wait between releases. ('5m', '1h25m', '30s')")
	flag.StringVar(&releaseOffset, "releaseOffset", "10m", "Duration to wait before the first release ('5m', '1h25m', '30s')")

	// command flags. TODO: move to commands
	flag.BoolVar(&statusCmd, "status", false, "CMD: Display status of targetBranches compared to base branch")
	flag.BoolVar(&pendingCmd, "pending", false, "CMD: Display commits that branch is behind base branch")
	flag.BoolVar(&deployCmd, "deploy", false, "CMD: Deploy base branch to target branches")
	flag.BoolVar(&createReleaseCmd, "createRelease", false, "CMD: Create a draft release on github comparing one target branch with the base branch")

	flag.Parse()

	// viper.AddConfigPath(".")
	// err = viper.ReadInConfig()
	// if err != nil {
	// 	fmt.Printf("error reading config file: %v\n", err)
	// 	fmt.Println("Proceeding without configuration file")
	// 	return
	// }
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName(".ergo")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("error reading config file: %v\n", err)
		fmt.Println("Proceeding without configuration file")
	}

	if baseBranch == "" {
		baseBranch = viper.GetString("generic.base-branch")
	}
	if branchesString == "" {
		branchesString = viper.GetString("generic.status-branches")
	}

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

	if statusCmd {
		printBranchCompare(diff)
		return

	}

	if pendingCmd {
		printDetail(diff)
		return
	}

	if deployCmd {
		deployBranches(baseBranch, branches, releaseOffset, releaseInterval, directory)
		return
	}

	if createReleaseCmd {
		tagName := "2018.04.22"
		name := "April 22 2018"

		releaseBody := github.ReleaseBody(diff, viper.GetString("github.release-body-prefix"))

		release, err := github.CreateDraftRelease(
			context.Background(),
			viper.GetString("github.access-token"),
			"taxibeat", "rest",
			name, tagName, releaseBody)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(release)
	}
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

func deployBranches(baseBranch string, branches []string, releaseOffset string, releaseInterval string, directory string) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)

	fmt.Println()
	blue.Print("Release from: ")
	yellow.Println(baseBranch)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Branch", "Start Time")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	intervalDuration, err := time.ParseDuration(releaseInterval)
	if err != nil {
		fmt.Printf("error parsing interval %v", err)
		return
	}
	offsetDuration, err := time.ParseDuration(releaseOffset)
	if err != nil {
		fmt.Printf("error parsing offset %v", err)
		return
	}

	t := time.Now()
	t = t.Add(offsetDuration)
	firstRelease := t
	for _, branch := range branches {
		tbl.AddRow(branch, t.Format("15:04:05"))
		t = t.Add(intervalDuration)
	}

	tbl.Print()
	reader := bufio.NewReader(os.Stdin)
	yellow.Printf("Press 'ok' to continue with Deployment:")
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No deployment\n")
		return
	}
	fmt.Println(text)

	if firstRelease.Before(time.Now()) {
		yellow.Println("\ndeployment stopped since first released time has passed. Please run again")
		return
	}

	d := firstRelease.Sub(time.Now())
	green.Printf("Deployment will start in %s\n", d.String())
	time.Sleep(d)
	for i, branch := range branches {
		if i != 0 {
			time.Sleep(intervalDuration)
			t = t.Add(intervalDuration)
		}
		green.Printf("%s Deploying %s\n", time.Now().Format("15:04:05"), branch)
		cmd := fmt.Sprintf("cd %s && git push origin origin/%s:%s", directory, baseBranch, branch)
		green.Printf("%s Executing %s\n", time.Now().Format("15:04:05"), cmd)
		out, err := exec.Command("sh", "-c", cmd).Output()

		if err != nil {
			fmt.Printf("error executing command: %s %v\n", cmd, err)
			return
		}
		green.Printf("%s Triggered Successfully %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))
	}
}
