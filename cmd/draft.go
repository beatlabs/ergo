package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dbaltas/ergo/github"
	"github.com/dbaltas/ergo/repo"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var releaseTag string

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().StringVar(&releaseTag, "releaseTag", "", "Tag for the release. If empty, curent date in YYYY.MM.DD will be used")
}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Create a draft release on github comparing one target branch with the base branch",
	Long:  `Create a draft release on github comparing one target branch with the base branch`,
	Run: func(cmd *cobra.Command, args []string) {
		draftRelease()
	},
}

func draftRelease() {
	yellow := color.New(color.FgYellow)
	branchMap := viper.GetStringMapString("release.branch-map")
	releaseRepo := ""
	if strings.Contains(viper.GetString("generic.release-repos"), repoName) {
		releaseRepo = repoName
	}
	if releaseRepo == "" {
		fmt.Printf("Repo is not configured for release support\nAdd '%s' to config generic.release-repos\n", repoName)
		return
	}

	t := time.Now()

	tagName := releaseTag
	if tagName == "" {
		tagName = fmt.Sprintf("%4d.%02d.%02d", t.Year(), t.Month(), t.Day())
	}
	name := fmt.Sprintf("%s %d %d", t.Month(), t.Day(), t.Year())

	var diff []repo.DiffCommitBranch

	branches := strings.Split(releaseBranchesString, ",")
	for _, branch := range branches {
		ahead, behind, err := gitRepo.CompareBranch(baseBranch, branch)
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

	releaseBody := github.ReleaseBody(diff, viper.GetString("github.release-body-prefix"), branchMap)

	fmt.Println(releaseBody)
	reader := bufio.NewReader(os.Stdin)
	yellow.Printf("Press 'ok' to continue with Drafting the release:")
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No draft\n")
		return
	}

	release, err := gc.CreateDraftRelease(name, tagName, releaseBody)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(release)
}
