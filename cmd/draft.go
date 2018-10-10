package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dbaltas/ergo/github"
	"github.com/dbaltas/ergo/repo"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var releaseTag string
var updateJiraFixVersions bool

func init() {
	rootCmd.AddCommand(draftCmd)
	draftCmd.Flags().StringVar(&releaseTag, "releaseTag", "", "Tag for the release. If empty, curent date in YYYY.MM.DD will be used")
	draftCmd.Flags().BoolVar(&updateJiraFixVersions, "update-jira-fix-versions", false, "Update fix versions on Jira based on the configuration string")

}

var draftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Create a draft release [github]",
	Long:  `Create a draft release on github comparing one target branch with the base branch`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return draftRelease()
	},
}

func draftRelease() error {
	yellow := color.New(color.FgYellow)
	branchMap := viper.GetStringMapString("release.branch-map")

	t := time.Now()

	tagName := releaseTag
	if tagName == "" {
		tagName = fmt.Sprintf("%4d.%02d.%02d", t.Year(), t.Month(), t.Day())
	}
	name := fmt.Sprintf("%s %d %d", t.Month(), t.Day(), t.Year())

	var diff []repo.DiffCommitBranch

	branches := strings.Split(releaseBranchesString, ",")
	for _, branch := range branches {
		ahead, behind, err := r.CompareBranch(baseBranch, branch)
		if err != nil {
			return fmt.Errorf("error comparing %s %s:%s", baseBranch, branch, err)
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
	if updateJiraFixVersions {
		// Eliminate duplicates by using a map
		uTasks := make(map[string]bool)
		taskRegExp := viper.GetString("jira.task-regex")

		re := regexp.MustCompile(fmt.Sprintf("(?m)(%v)", taskRegExp))
		for _, commit := range diff[0].Behind {
			res := re.FindAllStringSubmatch(commit.Message, -1)
			if len(res) > 0 {
				for _, task := range res {
					uTasks[task[0]] = true
				}
			}
		}

		// Get all the values
		tasks := make([]string, 0, len(uTasks))
		for task := range uTasks {
			tasks = append(tasks, task)
		}
		jc.UpdateIssueFixVersions(tasks)
	}

	fmt.Println(releaseBody)
	reader := bufio.NewReader(os.Stdin)
	yellow.Printf("Press 'ok' to continue with Drafting the release:")
	input, _ := reader.ReadString('\n')
	text := strings.Split(input, "\n")[0]
	if text != "ok" {
		fmt.Printf("No draft\n")
		return nil
	}

	release, err := gc.CreateDraftRelease(name, tagName, releaseBody)

	if err != nil {
		return err
	}
	fmt.Println(release)
	return nil
}
