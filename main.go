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
	var releaseInterval string
	var releaseOffset string
	var releaseTag string

	//command flags
	var statusCmd bool
	var pendingCmd bool
	var deployCmd bool
	var createReleaseCmd bool
	var lastReleaseCmd bool

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
	flag.StringVar(&releaseTag, "releaseTag", "", "Tag for the release. If empty, curent date in YYYY.MM.DD will be used")

	// command flags. TODO: move to commands
	flag.BoolVar(&statusCmd, "status", false, "CMD: Display status of targetBranches compared to base branch")
	flag.BoolVar(&lastReleaseCmd, "lastRelease", false, "CMD: Display last Release of repo")
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

	repository, err := repo.LoadOrClone(repoURL, directory, viper.GetString("generic.remote"), skipFetch)
	if err != nil {
		fmt.Printf("Error loading repo:%s\n", err)
		return
	}

	if branchesString == "" {
		fmt.Printf("no branches to compare, use -branches\n")
		return
	}

	rmt, err := repository.Remote(viper.GetString("generic.remote"))

	if err != nil {
		fmt.Println(err)
		return
	}

	parts := strings.Split(rmt.Config().URLs[0], "/")
	repoName := strings.TrimSuffix(parts[len(parts)-1], ".git")

	repoForRelease := ""
	if strings.Contains(viper.GetString("generic.release-repos"), repoName) {
		repoForRelease = repoName
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
		branchMap := viper.GetStringMapString("release.branch-map")
		deployBranches(
			viper.GetString("generic.organization"),
			viper.GetString("generic.remote"),
			repoForRelease,
			baseBranch,
			branches,
			releaseOffset,
			releaseInterval,
			directory,
			branchMap,
			viper.GetString("release.on-deploy.body-branch-suffix-find"),
			viper.GetString("release.on-deploy.body-branch-suffix-replace"),
			viper.GetString("github.access-token"),
		)
		return
	}

	if createReleaseCmd {
		if repoForRelease == "" {
			fmt.Printf("Repo is not configured for release support\nAdd '%s' to config generic.release-repos\n", repoName)
			return
		}

		t := time.Now()

		tagName := releaseTag
		if tagName == "" {
			tagName = fmt.Sprintf("%4d.%02d.%02d", t.Year(), t.Month(), t.Day())
		}
		name := fmt.Sprintf("%s %d %d", t.Month(), t.Day(), t.Year())

		releaseBody := github.ReleaseBody(diff, viper.GetString("github.release-body-prefix"))

		release, err := github.CreateDraftRelease(
			context.Background(),
			viper.GetString("github.access-token"),
			viper.GetString("generic.organization"),
			repoForRelease,
			name, tagName, releaseBody)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(release)
		return
	}

	if lastReleaseCmd {
		release, err := github.LastRelease(
			context.Background(),
			viper.GetString("github.access-token"),
			viper.GetString("generic.organization"),
			viper.GetString("generic.release-repo"),
		)
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

func deployBranches(organization string, remote string, releaseRepo string, baseBranch string, branches []string, releaseOffset string, releaseInterval string, directory string, branchMap map[string]string, suffixFind string, suffixReplace string, githubAccessToken string) {
	blue := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)

	reference := baseBranch

	integrateGithubRelease := releaseRepo != ""

	if integrateGithubRelease {
		release, err := github.LastRelease(
			context.Background(),
			githubAccessToken,
			organization,
			releaseRepo,
		)
		if err != nil {
			fmt.Println(err)
		}
		reference = release.GetTagName()
		green.Printf("Deploying %s\n", release.GetHTMLURL())
	}

	blue.Printf("Release reference: %s\n", reference)
	green.Println("Deployment start times are estimates.")

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
		tbl.AddRow(branch, t.Format("15:04"))
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
		cmd := ""
		// if reference is a branch name, use origin
		if reference == baseBranch {
			cmd = fmt.Sprintf("cd %s && git push %s %s/%s:%s", directory, remote, remote, reference, branch)
		} else { // if reference is a tag don't prefix with origin
			cmd = fmt.Sprintf("cd %s && git push %s %s:%s", directory, remote, reference, branch)
		}
		green.Printf("%s Executing %s\n", time.Now().Format("15:04:05"), cmd)
		out, err := exec.Command("sh", "-c", cmd).Output()

		if err != nil {
			fmt.Printf("error executing command: %s %v\n", cmd, err)
			return
		}
		green.Printf("%s Triggered Successfully %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))

		branchText, ok := branchMap[branch]
		if integrateGithubRelease && ok && suffixFind != "" {
			t := time.Now()
			green.Printf("%s Updating release on github %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))
			release, err := github.LastRelease(
				context.Background(),
				githubAccessToken,
				organization,
				releaseRepo,
			)
			if err != nil {
				fmt.Println(err)
				return
			}
			findText := fmt.Sprintf("%s%s", branchText, suffixFind)
			replaceText := fmt.Sprintf("%s-%d_%s_%d_%02d:%02d%s", branchText, t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), suffixReplace)
			newBody := strings.Replace(*(release.Body), findText, replaceText, -1)
			fmt.Println(newBody)
			release.Body = &newBody
			release, err = github.EditRelease(
				context.Background(),
				githubAccessToken,
				organization,
				releaseRepo,
				release,
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			green.Printf("%s Updated release on github %s\n", time.Now().Format("15:04:05"), strings.TrimSpace(string(out)))
		}
	}
}
